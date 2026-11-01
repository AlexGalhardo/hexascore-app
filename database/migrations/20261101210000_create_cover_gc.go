package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261101210000CreateCoverGC adds the deferred delete of cover objects (ADR 0014 as amended, SDD 06 §7.5): a
// trigger on items and editions files every covers/ key a row stops pointing at, whatever wrote the row (covers
// job, tmdb.expire, igdb.refresh, rollback, a delete), and covers.gc deletes the object 30 days later unless a row
// points at it again.
func m20261101210000CreateCoverGC() schema.Migration {
	return postgresMigration("20261101210000_create_cover_gc",
		[]string{
			`CREATE TABLE cover_gc (
  key text PRIMARY KEY,
  released_at timestamptz NOT NULL DEFAULT now()
)`,
			`CREATE INDEX cover_gc_released_at_idx ON cover_gc (released_at)`,
			`CREATE FUNCTION cover_gc_release() RETURNS trigger AS $$
DECLARE
  keep text[] := '{}';
BEGIN
  IF TG_OP = 'UPDATE' THEN
    keep := array_remove(ARRAY[NEW.cover_key, NEW.cover_thumb_key], NULL);
  END IF;
  INSERT INTO cover_gc (key)
  SELECT k FROM unnest(ARRAY[OLD.cover_key, OLD.cover_thumb_key]) AS k
  WHERE k LIKE 'covers/%' AND k <> ALL (keep)
  ON CONFLICT (key) DO UPDATE SET released_at = excluded.released_at;
  RETURN NULL;
END $$ LANGUAGE plpgsql`,
			`CREATE TRIGGER items_cover_gc AFTER UPDATE OF cover_key, cover_thumb_key OR DELETE ON items
  FOR EACH ROW EXECUTE FUNCTION cover_gc_release()`,
			`CREATE TRIGGER editions_cover_gc AFTER UPDATE OF cover_key, cover_thumb_key OR DELETE ON editions
  FOR EACH ROW EXECUTE FUNCTION cover_gc_release()`,
		},
		[]string{
			`DROP TRIGGER IF EXISTS editions_cover_gc ON editions`,
			`DROP TRIGGER IF EXISTS items_cover_gc ON items`,
			`DROP FUNCTION IF EXISTS cover_gc_release()`,
			`DROP TABLE IF EXISTS cover_gc`,
		},
	)
}
