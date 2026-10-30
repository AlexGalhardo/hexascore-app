package legacyimport

import (
	"context"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// catalogSources are the data_sources rows the imported titles and synopses reference (SDD 01 §5.2,
// SDD 06 §7.9 "seed data_sources"): legacy titles are Wikidata labels, legacy synopses are Wikipedia
// extracts (CC BY-SA 4.0, normalize.py SYNOPSIS_LICENSE) or HexaScore's own rewritten text.
const catalogSources = `INSERT INTO data_sources (code, name, licence, attribution, url) VALUES
  ('wikidata', 'Wikidata', 'CC0-1.0', NULL, 'https://www.wikidata.org'),
  ('wikipedia', 'Wikipedia', 'CC-BY-SA-4.0', 'Wikipedia (CC BY-SA 4.0)', 'https://www.wikipedia.org'),
  ('hexascore', 'HexaScore', 'proprietary', NULL, 'https://hexascore.app')
ON CONFLICT (code) DO NOTHING`

// afterKey is the batch's exclusive lower bound: $1 is "" on the first batch.
const afterKey = `coalesce(nullif($1, '')::uuid, '00000000-0000-0000-0000-000000000000')`

// inBatch bounds col to the batch's legacy item ids, ($1, $2].
func inBatch(col string) string {
	return col + ` > ` + afterKey + ` AND ` + col + ` <= $2::uuid`
}

// coverKey keeps the legacy Blob pathname as the bucket key (SDD 01 §5.4, 06-4); a cover that never
// went through the Blob (covers.py without a token stores the source URL) gets no key and is
// re-fetched by the cover job (SDD 06).
func coverKey(field, suffix string) string {
	return `substring(i.meta->>'` + field + `' from '^https://[^/]+\.blob\.vercel-storage\.com/(covers/[0-9a-f-]+-[0-9a-f]{16}` +
		suffix + `\.webp)$')`
}

// catalogSQL copies one batch of legacy items with their rows (SDD 01 §5.4). Each statement takes
// $1 < item id <= $2, and the import time as $3 when stamped; all are idempotent.
var catalogSQL = []struct {
	name, query string
	stamped     bool
}{
	// Score columns stay empty: SDD 02 recompute fills them (P2.F.4.1).
	{"items", `INSERT INTO items (id, type, slug, wikidata_qid, release_date, release_precision, release_status,
  visibility, runtime_minutes, season_count, episode_count, page_count, trailers, external_ids, cover_key,
  cover_thumb_key, cover_attribution, popularity_prior, created_at, updated_at)
SELECT i.id, i.type::text, i.slug, i.wikidata_qid, i.release_date,
  CASE WHEN i.release_date IS NULL THEN 'unknown' ELSE 'day' END,
  CASE WHEN i.release_date <= $3::date THEN 'released' ELSE 'announced' END,
  'public',
  round((i.meta->>'runtimeMinutes')::numeric)::int, round((i.meta->>'seasons')::numeric)::smallint,
  round((i.meta->>'episodes')::numeric)::int, round((i.meta->>'pages')::numeric)::int,
  coalesce(i.meta->'trailers', '[]'), i.external_ids, ` + coverKey("coverUrl", "") + `, ` +
		coverKey("coverThumbUrl", "-t") + `, i.meta->'coverAttribution', i.popularity, i.created_at, i.updated_at
FROM legacy_import.items i WHERE ` + inBatch("i.id") + `
ON CONFLICT (id) DO NOTHING`, true},
	// Titles are Wikidata labels; the synopsis source comes from its URL host, and an unknown host fails
	// the batch on the data_sources foreign key instead of guessing.
	{"item_translations", `INSERT INTO item_translations (item_id, locale, title, title_source, synopsis,
  synopsis_source, synopsis_licence, synopsis_source_url)
SELECT t.item_id, t.locale::text, t.title, 'wikidata', t.synopsis,
  CASE WHEN t.synopsis IS NULL THEN NULL
    WHEN t.synopsis_source_url IS NULL THEN 'hexascore'
    WHEN substring(t.synopsis_source_url from '^https?://([^/:?#]+)') ~ '(^|\.)wikipedia\.org$' THEN 'wikipedia'
    ELSE substring(t.synopsis_source_url from '^https?://([^/:?#]+)') END,
  coalesce(t.synopsis_license, CASE WHEN t.synopsis IS NOT NULL AND t.synopsis_source_url IS NULL
    THEN 'proprietary' END),
  t.synopsis_source_url
FROM legacy_import.item_translations t WHERE ` + inBatch("t.item_id") + `
ON CONFLICT (item_id, locale) DO NOTHING`, false},
	{"affiliate_links", `INSERT INTO affiliate_links (item_id, store, url, active)
SELECT a.item_id, a.store::text, a.url, a.active FROM legacy_import.affiliate_links a
WHERE ` + inBatch("a.item_id") + `
ON CONFLICT (item_id, store) DO NOTHING`, false},
	{
		"source_records",
		`INSERT INTO source_records (source, external_id, payload, hash, fetched_at, normalized_at, item_id)
SELECT s.source, s.external_id, s.payload, s.hash, s.fetched_at, s.normalized_at, s.item_id
FROM legacy_import.source_records s WHERE ` + inBatch("s.item_id") + `
ON CONFLICT (source, external_id) DO NOTHING`,
		false,
	},
}

// unlinkedRecords copies the source records no item points at; it runs with the first batch.
const unlinkedRecords = `INSERT INTO source_records (source, external_id, payload, hash, fetched_at, normalized_at)
SELECT s.source, s.external_id, s.payload, s.hash, s.fetched_at, s.normalized_at
FROM legacy_import.source_records s WHERE s.item_id IS NULL
ON CONFLICT (source, external_id) DO NOTHING`

// catalogBatch imports the next n legacy items with their translations, affiliate links and source
// records. external_scores is never read (D-09): it stays only in the legacy dump.
func catalogBatch(now func() time.Time) BatchFunc {
	return func(ctx context.Context, tx *database.Tx, after string, n int) (string, int, error) {
		var (
			last  string
			items int
		)

		err := tx.QueryRowContext(ctx, `SELECT coalesce(max(id::text), ''), count(*) FROM (SELECT id FROM legacy_import.items
WHERE id > `+afterKey+` ORDER BY id LIMIT $2) b`, after, n).
			Scan(&last, &items)
		if err != nil {
			return "", 0, err
		}

		if after == "" {
			for _, q := range []string{catalogSources, unlinkedRecords} {
				if _, err := tx.ExecContext(ctx, q); err != nil {
					return "", 0, fmt.Errorf("catalog seed: %w", err)
				}
			}
		}

		if items == 0 {
			return "", 0, nil
		}

		at := now().UTC()

		for _, s := range catalogSQL {
			args := []any{after, last}
			if s.stamped {
				args = append(args, at)
			}

			if _, err := tx.ExecContext(ctx, s.query, args...); err != nil {
				return "", 0, fmt.Errorf("catalog %s: %w", s.name, err)
			}
		}

		return last, items, nil
	}
}
