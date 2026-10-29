package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029190000CreateLegacyImportRunsTable holds one checkpoint per import stage (SDD 11 §7.6
// step 5, SDD 06 §7.9): the last imported key is written in the same transaction as its batch, so a
// killed run resumes after it. Dropped with legacy_import at T+30 d.
func m20261029190000CreateLegacyImportRunsTable() schema.Migration {
	return postgresMigration("20261029190000_create_legacy_import_runs_table",
		[]string{
			`CREATE TABLE legacy_import_runs (
  stage text PRIMARY KEY,
  status text NOT NULL CHECK (status IN ('running','done','failed')),
  last_key text,
  rows bigint NOT NULL DEFAULT 0,
  error text,
  started_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  finished_at timestamptz
)`,
		},
		[]string{`DROP TABLE IF EXISTS legacy_import_runs`},
	)
}
