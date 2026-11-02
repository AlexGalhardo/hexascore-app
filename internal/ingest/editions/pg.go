package editions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/entities"
)

// PG is the Store over the editions table.
type PG struct {
	Ex database.Executor
}

// WorkOfISBN finds the work that holds isbn13.
func (p PG) WorkOfISBN(ctx context.Context, isbn13 string) (string, bool, error) {
	var work string

	err := p.Ex.QueryRowContext(ctx, `SELECT work_id::text FROM editions WHERE isbn13 = $1`, isbn13).Scan(&work)
	if err == sql.ErrNoRows {
		return "", false, nil
	}

	return work, err == nil, err
}

// WorkOfQID finds the work that holds the edition item qid.
func (p PG) WorkOfQID(ctx context.Context, qid string) (string, bool, error) {
	var work string

	err := p.Ex.QueryRowContext(ctx, `SELECT work_id::text FROM editions WHERE wikidata_qid = $1`, qid).Scan(&work)
	if err == sql.ErrNoRows {
		return "", false, nil
	}

	return work, err == nil, err
}

// freeOLID is $4 as an Open Library id, or null when another edition already holds it: Save guards only ISBNs and
// QIDs, and the unique openlibrary_id must not fail the whole save.
const freeOLID = `(SELECT NULLIF($4, '') WHERE NOT EXISTS (SELECT 1 FROM editions WHERE openlibrary_id = $4))`

// Upsert updates the work's edition that carries one of the identifiers, or inserts a new one.
func (p PG) Upsert(ctx context.Context, work string, c Candidate) error {
	title, lang := c.Title, c.Language

	if title == "" {
		title = c.ISBN + c.QID + c.OpenLibraryID
	}

	if lang == "" {
		lang = "und"
	}

	res, err := p.Ex.ExecContext(ctx, `UPDATE editions SET title = $5, language = $6, market_country = NULLIF($7, ''),
publication_year = NULLIF($8, 0), publisher_id = coalesce(NULLIF($9, '')::uuid, publisher_id),
openlibrary_id = coalesce(openlibrary_id, `+freeOLID+`), updated_at = now()
WHERE work_id = $1 AND (isbn13 = NULLIF($2, '') OR wikidata_qid = NULLIF($3, '') OR openlibrary_id = NULLIF($4, ''))`,
		work, c.ISBN, c.QID, c.OpenLibraryID, title, lang, c.MarketCountry, c.Year, c.PublisherID)
	if err != nil {
		return fmt.Errorf("editions: update: %w", err)
	}

	if n, _ := res.RowsAffected(); n > 0 {
		return nil
	}

	if c.ISBN == "" && c.QID == "" {
		// Open Library-only: its id is the only identifier, so when another work holds it the insert would fail the
		// identifier CHECK and roll back the whole backfill chunk. Skip; the other work keeps the edition.
		var taken bool
		if err := p.Ex.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM editions WHERE openlibrary_id = $1)`,
			c.OpenLibraryID).Scan(&taken); err != nil {
			return fmt.Errorf("editions: openlibrary id lookup: %w", err)
		}

		if taken {
			return nil
		}
	}

	_, err = p.Ex.ExecContext(
		ctx,
		`INSERT INTO editions (id, work_id, isbn13, wikidata_qid, openlibrary_id, title, language,
market_country, publication_year, source, publisher_id)
VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), `+strings.ReplaceAll(freeOLID, "$4", "$5")+`, $6, $7, NULLIF($8, ''), NULLIF($9, 0), $10,
NULLIF($11, '')::uuid)`,
		uuid.NewV7().String(),
		work,
		c.ISBN,
		c.QID,
		c.OpenLibraryID,
		title,
		lang,
		c.MarketCountry,
		c.Year,
		c.Source,
		c.PublisherID,
	)
	if err != nil {
		return fmt.Errorf("editions: insert: %w", err)
	}

	return nil
}

// SetPrimary points the work at its primary stored edition (PrimaryEdition over every edition the work holds, so
// community and legacy editions count too). original is the work's language; "" or no match leaves the work as is.
func (p PG) SetPrimary(ctx context.Context, work, original string) error {
	if original == "" {
		return nil
	}

	rows, err := p.Ex.QueryContext(ctx, `SELECT id::text, language, coalesce(publication_year, 0), cover_key IS NOT NULL
