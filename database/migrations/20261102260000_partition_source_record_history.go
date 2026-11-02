package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261102260000PartitionSourceRecordHistory applies the storage measures of ADR 0025: source_records.payload
// switches to lz4 TOAST compression for new rows, and source_record_history becomes a table partitioned by month of
// superseded_at, so retention can DROP a month instead of DELETE-ing it and the cold month can be exported to the
// bucket first (app/jobs/source_records_export.go). The rows are copied across; a default partition catches any
// month nobody created yet, and store.EnsureHistoryPartitions (run by the export job) creates the months ahead.
// Rollback restores the plain table, copying the rows back.
func m20261102260000PartitionSourceRecordHistory() schema.Migration {
	return postgresMigration("20261102260000_partition_source_record_history",
		[]string{
			`ALTER TABLE source_records ALTER COLUMN payload SET COMPRESSION lz4`,
			`ALTER TABLE source_record_history RENAME TO source_record_history_plain`,
			`ALTER TABLE source_record_history_plain RENAME CONSTRAINT source_record_history_pkey TO source_record_history_plain_pkey`,
			`CREATE TABLE source_record_history (
  source text NOT NULL,
  external_id text NOT NULL,
  hash text NOT NULL,
  payload jsonb COMPRESSION lz4 NOT NULL,
  fetched_at timestamptz NOT NULL,
  superseded_at timestamptz NOT NULL DEFAULT now(),
  superseded_by_run bigint NOT NULL,
  checked_at timestamptz,
  CONSTRAINT source_record_history_pkey PRIMARY KEY (source, external_id, superseded_at)
) PARTITION BY RANGE (superseded_at)`,
			`CREATE TABLE source_record_history_default PARTITION OF source_record_history DEFAULT`,
			`DO $$
DECLARE m timestamptz;
BEGIN
  FOR m IN SELECT generate_series(
    date_trunc('month', now() AT TIME ZONE 'UTC') AT TIME ZONE 'UTC',
    (date_trunc('month', now() AT TIME ZONE 'UTC') + interval '3 months') AT TIME ZONE 'UTC',
    interval '1 month') LOOP
    EXECUTE format('CREATE TABLE source_record_history_%s PARTITION OF source_record_history FOR VALUES FROM (%L) TO (%L)',
      to_char(m AT TIME ZONE 'UTC', 'YYYY_MM'), m, m + interval '1 month');
  END LOOP;
END $$`,
			`INSERT INTO source_record_history (source, external_id, hash, payload, fetched_at, superseded_at, superseded_by_run, checked_at)
SELECT source, external_id, hash, payload, fetched_at, superseded_at, superseded_by_run, checked_at
FROM source_record_history_plain`,
			`DROP TABLE source_record_history_plain`,
		},
		[]string{
			`CREATE TABLE source_record_history_plain (
  source text NOT NULL,
  external_id text NOT NULL,
  hash text NOT NULL,
  payload jsonb NOT NULL,
  fetched_at timestamptz NOT NULL,
  superseded_at timestamptz NOT NULL DEFAULT now(),
  superseded_by_run bigint NOT NULL,
  checked_at timestamptz,
  CONSTRAINT source_record_history_plain_pkey PRIMARY KEY (source, external_id, superseded_at)
)`,
			`INSERT INTO source_record_history_plain (source, external_id, hash, payload, fetched_at, superseded_at, superseded_by_run, checked_at)
SELECT source, external_id, hash, payload, fetched_at, superseded_at, superseded_by_run, checked_at
FROM source_record_history`,
			`DROP TABLE source_record_history`,
			`ALTER TABLE source_record_history_plain RENAME TO source_record_history`,
			`ALTER TABLE source_record_history RENAME CONSTRAINT source_record_history_plain_pkey TO source_record_history_pkey`,
			`ALTER TABLE source_records ALTER COLUMN payload SET COMPRESSION default`,
		},
	)
}
