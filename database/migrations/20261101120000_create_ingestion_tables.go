package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261101120000CreateIngestionTables is SDD 06 section 5: the two missing `source_records` columns and the
// run, history, failure, review and data-quality tables. `source_records` itself came 1:1 from the legacy
// import (P2.F), so `checked_at` starts at `fetched_at`. The hash check is NOT VALID: it guards new rows
// without failing the migration on an imported row.
func m20261101120000CreateIngestionTables() schema.Migration {
	return postgresMigration("20261101120000_create_ingestion_tables",
		[]string{
			`ALTER TABLE source_records
  ADD COLUMN checked_at timestamptz NOT NULL DEFAULT now(),
  ADD COLUMN last_run_id bigint`,
			`UPDATE source_records SET checked_at = fetched_at`,
			`ALTER TABLE source_records ADD CONSTRAINT source_records_hash_check CHECK (hash ~ '^[0-9a-f]{64}$') NOT VALID`,
			`CREATE INDEX source_records_item_idx ON source_records (item_id)`,
			`CREATE TABLE source_record_history (
  source text NOT NULL,
  external_id text NOT NULL,
  hash text NOT NULL,
  payload jsonb NOT NULL,
  fetched_at timestamptz NOT NULL,
  superseded_at timestamptz NOT NULL DEFAULT now(),
  superseded_by_run bigint NOT NULL,
  PRIMARY KEY (source, external_id, superseded_at)
)`,
			`CREATE TABLE job_runs (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  job text NOT NULL,
  source text,
  mode text NOT NULL CHECK (mode IN ('bootstrap','incremental','full','followup','manual','backfill','rollback')),
  status text NOT NULL CHECK (status IN ('queued','running','succeeded','partial','failed','skipped','cancelled')),
  params jsonb NOT NULL DEFAULT '{}',
  cursor jsonb,
  attempt int NOT NULL DEFAULT 0,
  seen int NOT NULL DEFAULT 0,
  changed int NOT NULL DEFAULT 0,
  failed int NOT NULL DEFAULT 0,
  skipped int NOT NULL DEFAULT 0,
  rate_limited int NOT NULL DEFAULT 0,
  total_estimate int,
  triggered_by text NOT NULL,
  error text,
  started_at timestamptz,
  heartbeat_at timestamptz,
  finished_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
)`,
			`CREATE UNIQUE INDEX job_runs_one_active ON job_runs (job, coalesce(source, ''))
  WHERE status IN ('queued','running')`,
			`CREATE TABLE ingest_failures (
  source text NOT NULL,
  external_id text NOT NULL,
  last_run_id bigint NOT NULL,
  error_class text NOT NULL,
  error text NOT NULL,
  attempts int NOT NULL DEFAULT 1,
  parked boolean NOT NULL DEFAULT false,
  first_seen timestamptz NOT NULL DEFAULT now(),
  last_seen timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (source, external_id)
)`,
			`CREATE TABLE entity_match_reviews (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  role text NOT NULL,
  source text NOT NULL,
  raw_name text NOT NULL,
  raw_year smallint,
  raw_external_id text,
  candidates uuid[] NOT NULL,
  status text NOT NULL DEFAULT 'open' CHECK (status IN ('open','linked','created','dismissed')),
  resolved_by uuid,
  resolved_at timestamptz,
  UNIQUE (item_id, role, source, raw_name)
)`,
			`CREATE TABLE dq_results (
  run_id bigint NOT NULL REFERENCES job_runs(id),
  code text NOT NULL,
  severity text NOT NULL CHECK (severity IN ('error','warn')),
  passed boolean NOT NULL,
  observed numeric,
  threshold numeric,
  sample jsonb NOT NULL DEFAULT '[]',
  checked_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (run_id, code)
)`,
			`CREATE TABLE catalog_counts (
  day date NOT NULL,
  type text NOT NULL,
  items int NOT NULL,
  PRIMARY KEY (day, type)
)`,
			`CREATE TABLE field_coverage (
  run_id bigint NOT NULL REFERENCES job_runs(id),
  type text NOT NULL,
  field text NOT NULL,
  locale text NOT NULL DEFAULT '',
  sampled int NOT NULL,
  missing int NOT NULL,
  PRIMARY KEY (run_id, type, field, locale)
)`,
		},
		[]string{
			`DROP TABLE IF EXISTS field_coverage`,
			`DROP TABLE IF EXISTS catalog_counts`,
			`DROP TABLE IF EXISTS dq_results`,
			`DROP TABLE IF EXISTS entity_match_reviews`,
			`DROP TABLE IF EXISTS ingest_failures`,
			`DROP TABLE IF EXISTS job_runs`,
			`DROP TABLE IF EXISTS source_record_history`,
			`DROP INDEX IF EXISTS source_records_item_idx`,
			`ALTER TABLE source_records DROP CONSTRAINT IF EXISTS source_records_hash_check`,
			`ALTER TABLE source_records DROP COLUMN IF EXISTS last_run_id, DROP COLUMN IF EXISTS checked_at`,
		},
	)
}