FROM editions WHERE work_id = $1 ORDER BY id`, work)
	if err != nil {
		return fmt.Errorf("editions: list for primary: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var (
		ids   []string
		cands []Candidate
	)

	for rows.Next() {
		var (
			id string
			c  Candidate
		)
		if err := rows.Scan(&id, &c.Language, &c.Year, &c.HasCover); err != nil {
			return fmt.Errorf("editions: list for primary: %w", err)
		}

		// Legacy and community editions may store a region tag ("pt-BR"); PrimaryEdition compares base languages.
		c.Language = strings.ToLower(strings.SplitN(c.Language, "-", 2)[0])

		ids, cands = append(ids, id), append(cands, c)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("editions: list for primary: %w", err)
	}

	i := PrimaryEdition(cands, original)
	if i < 0 {
		return nil
	}

	if _, err := p.Ex.ExecContext(ctx, `UPDATE items SET primary_edition_id = $2
WHERE id = $1 AND primary_edition_id IS DISTINCT FROM $2::uuid`, work, ids[i]); err != nil {
		return fmt.Errorf("editions: set primary: %w", err)
	}

	return nil
}

// publisherStore resolves an edition publisher with entities.Resolve without writing a work credit: Link only keeps
// the entity id, and Create makes no native entity from a bare QID (the edition entity carries no publisher label).
type publisherStore struct {
	entities.PG

	id string
}

func (s *publisherStore) Link(_ context.Context, id string, _ entities.Credit) error {
	s.id = id

	return nil
}

func (s *publisherStore) Create(ctx context.Context, c entities.Credit) (string, error) {
	if c.Name == "" {
		return "", nil
	}

	return s.PG.Create(ctx, c)
}

// resolvePublisher returns the entity id of c's publisher, or "" when it stays unknown or waits in review (an
// ambiguous match opens a "publisher" review on the work, like a work credit).
func resolvePublisher(ctx context.Context, ex database.Executor, work string, c Candidate) (string, error) {
	if c.PublisherQID == "" && c.PublisherName == "" {
		return "", nil
	}

	s := &publisherStore{}
	s.Ex = ex

	if _, err := entities.Resolve(ctx, s, entities.Credit{
		ItemID: work, Role: "publisher", Kind: "organization", Source: c.Source,
		QID: c.PublisherQID, Name: c.PublisherName,
	}); err != nil {
		return "", fmt.Errorf("editions: publisher %s: %w", c.PublisherQID, err)
	}

	return s.id, nil
}

// FailureSource is the ingest_failures.source of an edition that could not be saved. It is not the job name, so a
// standing identifier conflict never parks the item itself.
const FailureSource = "editions"

// Clash is an edition identifier held by another work. Class is the ingest_failures.error_class.
type Clash struct{ ID, Class string }

// SaveAll saves the candidates of one work, resolves their publishers and points the work at its primary edition
// (original is the work's language, see SetPrimary). An ISBN or edition QID held by another work moves nothing: it opens an
// entity_match_reviews row (role "edition", the holder as the candidate, once per identifier) and is returned, so
// the caller records the failure.
func SaveAll(ctx context.Context, ex database.Executor, work, original string, in []Candidate) ([]Clash, error) {
	store := PG{Ex: ex}

	var clashes []Clash

	for _, c := range in {
		var err error
		if c.PublisherID, err = resolvePublisher(ctx, ex, work, c); err != nil {
			return nil, err
		}

		out, err := Save(ctx, store, work, c)
		if err != nil {
			return nil, err
		}

		if out != Conflict {
			continue
		}

		clash, owner, err := holder(ctx, store, work, c)
		if err != nil {
			return nil, err
		}

		if _, err := ex.ExecContext(
			ctx,
			`INSERT INTO entity_match_reviews (item_id, role, source, raw_name, raw_external_id, candidates)
VALUES ($1, 'edition', $2, $3, $3, ARRAY[$4::uuid]) ON CONFLICT (item_id, role, source, raw_name) DO NOTHING`,
			work,
			c.Source,
			clash.ID,
			owner,
		); err != nil {
			return nil, fmt.Errorf("editions: open review: %w", err)
		}

		clashes = append(clashes, clash)
	}

	if err := store.SetPrimary(ctx, work, original); err != nil {
		return nil, err
	}

	return clashes, nil
}

// holder says which identifier of c belongs to another work, and to which. The ISBN is checked first, as Save does.
func holder(ctx context.Context, store PG, work string, c Candidate) (Clash, string, error) {
	if c.ISBN != "" {
		owner, found, err := store.WorkOfISBN(ctx, c.ISBN)
		if err != nil || found && owner != work {
			return Clash{c.ISBN, "isbn_conflict"}, owner, err
		}
	}

	owner, _, err := store.WorkOfQID(ctx, c.QID)

	return Clash{c.QID, "qid_conflict"}, owner, err
}
