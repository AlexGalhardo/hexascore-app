package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029150500CreateDataExportsTable is SDD 14 §5 with users.id as bigint (SDD 03): one row per LGPD export
// request; the 7-day cooldown reads the latest non-failed row, the hourly sweep expires ready files.
func m20261029150500CreateDataExportsTable() schema.Migration {
	return postgresMigration("20261029150500_create_data_exports_table",
		[]string{
			`CREATE TABLE data_exports (
  id uuid PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status text NOT NULL CHECK (status IN ('queued','running','ready','failed','expired')),
  path text,
  size_bytes bigint,
  requested_at timestamptz NOT NULL DEFAULT now(),
  ready_at timestamptz,
  expires_at timestamptz
)`,
			`CREATE INDEX data_exports_user_idx ON data_exports (user_id, requested_at DESC)`,
			`CREATE INDEX data_exports_expiry_idx ON data_exports (expires_at) WHERE status = 'ready'`,
		},
		[]string{`DROP TABLE IF EXISTS data_exports`},
	)
}
