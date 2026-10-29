package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029140004AddNormColumnsToSearchTables is SDD 07 §5: normalised title and name columns on the other search sources. They stay generated because hs_norm is an IMMUTABLE wrapper.
func m20261029140004AddNormColumnsToSearchTables() schema.Migration {
	return postgresMigration("20261029140004_add_norm_columns_to_search_tables",
		[]string{
			`ALTER TABLE item_alt_titles ADD COLUMN title_norm text GENERATED ALWAYS AS (hs_norm(title)) STORED`,
			`ALTER TABLE editions ADD COLUMN title_norm text GENERATED ALWAYS AS (hs_norm(title)) STORED`,
			`ALTER TABLE entity_translations ADD COLUMN name_norm text GENERATED ALWAYS AS (hs_norm(name)) STORED`,
			`ALTER TABLE entities ADD COLUMN name_norm text GENERATED ALWAYS AS (hs_norm(default_name)) STORED`,
		},
		[]string{
			`ALTER TABLE entities DROP COLUMN IF EXISTS name_norm`,
			`ALTER TABLE entity_translations DROP COLUMN IF EXISTS name_norm`,
			`ALTER TABLE editions DROP COLUMN IF EXISTS title_norm`,
			`ALTER TABLE item_alt_titles DROP COLUMN IF EXISTS title_norm`,
		},
	)
}
