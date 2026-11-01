package normalize

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/entities"
)

// Sources is the list of source_records sources that feed an item; the other
// sources (wikidata:entity) share the QID space but are never items.
var Sources = []string{
	SourceWikidata, "wikipedia:en", "wikipedia:pt", "wikipedia:es", SourceTMDB, SourceIGDB,
	SourceSteam, SourceGOG, SourceOpenLibrary, "youtube", "synopsis",
}

// existing is the stored state of an item the writer is about to update.
type existing struct {
	ID         string
	Slug       string
	Public     bool
	Date       sql.NullTime
	Precision  string
	Status     string
	ExternalID map[string]string
	Provenance Provenance
}

// SeedDataSources inserts the data_sources rows every provenance value and
// *_source column references (DQ-11). Stored rows win, so a licence the owner
// corrected after legal review is never reverted; only a pre-review placeholder
// licence is replaced.
func SeedDataSources(ctx context.Context, ex database.Executor) error {
	for _, s := range DataSources {
		if _, err := ex.ExecContext(ctx, `INSERT INTO data_sources (code, name, licence, attribution, url)
VALUES ($1, $2, $3, NULLIF($4, ''), $5) ON CONFLICT (code) DO UPDATE SET licence = excluded.licence
WHERE data_sources.licence = ANY($6::text[]) AND data_sources.licence <> excluded.licence`,
			s.Code, s.Name, s.Licence, s.Attribution, s.URL, licencePlaceholders); err != nil {
			return fmt.Errorf("seed data_sources %s: %w", s.Code, err)
		}
	}

	return nil
}

