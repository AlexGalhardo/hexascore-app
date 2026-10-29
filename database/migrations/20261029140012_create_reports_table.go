package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029140012CreateReportsTable is SDD 14 §5 with users.id as bigint (the SDD's text is superseded by SDD 03): member reports on a list or a user. The partial unique index keeps one open report per (reporter, target); resolved_by and reporter_id survive user deletion as NULL.
func m20261029140012CreateReportsTable() schema.Migration {
	return postgresMigration("20261029140012_create_reports_table",
		[]string{
			`CREATE TABLE reports (
  id uuid PRIMARY KEY,
  reporter_id bigint REFERENCES users(id) ON DELETE SET NULL,
  target_type text NOT NULL CHECK (target_type IN ('list','user','suggestion')),
  target_id text NOT NULL,
  reason text NOT NULL CHECK (reason IN ('spam','offensive','harassment','impersonation','other')),
  details text CHECK (char_length(details) <= 500),
  status text NOT NULL DEFAULT 'open' CHECK (status IN ('open','resolved','dismissed')),
  outcome text CHECK (outcome IN ('dismissed','list_hidden','user_banned','suggestion_rejected','no_action')),
  admin_note text CHECK (char_length(admin_note) <= 1000),
  resolved_by bigint REFERENCES users(id) ON DELETE SET NULL,
  resolved_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
)`,
			`CREATE INDEX reports_status_idx ON reports (status, created_at);
CREATE INDEX reports_target_idx ON reports (target_type, target_id);
CREATE UNIQUE INDEX reports_open_unique_idx ON reports (reporter_id, target_type, target_id) WHERE status = 'open'`,
		},
		[]string{`DROP TABLE IF EXISTS reports`},
	)
}
