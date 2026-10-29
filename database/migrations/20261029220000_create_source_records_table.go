package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029220000CreateSourceRecordsTable is SDD 06 §5 `source_records` in the legacy shape, so the
// catalog import stage (P2.F.2.4) copies it 1:1. P6.A.1.2 adds `checked_at`, `last_run_id` and
// `source_records_item_idx`.
func m20261029220000CreateSourceRecordsTable() schema.Migration {
	return postgresMigration("20261029220000_create_source_records_table",
		[]string{
			`CREATE TABLE source_records (
  source text NOT NULL,
  external_id text NOT NULL,
  payload jsonb NOT NULL,
  hash text NOT NULL,
  fetched_at timestamptz NOT NULL DEFAULT now(),
  normalized_at timestamptz,
  item_id uuid REFERENCES items(id) ON DELETE SET NULL,
  PRIMARY KEY (source, external_id)
)`,
			`CREATE INDEX source_records_pending_idx ON source_records (source)
  WHERE normalized_at IS NULL OR normalized_at < fetched_at`,
		},
		[]string{`DROP TABLE IF EXISTS source_records`},
	)
}
