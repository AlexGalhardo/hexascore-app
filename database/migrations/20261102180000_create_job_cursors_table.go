package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261102180000CreateJobCursorsTable holds the resume point of chunked jobs (scoring.recompute_all). One row per
// job, deleted when the pass completes, so the table stays at a handful of rows and needs no prune.
func m20261102180000CreateJobCursorsTable() schema.Migration {
	return postgresMigration("20261102180000_create_job_cursors_table",
		[]string{`CREATE TABLE job_cursors (
  job text PRIMARY KEY,
  cursor jsonb NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
)`},
		[]string{`DROP TABLE IF EXISTS job_cursors`},
	)
}
