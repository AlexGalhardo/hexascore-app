// Package offers builds the where-to-watch and where-to-buy rows of RFC-003: the stages that need no network
// (Derive), the TMDB watch-provider stage and the verification job.
package offers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// DeriveJob is the job_runs.job of the derivation stage.
const DeriveJob = "offers.derive"

// BatchSize is how many items share one transaction.
const BatchSize = 500

// offerKey is the unique constraint of offers (item, edition, country, store, kind); ON CONFLICT names it because
// the key holds a nullable edition_id (NULLS NOT DISTINCT).
const offerKey = "offers_item_id_edition_id_country_store_kind_key"

// storesSQL derives one offer per store id kept in items.external_ids. Every id must match the shape of that
// store's ids before it is put in a URL, so an odd value in an external source can never add a path, a query or a
// host. A row the admin entered or the verify job deactivated is not overwritten: only url and external_id of a
// wikidata row follow the source. $1 is the page's item ids.
const storesSQL = `
WITH m(key, store, kind, pattern) AS (VALUES
  ('steam',       'steam',       'buy_digital', '^[0-9]{1,10}$'),
  ('gog',         'gog',         'buy_digital', '^[a-z]+/[a-z0-9_-]{1,100}$'),
  ('epic',        'epic',        'buy_digital', '^[a-z0-9-]{1,100}$'),
  ('google_play', 'google_play', 'mobile',      '^[A-Za-z][A-Za-z0-9_]*(\.[A-Za-z0-9_]+)+$'),
  ('app_store',   'app_store',   'mobile',      '^[0-9]{5,12}$'))
INSERT INTO offers (item_id, country, store, kind, url, external_id, source, licence, confidence)
SELECT i.id, 'ZZ', s.code, m.kind, replace(s.url_template, '{id}', i.external_ids ->> m.key),
  i.external_ids ->> m.key, 'wikidata', 'CC0-1.0', 90
FROM items i
JOIN m ON (i.external_ids ->> m.key) ~ m.pattern
JOIN stores s ON s.code = m.store AND s.active AND s.url_template IS NOT NULL
WHERE i.id = ANY($1::uuid[])
ON CONFLICT ON CONSTRAINT ` + offerKey + ` DO UPDATE SET url = EXCLUDED.url, external_id = EXCLUDED.external_id
WHERE offers.source = 'wikidata' AND (offers.url, offers.external_id) IS DISTINCT FROM (EXCLUDED.url, EXCLUDED.external_id)`

// booksSQL derives an Amazon search-by-ISBN offer per edition and marketplace (RFC-003 section 3). The ISBN must be
// 13 digits. The row stays editor-neutral: the affiliate tag is added when the page renders, never stored.
const booksSQL = `
INSERT INTO offers (item_id, edition_id, country, store, kind, url, external_id, source, licence, confidence)
SELECT e.work_id, e.id, m.country, s.code, 'buy', replace(s.url_template, '{isbn13}', e.isbn13::text),
  e.isbn13::text, 'editions', 'derived', 70
FROM editions e
JOIN items i ON i.id = e.work_id AND i.visibility = 'public'
JOIN (VALUES ('BR', 'amazon_br'), ('US', 'amazon_us')) m(country, store) ON true
JOIN stores s ON s.code = m.store AND s.active AND s.url_template IS NOT NULL
WHERE e.work_id = ANY($1::uuid[]) AND e.isbn13::text ~ '^[0-9]{13}$'
ON CONFLICT ON CONSTRAINT ` + offerKey + ` DO NOTHING`

// Derive builds the offers that need no network: store offers from the ids already in the catalogue and Amazon
// offers from edition ISBNs. It walks the public items by id, one page per transaction; every statement is an upsert,
// so a rerun or a resumed run changes nothing that is already right.
type Derive struct {
	// Batch overrides BatchSize (tests).
	Batch int
}

type cursor struct {
	After string `json:"after"`
}

// Step processes one keyset page. The cursor holds the last item id seen.
func (d Derive) Step(ctx context.Context, tx *database.Tx, r run.Run) (run.Result, error) {
	var cur cursor
	if len(r.Cursor) > 0 && string(r.Cursor) != "null" {
		if err := json.Unmarshal(r.Cursor, &cur); err != nil {
			return run.Result{}, fmt.Errorf("offers derive: cursor: %w: %w", err, queue.ErrPermanent)
		}
	}

	batch := d.Batch
	if batch <= 0 {
		batch = BatchSize
	}

	ids, err := page(ctx, tx, cur.After, batch)
	if err != nil {
		return run.Result{}, err
	}

	res := run.Result{Done: len(ids) < batch, Targets: len(ids), Seen: len(ids)}
	if len(ids) == 0 {
		res.Cursor, err = json.Marshal(cur)

		return res, err
	}

	arr := "{" + strings.Join(ids, ",") + "}"

	for _, q := range []string{storesSQL, booksSQL} {
		out, err := tx.ExecContext(ctx, q, arr)
		if err != nil {
			return run.Result{}, fmt.Errorf("offers derive: %w", err)
		}

		n, _ := out.RowsAffected()
		res.Changed += int(n)
	}

	cur.After = ids[len(ids)-1]
	res.Cursor, err = json.Marshal(cur)

	return res, err
}

// page lists the next public item ids after the cursor.
func page(ctx context.Context, tx *database.Tx, after string, limit int) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id::text FROM items
WHERE visibility = 'public' AND ($1 = '' OR id > $1::uuid) ORDER BY id LIMIT $2`, after, limit)
	if err != nil {
		return nil, fmt.Errorf("offers derive: page: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var ids []string

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}
