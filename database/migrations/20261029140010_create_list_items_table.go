package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029140010CreateListItemsTable is SDD 14 §5: positions are dense 1..n after a reorder (gaps allowed after removals) and the (list_id, position) uniqueness is deferred so a reorder can renumber in one statement.
func m20261029140010CreateListItemsTable() schema.Migration {
	return postgresMigration("20261029140010_create_list_items_table",
		[]string{
			`CREATE TABLE list_items (
  list_id uuid NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  position int NOT NULL CHECK (position > 0),
  note text CHECK (char_length(note) <= 280),
  added_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (list_id, item_id),
  UNIQUE (list_id, position) DEFERRABLE INITIALLY DEFERRED
)`,
			`CREATE INDEX list_items_item_idx ON list_items (item_id)`,
		},
		[]string{`DROP TABLE IF EXISTS list_items`},
	)
}
