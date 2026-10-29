package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029140009CreateListsTable is SDD 14 §5 with users.id as bigint (the SDD's text is superseded by SDD 03): a user's ordered list of one item type; `hidden` is the moderation flag and `version` guards reorders.
func m20261029140009CreateListsTable() schema.Migration {
	return postgresMigration("20261029140009_create_lists_table",
		[]string{
			`CREATE TABLE lists (
  id uuid PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type text NOT NULL CHECK (type IN ('game','movie','series','book')),
  title text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 80),
  description text CHECK (char_length(description) <= 500),
  is_public boolean NOT NULL DEFAULT true,
  hidden boolean NOT NULL DEFAULT false,
  item_count int NOT NULL DEFAULT 0 CHECK (item_count BETWEEN 0 AND 500),
  version int NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
)`,
			`CREATE INDEX lists_user_type_idx ON lists (user_id, type);
CREATE INDEX lists_public_idx ON lists (updated_at DESC) WHERE is_public AND NOT hidden`,
			`CREATE FUNCTION lists_type_immutable() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF NEW.type IS DISTINCT FROM OLD.type THEN
    RAISE EXCEPTION 'lists.type is immutable';
  END IF;
  RETURN NEW;
END $$`,
			`CREATE TRIGGER lists_type_immutable_trg BEFORE UPDATE OF type ON lists
FOR EACH ROW EXECUTE FUNCTION lists_type_immutable()`,
		},
		[]string{`DROP TABLE IF EXISTS lists CASCADE`, `DROP FUNCTION IF EXISTS lists_type_immutable()`},
	)
}
