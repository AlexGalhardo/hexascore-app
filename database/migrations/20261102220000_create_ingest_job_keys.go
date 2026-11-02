package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261102220000CreateIngestJobKeys holds the idempotency keys of automatic ingest dispatches (RFC-001 section 4):
// sha256 of job, source, mode, params, cursor and UTC day. Rows are pruned after 7 days (run.Prune). The job_id
// lookup index comes in the next migration because it is built concurrently.
func m20261102220000CreateIngestJobKeys() schema.Migration {
	return postgresMigration("20261102220000_create_ingest_job_keys",
		[]string{`CREATE TABLE ingest_job_keys (
  key text PRIMARY KEY,
  job_id bigint NOT NULL,
  state text NOT NULL CHECK (state IN ('queued', 'running', 'done', 'failed')),
  created_at timestamptz NOT NULL DEFAULT now(),
  done_at timestamptz
)`},
		[]string{`DROP TABLE IF EXISTS ingest_job_keys`},
	)
}

// m20261102220100AddIngestJobKeysJobIdx serves Finish (state by job_id). One concurrent
// statement per NoTransaction migration.
func m20261102220100AddIngestJobKeysJobIdx() schema.Migration {
	m := postgresMigration("20261102220100_add_ingest_job_keys_job_idx",
		[]string{`CREATE INDEX CONCURRENTLY IF NOT EXISTS ingest_job_keys_job_idx ON ingest_job_keys (job_id)`},
		[]string{`DROP INDEX CONCURRENTLY IF EXISTS ingest_job_keys_job_idx`},
	)
	m.NoTransaction = true

	return m
}
