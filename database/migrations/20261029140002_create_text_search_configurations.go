package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029140002CreateTextSearchConfigurations is SDD 07 §5: per-locale configs that unaccent before stemming. `asciiword` keeps the stock mapping. A changed mapping ships as a new config name (hs_pt_v2): altering one does not rewrite stored vectors.
func m20261029140002CreateTextSearchConfigurations() schema.Migration {
	return postgresMigration("20261029140002_create_text_search_configurations",
		[]string{
			`CREATE TEXT SEARCH CONFIGURATION hs_en (COPY = pg_catalog.english)`,
			`ALTER TEXT SEARCH CONFIGURATION hs_en ALTER MAPPING FOR hword, hword_part, word WITH unaccent, english_stem`,
			`CREATE TEXT SEARCH CONFIGURATION hs_pt (COPY = pg_catalog.portuguese)`,
			`ALTER TEXT SEARCH CONFIGURATION hs_pt ALTER MAPPING FOR hword, hword_part, word WITH unaccent, portuguese_stem`,
			`CREATE TEXT SEARCH CONFIGURATION hs_es (COPY = pg_catalog.spanish)`,
			`ALTER TEXT SEARCH CONFIGURATION hs_es ALTER MAPPING FOR hword, hword_part, word WITH unaccent, spanish_stem`,
		},
		[]string{
			`DROP TEXT SEARCH CONFIGURATION IF EXISTS hs_es`,
			`DROP TEXT SEARCH CONFIGURATION IF EXISTS hs_pt`,
			`DROP TEXT SEARCH CONFIGURATION IF EXISTS hs_en`,
		},
	)
}
