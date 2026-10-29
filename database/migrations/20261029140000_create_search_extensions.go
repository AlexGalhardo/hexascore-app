package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029140000CreateSearchExtensions is SDD 07 §5: pg_trgm for fuzzy title matching and unaccent for accent-insensitive search (PostgreSQL 15+).
func m20261029140000CreateSearchExtensions() schema.Migration {
	return postgresMigration("20261029140000_create_search_extensions",
		[]string{
			`CREATE EXTENSION IF NOT EXISTS pg_trgm`,
			`CREATE EXTENSION IF NOT EXISTS unaccent`,
		},
		[]string{
			`DROP EXTENSION IF EXISTS unaccent`,
			`DROP EXTENSION IF EXISTS pg_trgm`,
		},
	)
}
