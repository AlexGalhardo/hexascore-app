package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029100005CreateEntityTranslationsTable is SDD 01 §5.2 `entity_translations`: localised names and descriptions with their source and licence.
func m20261029100005CreateEntityTranslationsTable() schema.Migration {
	return schema.Migration{
		ID: "20261029100005_create_entity_translations_table",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE TABLE entity_translations (
  entity_id uuid NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
  locale text NOT NULL CHECK (locale IN ('en','pt-BR','es')),
  name text NOT NULL,
  name_source text NOT NULL REFERENCES data_sources(code),
  description text,
  description_source text REFERENCES data_sources(code),
  description_licence text,
  description_source_url text,
  PRIMARY KEY (entity_id, locale)
)`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DROP TABLE IF EXISTS entity_translations`)
		},
	}
}
