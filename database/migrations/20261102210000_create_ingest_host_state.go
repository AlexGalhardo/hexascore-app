package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261102210000CreateIngestHostState holds the small shared state between the web and the ingestion worker
// (RFC-001 section 3): the pressure signal ("web"), the worker heartbeat ("worker") and the persisted breaker and
// throttle state of each outbound host ("host:<name>"). One row per key, upserted in place, so the table stays at
// a few dozen rows and needs no prune.
func m20261102210000CreateIngestHostState() schema.Migration {
	return postgresMigration("20261102210000_create_ingest_host_state",
		[]string{`CREATE TABLE ingest_host_state (
  key text PRIMARY KEY,
  data jsonb NOT NULL DEFAULT '{}'::jsonb,
  updated_at timestamptz NOT NULL DEFAULT now()
)`},
		[]string{`DROP TABLE IF EXISTS ingest_host_state`},
	)
}
