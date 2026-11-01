package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261101170000AddEntitiesNameNormIndex backs step 3 of the credit matcher (internal/ingest/entities): without it
// every credit without an id scans the whole entities table.
func m20261101170000AddEntitiesNameNormIndex() schema.Migration {
	m := postgresMigration(
		"20261101170000_add_entities_name_norm_index",
		[]string{
			`CREATE INDEX CONCURRENTLY IF NOT EXISTS entities_name_norm_idx ON entities (kind, hs_norm(default_name))`,
		},
		[]string{`DROP INDEX CONCURRENTLY IF EXISTS entities_name_norm_idx`},
	)
	m.NoTransaction = true

	return m
}