func loadExisting(ctx context.Context, ex database.Executor, qids []string) (map[string]*existing, error) {
	rows, err := ex.QueryContext(ctx, `SELECT wikidata_qid, id::text, slug, release_date, release_precision,
release_status, external_ids, provenance, visibility = 'public' FROM items WHERE wikidata_qid IN (SELECT jsonb_array_elements_text($1::jsonb))`, jsonList(qids))
	if err != nil {
		return nil, fmt.Errorf("read known items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := map[string]*existing{}

	for rows.Next() {
		var (
			qid      string
			e        existing
			ext, prv []byte
		)

		if err := rows.Scan(&qid, &e.ID, &e.Slug, &e.Date, &e.Precision, &e.Status, &ext, &prv, &e.Public); err != nil {
			return nil, fmt.Errorf("read known items: %w", err)
		}

		if err := errors.Join(json.Unmarshal(ext, &e.ExternalID), json.Unmarshal(prv, &e.Provenance)); err != nil {
			return nil, fmt.Errorf("decode stored %s: %w", qid, err)
		}

		out[qid] = &e
	}

	return out, rows.Err()
}

// chooseSlug returns the first free candidate. An item that already has a
// slug keeps it: slugs never change (SDD 06 §7.3).
func chooseSlug(ctx context.Context, ex database.Executor, it *Item) (string, error) {
	year := 0
	if it.ReleaseDate != nil {
		year = it.ReleaseDate.Year()
	}

	for _, slug := range SlugCandidates(it.Titles["en"], year, it.QID) {
		var taken bool
		if err := ex.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM items
WHERE type = $1 AND slug = $2 AND wikidata_qid <> $3)`, it.Type, slug, it.QID).Scan(&taken); err != nil {
			return "", fmt.Errorf("check slug %s: %w", slug, err)
		}

		if !taken {
			return slug, nil
		}
	}

	return "", fmt.Errorf("no free slug for %s", it.QID)
}

// positive maps 0 and out-of-range counts to NULL; the CHECK constraints
// would otherwise abort the whole batch.
func positive(n, hi int) any {
	if n < 1 || n > hi {
		return nil
	}

	return n
}

// saveItem upserts the item and its translations. Fields in it.Skip belong to
// the community and keep their stored value and provenance; cover_* is the
// cover job's. It reports whether the release date, precision or status
// differs from what was stored (a new item has nothing to differ from).
func saveItem(
	ctx context.Context,
	ex database.Executor,
	id, slug string,
	old *existing,
	it *Item,
	now time.Time,
) (bool, error) {
	date, precision, status := it.ReleaseDate, it.ReleasePrecision, it.ReleaseStatus
	prov := maps.Clone(it.Provenance)
	ext := maps.Clone(it.ExternalIDs)

	// Ids a source stopped listing (or never listed) survive: the stored object is merged, not replaced.
	if old != nil {
		ext = maps.Clone(old.ExternalID)
		maps.Copy(ext, it.ExternalIDs)
	}

	var dateArg any
	if date != nil {
		dateArg = date.Format("2006-01-02")
	}

	dateChanged := false

	if old != nil {
		// cancelled is set by an editor; ingestion never revives the item.
		if old.Status == StatusCancelled {
			status = StatusCancelled
		}

		if it.Skip["release_date"] {
			dateArg, precision, status = nullDate(old.Date), old.Precision, old.Status
			if v, ok := old.Provenance["release_date"]; ok {
				prov["release_date"] = v
			}
		}

		for field := range it.Skip {
			key, ok := strings.CutPrefix(field, "external_ids.")
			if !ok {
				continue
			}

			if v, has := old.ExternalID[key]; has {
				ext[key] = v
			} else {
				delete(ext, key)
			}

			if v, has := old.Provenance[field]; has {
				prov[field] = v
			}
		}

		for field := range it.Skip {
			if v, has := old.Provenance[field]; has {
				prov[field] = v
			}
		}

		dateChanged = !sameDate(old.Date, dateArg) || old.Precision != precision || old.Status != status
	}

	extJSON, _ := json.Marshal(ext)
	provJSON, _ := json.Marshal(prov)
	trailers, _ := json.Marshal(append([]string{}, it.Meta.Trailers...))

	if _, err := ex.ExecContext(ctx, `INSERT INTO items (id, type, slug, wikidata_qid, release_date, release_precision,
release_status, runtime_minutes, season_count, episode_count, page_count, trailers, external_ids, provenance,
popularity_prior, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12::jsonb, $13::jsonb, $14::jsonb, $15, $16, $16)
ON CONFLICT (id) DO UPDATE SET release_date = excluded.release_date, release_precision = excluded.release_precision,
release_status = excluded.release_status, runtime_minutes = excluded.runtime_minutes,
season_count = excluded.season_count, episode_count = excluded.episode_count, page_count = excluded.page_count,
trailers = excluded.trailers, external_ids = excluded.external_ids, provenance = excluded.provenance,
popularity_prior = excluded.popularity_prior, updated_at = excluded.updated_at`,
		id, it.Type, slug, it.QID, dateArg, precision, status,
		positive(it.Meta.RuntimeMinutes, 6000), positive(it.Meta.Seasons, 32767), positive(it.Meta.Episodes, 1<<30),
		positive(it.Meta.Pages, 20000), string(trailers), string(extJSON), string(provJSON),
		it.Popularity, now); err != nil {
		return false, fmt.Errorf("save item %s: %w", it.QID, err)
	}

	for loc, title := range it.Titles {
		syn := it.Synopses[loc]

		if _, err := ex.ExecContext(ctx, `INSERT INTO item_translations (item_id, locale, title, title_source, synopsis,
synopsis_source, synopsis_licence, synopsis_source_url)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''))
ON CONFLICT (item_id, locale) DO UPDATE SET
  title = CASE WHEN $9 THEN item_translations.title ELSE excluded.title END,
  title_source = CASE WHEN $9 THEN item_translations.title_source ELSE excluded.title_source END,
  synopsis = CASE WHEN $10 THEN item_translations.synopsis ELSE excluded.synopsis END,
  synopsis_source = CASE WHEN $10 THEN item_translations.synopsis_source ELSE excluded.synopsis_source END,
  synopsis_licence = CASE WHEN $10 THEN item_translations.synopsis_licence ELSE excluded.synopsis_licence END,
  synopsis_source_url = CASE WHEN $10 THEN item_translations.synopsis_source_url ELSE excluded.synopsis_source_url END`,
			id, loc, title, it.Provenance["translations."+loc+".title"], syn.Text, syn.Source, syn.Licence, syn.URL,
			it.Skip["translations."+loc+".title"], it.Skip["synopsis."+loc]); err != nil {
			return false, fmt.Errorf("save translation %s %s: %w", it.QID, loc, err)
		}
	}

	return dateChanged, nil
}

func saveRelations(ctx context.Context, ex database.Executor, id string, rel *Relations, ids map[string]string) error {
	for _, a := range rel.AltTitles {
		if _, err := ex.ExecContext(ctx, `INSERT INTO item_alt_titles (item_id, title, lang, kind, source)
VALUES ($1, $2, NULLIF($3, ''), $4, $5) ON CONFLICT (item_id, title) DO NOTHING`,
			id, a.Title, a.Lang, a.Kind, a.Source); err != nil {
			return fmt.Errorf("save alt title: %w", err)
		}
	}

	// Credits only name Wikidata QIDs here; source external ids and years arrive with the TMDB and Open Library credits.
	store := entities.PG{Ex: ex}

	for _, c := range rel.Credits {
		_, err := entities.Resolve(ctx, store, entities.Credit{
			ItemID:   id,
			Role:     c.Role,
			Kind:     c.Kind,
			Source:   wikidataSource,
			QID:      c.QID,
			Name:     c.Name,
			Position: c.Position,
		})
		if err != nil {
			return fmt.Errorf("save credit %s %s: %w", c.Role, c.QID, err)
		}
	}

	for _, l := range rel.Links {
		to, ok := ids[l.ToQID]
		if !ok {
			continue
		}

		if _, err := ex.ExecContext(ctx, `INSERT INTO item_links (from_item_id, to_item_id, relation, source)
VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`, id, to, l.Relation, l.Source); err != nil {
			return fmt.Errorf("save item link: %w", err)
		}
	}

	return nil
}

func nullDate(d sql.NullTime) any {
	if !d.Valid {
		return nil
	}

	return d.Time.Format("2006-01-02")
}

func sameDate(old sql.NullTime, arg any) bool {
	return nullDate(old) == arg
}

// jsonList encodes values as a JSON array, the driver-neutral way to bind a
// list to "IN (SELECT jsonb_array_elements_text($n::jsonb))".
func jsonList(values []string) string {
	b, _ := json.Marshal(values)

	return string(b)
}
