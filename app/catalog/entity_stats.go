package catalog

import (
	"context"
	"fmt"

	"github.com/AlexGalhardo/vukem/database"
)

// refreshEntityStatsSQL recounts public items per entity and item type in one
// statement: upsert the fresh rows, delete the ones whose entity no longer has
// a public item of that type. avg_score and scored_count stay 0/null until
// SDD 02 scores exist (P1.D.3.2).
//
// shortcut: one statement over the whole table instead of the 10k-entity
// batches of SDD 01 §7.4, upgrade when the catalogue passes ~1M item_entities rows.
const refreshEntityStatsSQL = `WITH fresh AS (
  SELECT ie.entity_id, i.type, count(DISTINCT ie.item_id)::int AS item_count
  FROM item_entities ie JOIN items i ON i.id = ie.item_id
  WHERE i.visibility = 'public'
  GROUP BY ie.entity_id, i.type
), upserted AS (
  INSERT INTO entity_stats (entity_id, type, item_count, scored_count, avg_score, refreshed_at)
  SELECT entity_id, type, item_count, 0, NULL, now() FROM fresh
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
