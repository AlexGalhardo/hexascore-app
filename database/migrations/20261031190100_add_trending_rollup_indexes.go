package migrations

import "github.com/AlexGalhardo/vukem/schema"

// trendingRollupIndex builds one single-statement migration: CREATE INDEX CONCURRENTLY cannot run inside a transaction
// and a half-applied multi-statement NoTransaction migration would be unrecoverable (vukem schema.Migration.NoTransaction).
// The hourly trending rollup (app/recommend/trending/rollup.go) filters each of these event-time columns by UTC day, and none led an index.
func trendingRollupIndex(id, name, table, column string) schema.Migration {
	m := postgresMigration(id,
		[]string{`CREATE INDEX CONCURRENTLY IF NOT EXISTS ` + name + ` ON ` + table + ` (` + column + `)`},
		[]string{`DROP INDEX CONCURRENTLY IF EXISTS ` + name},
	)
	m.NoTransaction = true

	return m
}

func m20261031190100AddRatingEventsAtIndex() schema.Migration {
	return trendingRollupIndex(
		"20261031190100_add_rating_events_at_index",
		"rating_events_at_idx",
		"rating_events",
		"at",
	)
}

func m20261031190200AddItemStatusUpdatedAtIndex() schema.Migration {
	return trendingRollupIndex(
		"20261031190200_add_item_status_updated_at_index",
		"item_status_updated_at_idx",
		"item_status",
		"updated_at",
	)
}

func m20261031190300AddListItemsAddedAtIndex() schema.Migration {
	return trendingRollupIndex(
		"20261031190300_add_list_items_added_at_index",
		"list_items_added_at_idx",
		"list_items",
		"added_at",
	)
}

func m20261031190400AddSuggestionEventsCreatedAtIndex() schema.Migration {
	return trendingRollupIndex(
		"20261031190400_add_suggestion_events_created_at_index",
		"suggestion_events_created_at_idx",
		"suggestion_events",
		"created_at",
	)
}
