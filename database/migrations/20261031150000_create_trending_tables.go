package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261031150000CreateTrendingTables creates the daily activity counters and the hourly trending
// snapshots (SDD 08 section 5). Exclusion is items.visibility, so there is no exclusions table.
func m20261031150000CreateTrendingTables() schema.Migration {
	return postgresMigration("20261031150000_create_trending_tables",
		[]string{
			`CREATE TABLE item_activity_daily (
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  day date NOT NULL,
  item_type text NOT NULL,
  views int NOT NULL DEFAULT 0,
  search_clicks int NOT NULL DEFAULT 0,
  ratings_new int NOT NULL DEFAULT 0,
  ratings_updated int NOT NULL DEFAULT 0,
  status_want int NOT NULL DEFAULT 0,
  status_in_progress int NOT NULL DEFAULT 0,
  status_completed int NOT NULL DEFAULT 0,
  status_dropped int NOT NULL DEFAULT 0,
  list_adds int NOT NULL DEFAULT 0,
  suggestions int NOT NULL DEFAULT 0,
  ratings_new_w real NOT NULL DEFAULT 0,
  ratings_updated_w real NOT NULL DEFAULT 0,
  want_w real NOT NULL DEFAULT 0,
  in_progress_w real NOT NULL DEFAULT 0,
  completed_w real NOT NULL DEFAULT 0,
  dropped_w real NOT NULL DEFAULT 0,
  list_adds_w real NOT NULL DEFAULT 0,
  suggestions_w real NOT NULL DEFAULT 0,
  PRIMARY KEY (item_id, day)
)`,
			`CREATE INDEX item_activity_daily_type_day_idx ON item_activity_daily (item_type, day)`,
			`CREATE TABLE trending_snapshots (
  computed_at timestamptz NOT NULL,
  item_type text NOT NULL,
  rank smallint NOT NULL CHECK (rank BETWEEN 1 AND 10),
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  trend_score real NOT NULL,
  basis text NOT NULL CHECK (basis IN ('trend', 'prior')),
  PRIMARY KEY (computed_at, item_type, rank),
  UNIQUE (computed_at, item_type, item_id)
)`,
		},
		[]string{`DROP TABLE IF EXISTS trending_snapshots`, `DROP TABLE IF EXISTS item_activity_daily`},
	)
}
