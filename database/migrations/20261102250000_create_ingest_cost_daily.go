package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261102250000CreateIngestCostDaily holds one row per day of what ingestion spent (RFC-001 section 8): LLM dollars
// (CW-10 writes them), requests per host, bytes downloaded and cover bytes written. The 30-minute digest sums the
// month to date. At one row a day the table needs no prune; no user data.
func m20261102250000CreateIngestCostDaily() schema.Migration {
	return postgresMigration("20261102250000_create_ingest_cost_daily",
		[]string{`CREATE TABLE ingest_cost_daily (
  day date PRIMARY KEY,
  llm_usd numeric(10, 4) NOT NULL DEFAULT 0,
  requests_by_host jsonb NOT NULL DEFAULT '{}'::jsonb,
  bytes_down bigint NOT NULL DEFAULT 0,
  covers_bytes bigint NOT NULL DEFAULT 0
)`},
		[]string{`DROP TABLE IF EXISTS ingest_cost_daily`},
	)
}
