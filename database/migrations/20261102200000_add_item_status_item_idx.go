package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261102200000AddItemStatusItemIdx indexes item_status by item: the PK leads with user_id, so the per-item status
// counts scanned the table (audit T-16). One concurrent statement per NoTransaction migration (see
// trendingRollupIndex).
func m20261102200000AddItemStatusItemIdx() schema.Migration {
	m := postgresMigration(
		"20261102200000_add_item_status_item_idx",
		[]string{
			`CREATE INDEX CONCURRENTLY IF NOT EXISTS item_status_item_status_idx ON item_status (item_id, status)`,
		},
		[]string{`DROP INDEX CONCURRENTLY IF EXISTS item_status_item_status_idx`},
	)
	m.NoTransaction = true

	return m
}
