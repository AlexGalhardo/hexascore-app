package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029100009CreateEntityStatsTable is SDD 01 §5.2 `entity_stats`: per entity and item type counters refreshed hourly (§7.4).
func m20261029100009CreateEntityStatsTable() schema.Migration {
	return schema.Migration{
		ID: "20261029100009_create_entity_stats_table",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE TABLE entity_stats (
  entity_id uuid NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
  type text NOT NULL,
  item_count int NOT NULL,
  scored_count int NOT NULL,
  avg_score real,
  refreshed_at timestamptz NOT NULL,
  PRIMARY KEY (entity_id, type)
)`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DROP TABLE IF EXISTS entity_stats`)
		},
	}
}
