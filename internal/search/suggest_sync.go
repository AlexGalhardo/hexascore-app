package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
)

// SuggestBatch is the most search_dirty rows one transaction drains (SDD 07 §7.5).
const SuggestBatch = 5000

// SyncSuggest drains up to limit search_dirty rows: for each claimed ref it deletes its search_suggest rows, inserts
// the current ones and removes the dirty row, in one transaction. Claims use FOR UPDATE SKIP LOCKED, so concurrent
// runners take disjoint refs; a ref queued again while claimed waits for the commit and stays queued. It returns the
// number of refs synced and is idempotent.
func SyncSuggest(ctx context.Context, db *database.DB, limit int) (int, error) {
	var synced int

	err := db.Tx(ctx, func(tx *database.Tx) error {
		items, entities, err := claimDirty(ctx, tx, limit)
		if err != nil {
			return err
		}

		synced = len(items) + len(entities)
		if synced == 0 {
			return nil
		}

		if _, err := tx.ExecContext(
			ctx,
			`DELETE FROM search_suggest WHERE ref_id = ANY($1::uuid[]) OR ref_id = ANY($2::uuid[])`,
			uuidArray(items),
			uuidArray(entities),
		); err != nil {
			return fmt.Errorf("delete stale suggest rows: %w", err)
		}

		if len(items) > 0 {
			if _, err := tx.ExecContext(ctx, itemSuggestSQL, uuidArray(items)); err != nil {
				return fmt.Errorf("insert item suggest rows: %w", err)
			}
		}

		if len(entities) > 0 {
			if _, err := tx.ExecContext(ctx, entitySuggestSQL, uuidArray(entities)); err != nil {
				return fmt.Errorf("insert entity suggest rows: %w", err)
			}
		}

		if _, err := tx.ExecContext(
			ctx,
			`DELETE FROM search_dirty WHERE ref_id = ANY($1::uuid[]) OR ref_id = ANY($2::uuid[])`,
			uuidArray(items),
			uuidArray(entities),
		); err != nil {
			return fmt.Errorf("clear dirty rows: %w", err)
		}

		return nil
	})

	if err == nil && synced > 0 {
		BumpEpoch()
	}

	return synced, err
}

// DrainSuggest runs SyncSuggest until the queue is empty or ctx ends.
func DrainSuggest(ctx context.Context, db *database.DB) error {
	for {
		n, err := SyncSuggest(ctx, db, SuggestBatch)
		if err != nil || n < SuggestBatch {
			return err
		}
	}
}

func claimDirty(ctx context.Context, tx *database.Tx, limit int) (items, entities []string, err error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT ref_id::text, kind FROM search_dirty ORDER BY queued_at LIMIT $1 FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return nil, nil, fmt.Errorf("claim dirty rows: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, kind string
		if err := rows.Scan(&id, &kind); err != nil {
			return nil, nil, err
		}

		if kind == "item" {
			items = append(items, id)
		} else {
			entities = append(entities, id)
		}
	}

	return items, entities, rows.Err()
}

// uuidArray renders ids as a Postgres array literal; ids come from a uuid column, never from user input.
func uuidArray(ids []string) string { return "{" + strings.Join(ids, ",") + "}" }

// Only public items are suggested. Alternative titles and edition titles are locale '*'; ON CONFLICT keeps the
// first label when two sources normalise alike.
const itemSuggestSQL = `INSERT INTO search_suggest (ref_id, kind, locale, label, label_norm, slug, year, thumb_key, popularity)
SELECT i.id, i.type, l.locale, left(l.label, 300), hs_norm(l.label), i.slug, extract(year FROM i.release_date)::smallint,
  i.cover_thumb_key, i.popularity_prior
FROM items i
JOIN (
  SELECT item_id, locale, title AS label FROM item_translations
  UNION ALL SELECT item_id, '*', title FROM item_alt_titles
  UNION ALL SELECT work_id, '*', title FROM editions
) l ON l.item_id = i.id
WHERE i.id = ANY($1::uuid[]) AND i.visibility = 'public' AND hs_norm(l.label) <> ''
ON CONFLICT (ref_id, locale, label_norm) DO NOTHING`

const entitySuggestSQL = `INSERT INTO search_suggest (ref_id, kind, locale, label, label_norm, slug, year, thumb_key, popularity)
SELECT e.id, e.kind, l.locale, left(l.label, 300), hs_norm(l.label), e.slug,
  COALESCE(e.founded_year, extract(year FROM e.born_on))::smallint, e.image_key, 0
FROM entities e
JOIN (
  SELECT entity_id, locale, name AS label FROM entity_translations
  UNION ALL SELECT id, '*', default_name FROM entities
) l ON l.entity_id = e.id
WHERE e.id = ANY($1::uuid[]) AND e.kind IN ('person','organization','franchise') AND hs_norm(l.label) <> ''
ON CONFLICT (ref_id, locale, label_norm) DO NOTHING`

// Rebuild queues every item and entity and drains the queue: the full suggest rebuild after a bulk load such as
// the legacy import (SDD 11 §7.6 step 5). The tsvectors are trigger-maintained, so only the suggest rows need it.
func Rebuild(ctx context.Context, db *database.DB) error {
	if _, err := db.ExecContext(ctx, `INSERT INTO search_dirty (ref_id, kind)
SELECT id, 'item' FROM items UNION ALL SELECT id, 'entity' FROM entities
ON CONFLICT (ref_id) DO NOTHING`); err != nil {
		return fmt.Errorf("queue search rebuild: %w", err)
	}

	return DrainSuggest(ctx, db)
}

// searchIndexes are every index SDD 07 §5 creates; Reindex rewrites them to reclaim bloat.
var searchIndexes = []string{
	"item_translations_tsv_gin", "item_translations_tsv_api_gin", "item_translations_title_trgm",
	"item_alt_titles_trgm", "editions_title_trgm", "entity_translations_name_trgm", "entities_name_trgm",
	"search_suggest_prefix", "search_suggest_trgm",
}

// Reindex rewrites every search index without blocking writes (REINDEX CONCURRENTLY cannot run inside a
// transaction, so each statement runs on its own) and then rebuilds the suggest rows (runbook R6).
func Reindex(ctx context.Context, db *database.DB) error {
	for _, name := range searchIndexes {
		// name comes from the constant list above, never from input.
		if _, err := db.ExecContext(ctx, `REINDEX INDEX CONCURRENTLY `+name); err != nil {
			return fmt.Errorf("search: reindex %s: %w", name, err)
		}
	}

	return Rebuild(ctx, db)
}
