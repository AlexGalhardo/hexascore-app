package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261101220000AddRollbackGapFixes closes two ingestion gaps. source_record_history keeps the checked_at (the TMDB
// expiry clock, SDD 06 §7.4) of the version it archives, so a rollback restores it instead of the bad run's. A
// trigger on items.visibility feeds /changes (SDD 05 §7.3): leaving public is `deleted` and coming back is `created`,
// because the paid API answers 404 for a non-public item, whatever wrote the column (rollback, an admin, SQL).
func m20261101220000AddRollbackGapFixes() schema.Migration {
	return postgresMigration("20261101220000_add_rollback_gap_fixes",
		[]string{
			`ALTER TABLE source_record_history ADD COLUMN checked_at timestamptz`,
			`CREATE FUNCTION items_visibility_changes() RETURNS trigger AS $$
BEGIN
  IF NEW.visibility = 'public' THEN
    INSERT INTO api_changes (item_id, item_type, action) VALUES (NEW.id, NEW.type, 'created');
  ELSIF OLD.visibility = 'public' THEN
    INSERT INTO api_changes (item_id, item_type, action) VALUES (NEW.id, NEW.type, 'deleted');
  END IF;
  RETURN NULL;
END $$ LANGUAGE plpgsql`,
			`CREATE TRIGGER items_visibility_changes AFTER UPDATE OF visibility ON items
  FOR EACH ROW WHEN (OLD.visibility IS DISTINCT FROM NEW.visibility)
  EXECUTE FUNCTION items_visibility_changes()`,
		},
		[]string{
			`DROP TRIGGER IF EXISTS items_visibility_changes ON items`,
			`DROP FUNCTION IF EXISTS items_visibility_changes()`,
			`ALTER TABLE source_record_history DROP COLUMN IF EXISTS checked_at`,
		},
	)
}
