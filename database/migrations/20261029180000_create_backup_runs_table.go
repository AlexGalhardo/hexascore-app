package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029180000CreateBackupRunsTable is SDD 11 §5: one row per backup, drill or off-site copy; the
// index serves the latest-run lookups of the missed-backup alert and the 7-day size median (§7.3).
func m20261029180000CreateBackupRunsTable() schema.Migration {
	return postgresMigration("20261029180000_create_backup_runs_table",
		[]string{
			`CREATE TABLE backup_runs (
  id uuid PRIMARY KEY,
  kind text NOT NULL CHECK (kind IN ('daily','monthly','drill','offsite')),
  object_key text,
  bytes bigint,
  sha256 text,
  status text NOT NULL CHECK (status IN ('running','uploaded','verified','failed')),
  duration_s integer,
  error text,
  started_at timestamptz NOT NULL,
  finished_at timestamptz
)`,
			`CREATE INDEX backup_runs_kind_started_idx ON backup_runs (kind, started_at DESC)`,
		},
		[]string{`DROP TABLE IF EXISTS backup_runs`},
	)
}
