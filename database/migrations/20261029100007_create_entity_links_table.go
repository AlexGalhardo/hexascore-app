package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029100007CreateEntityLinksTable is SDD 01 §5.2 `entity_links`: platform to holder, imprint to publisher, franchise to parent.
func m20261029100007CreateEntityLinksTable() schema.Migration {
	return schema.Migration{
		ID: "20261029100007_create_entity_links_table",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE TABLE entity_links (
  from_id uuid NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
  to_id uuid NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
  relation text NOT NULL CHECK (relation IN ('platform_holder','imprint_of','part_of')),
  PRIMARY KEY (from_id, relation, to_id)
)`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DROP TABLE IF EXISTS entity_links`)
		},
	}
}
