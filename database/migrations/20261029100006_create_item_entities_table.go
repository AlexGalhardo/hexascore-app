package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029100006CreateItemEntitiesTable is SDD 01 §5.2 `item_entities`: the role-typed item to entity relation. Role and kind compatibility is enforced in Go (app/catalog/roles.go); engines and edition contributors have no role values (D-30).
func m20261029100006CreateItemEntitiesTable() schema.Migration {
	return schema.Migration{
		ID: "20261029100006_create_item_entities_table",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE TABLE item_entities (
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  entity_id uuid NOT NULL REFERENCES entities(id) ON DELETE RESTRICT,
  role text NOT NULL CHECK (role IN ('director','writer','cast','creator','author',
    'developer','publisher','production_company','network','platform','genre',
    'game_mode','franchise','series','country','original_language')),
  position smallint NOT NULL DEFAULT 0,
  character_name text,
  source text NOT NULL REFERENCES data_sources(code),
  PRIMARY KEY (item_id, role, entity_id)
);
CREATE INDEX item_entities_entity_idx ON item_entities (entity_id, role, item_id)`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DROP TABLE IF EXISTS item_entities`)
		},
	}
}
