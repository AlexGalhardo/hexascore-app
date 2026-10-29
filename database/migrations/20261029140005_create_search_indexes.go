package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029140005CreateSearchIndexes is SDD 07 §5: GIN on both tsvectors and trigram GIN on every normalised title or name. Plain CREATE INDEX inside the migration transaction: the tables are near-empty before cutover (VF-017); later rebuilds use the concurrent pre-deploy script. gin_pending_list_limit is 4MB on trigram indexes (§7.10).
func m20261029140005CreateSearchIndexes() schema.Migration {
	return postgresMigration("20261029140005_create_search_indexes",
		[]string{
			`CREATE INDEX item_translations_tsv_gin ON item_translations USING gin (search_tsv)`,
			`CREATE INDEX item_translations_tsv_api_gin ON item_translations USING gin (search_tsv_api)`,
			`CREATE INDEX item_translations_title_trgm ON item_translations USING gin (title_norm gin_trgm_ops) WITH (gin_pending_list_limit = 4096)`,
			`CREATE UNIQUE INDEX item_alt_titles_norm_uq ON item_alt_titles (item_id, title_norm)`,
			`CREATE INDEX item_alt_titles_trgm ON item_alt_titles USING gin (title_norm gin_trgm_ops) WITH (gin_pending_list_limit = 4096)`,
			`CREATE INDEX editions_title_trgm ON editions USING gin (title_norm gin_trgm_ops) WITH (gin_pending_list_limit = 4096)`,
			`CREATE INDEX entity_translations_name_trgm ON entity_translations USING gin (name_norm gin_trgm_ops) WITH (gin_pending_list_limit = 4096)`,
			`CREATE INDEX entities_name_trgm ON entities USING gin (name_norm gin_trgm_ops) WITH (gin_pending_list_limit = 4096)
  WHERE kind IN ('person','organization','franchise')`,
			`ALTER TABLE item_translations SET (autovacuum_vacuum_scale_factor = 0.05)`,
		},
		[]string{
			`ALTER TABLE item_translations RESET (autovacuum_vacuum_scale_factor)`,
			`DROP INDEX IF EXISTS entities_name_trgm`,
			`DROP INDEX IF EXISTS entity_translations_name_trgm`,
			`DROP INDEX IF EXISTS editions_title_trgm`,
			`DROP INDEX IF EXISTS item_alt_titles_trgm`,
			`DROP INDEX IF EXISTS item_alt_titles_norm_uq`,
			`DROP INDEX IF EXISTS item_translations_title_trgm`,
			`DROP INDEX IF EXISTS item_translations_tsv_api_gin`,
			`DROP INDEX IF EXISTS item_translations_tsv_gin`,
		},
	)
}
