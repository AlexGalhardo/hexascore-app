package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261030160000CreateClientReportsTable is SDD 12 §5. Rows are anonymous:
// no user id, token id, IP or e-mail. privacy.Run deletes them after 7 days.
func m20261030160000CreateClientReportsTable() schema.Migration {
	return postgresMigration("20261030160000_create_client_reports_table",
		[]string{
			`CREATE TABLE client_reports (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  kind text NOT NULL CHECK (kind IN ('error','web_vital')),
  platform text NOT NULL CHECK (platform IN ('android','web')),
  app_version text,
  route text,
  name text,
  value real,
  message text CHECK (length(message) <= 1000),
  stack text CHECK (length(stack) <= 8192),
  received_at timestamptz NOT NULL DEFAULT now()
)`,
			`CREATE INDEX client_reports_received_idx ON client_reports (received_at)`,
		},
		[]string{`DROP TABLE IF EXISTS client_reports`},
	)
}
