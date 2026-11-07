package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261107000000AddItemsListScoreIdx serves the default listing order (catalog listing, sort by score, covered public
// items of one type): the expression is the ORDER BY of the query, so page one reads 41 index entries instead of
// sorting every item of the type. EXPLAIN ANALYZE on 22,500 games: 42.8 ms (parallel seq scan, top-N sort) to 0.12 ms.
// One concurrent statement per NoTransaction migration.
func m20261107000000AddItemsListScoreIdx() schema.Migration {
	m := postgresMigration("20261107000000_add_items_list_score_idx",
		[]string{
			`CREATE INDEX CONCURRENTLY IF NOT EXISTS items_list_score_idx ON items
  (type, (coalesce(CASE WHEN (rating_count >= 5 AND score IS NOT NULL) THEN score END, 1000)), id)
  WHERE visibility = 'public' AND coalesce(cover_key, '') <> ''`,
		},
		[]string{`DROP INDEX CONCURRENTLY IF EXISTS items_list_score_idx`},
	)
	m.NoTransaction = true

	return m
}

// m20261107000100AddItemsListCoveredIdx makes the listing's total count an index-only scan: 39.2 ms to 1.3 ms on
// 22,500 games, 0 heap fetches after vacuum.
func m20261107000100AddItemsListCoveredIdx() schema.Migration {
	m := postgresMigration("20261107000100_add_items_list_covered_idx",
		[]string{
			`CREATE INDEX CONCURRENTLY IF NOT EXISTS items_list_covered_idx ON items (type)
  WHERE visibility = 'public' AND coalesce(cover_key, '') <> ''`,
		},
		[]string{`DROP INDEX CONCURRENTLY IF EXISTS items_list_covered_idx`},
	)
	m.NoTransaction = true

	return m
}
