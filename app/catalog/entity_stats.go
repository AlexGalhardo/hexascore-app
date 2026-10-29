package catalog

import (
	"context"
	"fmt"

	"github.com/AlexGalhardo/vukem/database"
)

// refreshEntityStatsSQL recounts public items per entity and item type in one
// statement: upsert the fresh rows, delete the ones whose entity no longer has
// a public item of that type. scored_count and avg_score cover items with a visible score (N >= 5, SDD 02 §7.1);
//
// shortcut: one statement over the whole table instead of the 10k-entity
// batches of SDD 01 §7.4, upgrade when the catalogue passes ~1M item_entities rows.
const refreshEntityStatsSQL = `WITH fresh AS (
  SELECT d.entity_id, d.type, count(*)::int AS item_count,
    count(*) FILTER (WHERE d.rating_count >= 5 AND d.score IS NOT NULL)::int AS scored_count,
    avg(d.score) FILTER (WHERE d.rating_count >= 5)::real AS avg_score
  FROM (
    SELECT DISTINCT ie.entity_id, i.id, i.type, i.rating_count, i.score
    FROM item_entities ie JOIN items i ON i.id = ie.item_id
    WHERE i.visibility = 'public'
  ) d
  GROUP BY d.entity_id, d.type
), upserted AS (
  INSERT INTO entity_stats (entity_id, type, item_count, scored_count, avg_score, refreshed_at)
  SELECT entity_id, type, item_count, scored_count, avg_score, now() FROM fresh
  ON CONFLICT (entity_id, type) DO UPDATE
  SET item_count = EXCLUDED.item_count, scored_count = EXCLUDED.scored_count,
      avg_score = EXCLUDED.avg_score, refreshed_at = EXCLUDED.refreshed_at
)
DELETE FROM entity_stats es
WHERE NOT EXISTS (SELECT 1 FROM fresh f WHERE f.entity_id = es.entity_id AND f.type = es.type)`

// RefreshEntityStats recomputes entity_stats (SDD 01 §7.4). It is idempotent.
func RefreshEntityStats(ctx context.Context, db *database.DB) error {
	if _, err := db.ExecContext(ctx, refreshEntityStatsSQL); err != nil {
		return fmt.Errorf("refresh entity stats: %w", err)
	}
	return nil
}
