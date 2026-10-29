package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029100008CreateItemLinksTable is SDD 01 §5.2 `item_links`: item to item relations (open issue 08-1), filled by SDD 06.
func m20261029100008CreateItemLinksTable() schema.Migration {
	return schema.Migration{
		ID: "20261029100008_create_item_links_table",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE TABLE item_links (
  from_item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  to_item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  relation text NOT NULL CHECK (relation IN ('based_on')),
  source text NOT NULL REFERENCES data_sources(code),
  PRIMARY KEY (from_item_id, relation, to_item_id),
  CHECK (from_item_id <> to_item_id)
);
CREATE INDEX item_links_to_idx ON item_links (to_item_id, relation)`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DROP TABLE IF EXISTS item_links`)
		},
	}
}
