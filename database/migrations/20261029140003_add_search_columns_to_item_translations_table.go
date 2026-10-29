package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029140003AddSearchColumnsToItemTranslationsTable is SDD 07 §5 and open issue 07-1: the tsvectors are trigger-maintained plain columns (a CASE over regconfig may fail the generated-column immutability check). `search_tsv` serves the site (title A, any synopsis D); `search_tsv_api` holds allowlisted-source text only (07-7). Adding a synopsis source to CATALOG_API_SOURCES means a migration that edits the CASE below and runs search:rebuild.
func m20261029140003AddSearchColumnsToItemTranslationsTable() schema.Migration {
	return postgresMigration("20261029140003_add_search_columns_to_item_translations_table",
		[]string{
			`ALTER TABLE item_translations
  ADD COLUMN title_norm text GENERATED ALWAYS AS (hs_norm(title)) STORED,
  ADD COLUMN search_tsv tsvector NOT NULL DEFAULT '',
  ADD COLUMN search_tsv_api tsvector NOT NULL DEFAULT ''`,
			`CREATE FUNCTION item_translations_tsv() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE cfg regconfig := CASE new.locale WHEN 'en' THEN 'hs_en' WHEN 'pt-BR' THEN 'hs_pt' ELSE 'hs_es' END;
BEGIN
  new.search_tsv := setweight(to_tsvector(cfg, new.title), 'A')
    || setweight(to_tsvector(cfg, coalesce(new.synopsis, '')), 'D');
  new.search_tsv_api := setweight(to_tsvector(cfg, new.title), 'A')
    || setweight(to_tsvector(cfg, CASE WHEN new.synopsis_source = 'wikipedia' THEN coalesce(new.synopsis, '') ELSE '' END), 'D');
  RETURN new;
END $$`,
			`CREATE TRIGGER item_translations_tsv_trg BEFORE INSERT OR UPDATE OF title, synopsis, synopsis_source, locale
  ON item_translations FOR EACH ROW EXECUTE FUNCTION item_translations_tsv()`,
		},
		[]string{
			`DROP TRIGGER IF EXISTS item_translations_tsv_trg ON item_translations`,
			`DROP FUNCTION IF EXISTS item_translations_tsv()`,
			`ALTER TABLE item_translations DROP COLUMN IF EXISTS search_tsv_api, DROP COLUMN IF EXISTS search_tsv, DROP COLUMN IF EXISTS title_norm`,
		},
	)
}
