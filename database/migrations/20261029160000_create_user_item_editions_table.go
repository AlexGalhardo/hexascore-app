package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029160000CreateUserItemEditionsTable is SDD 14 §5: the optional book edition a user picked for a work. The composite foreign key (edition_id, item_id) -> editions(id, work_id) enforces "the edition belongs to the work" in the database; ON UPDATE CASCADE lets a catalog merge repoint editions (and with them these rows) at the target work.
func m20261029160000CreateUserItemEditionsTable() schema.Migration {
	return postgresMigration("20261029160000_create_user_item_editions_table",
		[]string{
			`CREATE TABLE user_item_editions (
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  edition_id uuid NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, item_id),
  FOREIGN KEY (edition_id, item_id) REFERENCES editions(id, work_id) ON DELETE CASCADE ON UPDATE CASCADE
)`,
		},
		[]string{`DROP TABLE IF EXISTS user_item_editions`},
	)
}
