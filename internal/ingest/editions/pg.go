package editions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"
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
publication_year = NULLIF($8, 0), updated_at = now()
WHERE work_id = $1 AND (isbn13 = NULLIF($2, '') OR wikidata_qid = NULLIF($3, '') OR openlibrary_id = NULLIF($4, ''))`,
		work, c.ISBN, c.QID, c.OpenLibraryID, title, lang, c.MarketCountry, c.Year)
	if err != nil {
		return fmt.Errorf("editions: update: %w", err)
	}

	if n, _ := res.RowsAffected(); n > 0 {
		return nil
	}

	_, err = p.Ex.ExecContext(
		ctx,
		`INSERT INTO editions (id, work_id, isbn13, wikidata_qid, openlibrary_id, title, language,
market_country, publication_year, source)
VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), $6, $7, NULLIF($8, ''), NULLIF($9, 0), $10)`,
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
	)
	if err != nil {
		return fmt.Errorf("editions: insert: %w", err)
	}

	return nil
}

// SetPrimary points the work at the stored edition that carries one of c's identifiers.
func (p PG) SetPrimary(ctx context.Context, work string, c Candidate) error {
	_, err := p.Ex.ExecContext(ctx, `UPDATE items SET primary_edition_id = (SELECT id FROM editions WHERE work_id = $1
AND (isbn13 = NULLIF($2, '') OR wikidata_qid = NULLIF($3, '') OR openlibrary_id = NULLIF($4, ''))
ORDER BY isbn13 NULLS LAST LIMIT 1) WHERE id = $1`, work, c.ISBN, c.QID, c.OpenLibraryID)
	if err != nil {
		return fmt.Errorf("editions: set primary: %w", err)
	}

	return nil
}

// FailureSource is the ingest_failures.source of an edition that could not be saved. It is not the job name, so a
// standing ISBN conflict never parks the item itself.
const FailureSource = "editions"

// SaveAll saves the candidates of one work. An ISBN held by another work moves nothing: it opens an
// entity_match_reviews row (role "edition", the holder as the candidate, once per ISBN) and is returned, so the
// caller records the failure.
func SaveAll(ctx context.Context, ex database.Executor, work string, in []Candidate) ([]string, error) {
	store := PG{Ex: ex}

	var conflicts []string

	for _, c := range in {
		out, err := Save(ctx, store, work, c)
		if err != nil {
			return nil, err
		}

		if out != Conflict {
			continue
		}

		owner, _, err := store.WorkOfISBN(ctx, c.ISBN)
		if err != nil {
			return nil, err
		}

		if _, err := ex.ExecContext(
			ctx,
			`INSERT INTO entity_match_reviews (item_id, role, source, raw_name, raw_external_id, candidates)
VALUES ($1, 'edition', $2, $3, $3, ARRAY[$4::uuid]) ON CONFLICT (item_id, role, source, raw_name) DO NOTHING`,
			work,
			c.Source,
			c.ISBN,
			owner,
		); err != nil {
			return nil, fmt.Errorf("editions: open review: %w", err)
		}

		conflicts = append(conflicts, c.ISBN)
	}

	return conflicts, nil
}
