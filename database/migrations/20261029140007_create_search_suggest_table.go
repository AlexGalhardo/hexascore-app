package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029140007CreateSearchSuggestTable is SDD 07 §5: one row per suggestable label (a translated title, an
// alternative title or edition title as locale '*', an entity name) with its normalised form for the SP-GiST prefix
// and trigram indexes. The sync runner (internal/search) keeps it equal to the source tables from search_dirty.
func m20261029140007CreateSearchSuggestTable() schema.Migration {
	return postgresMigration("20261029140007_create_search_suggest_table",
		[]string{
			`CREATE TABLE search_suggest (
  ref_id uuid NOT NULL,
  kind text NOT NULL CHECK (kind IN ('game','movie','series','book','person','organization','franchise')),
  locale text NOT NULL CHECK (locale IN ('en','pt-BR','es','*')),
  label text NOT NULL CHECK (length(label) <= 300),
  label_norm text NOT NULL,
  slug text NOT NULL,
  year smallint,
  thumb_key text,
  popularity real NOT NULL DEFAULT 0,
  PRIMARY KEY (ref_id, locale, label_norm)
)`,
			`CREATE INDEX search_suggest_prefix ON search_suggest USING spgist (label_norm)`,
			`CREATE INDEX search_suggest_trgm ON search_suggest USING gin (label_norm gin_trgm_ops)`,
			`ALTER TABLE search_suggest SET (autovacuum_vacuum_scale_factor = 0.05)`,
			// The five SDD 07 source tables miss changes on items itself: visibility, slug, year and thumbnail are suggest columns.
			`CREATE TRIGGER items_search_dirty_trg AFTER UPDATE OF visibility, slug, release_date, cover_thumb_key ON items
  FOR EACH ROW EXECUTE FUNCTION search_dirty_enqueue('item', 'id')`,
		},
		[]string{`DROP TRIGGER IF EXISTS items_search_dirty_trg ON items`, `DROP TABLE IF EXISTS search_suggest`},
	)
}
