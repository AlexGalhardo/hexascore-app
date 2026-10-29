package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029100002CreateItemTranslationsTable: per-locale title and synopsis
// with their sources; a synopsis always carries its source and licence.
func m20261029100002CreateItemTranslationsTable() schema.Migration {
	return schema.Migration{
		ID: "20261029100002_create_item_translations_table",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE TABLE item_translations (
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  locale text NOT NULL CHECK (locale IN ('en','pt-BR','es')),
  title text NOT NULL CHECK (length(title) BETWEEN 1 AND 500),
  title_source text NOT NULL REFERENCES data_sources(code),
  synopsis text,
  synopsis_source text REFERENCES data_sources(code),
  synopsis_licence text,
  synopsis_source_url text,
  PRIMARY KEY (item_id, locale),
  CHECK (synopsis IS NULL OR (synopsis_source IS NOT NULL AND synopsis_licence IS NOT NULL))
)`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DROP TABLE IF EXISTS item_translations`)
		},
	}
}
