package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029140006CreateSearchDirtyTable is SDD 07 §5: queue of item/entity ids whose suggest rows need a resync. AFTER triggers on the five source tables enqueue the owning item or entity id (editions enqueue their work); the sync runner (P1.E.1.2) resolves the concrete kind and drains the queue. kind is 'item' or 'entity'.
func m20261029140006CreateSearchDirtyTable() schema.Migration {
	return postgresMigration("20261029140006_create_search_dirty_table",
		[]string{
			`CREATE TABLE search_dirty (
  ref_id uuid PRIMARY KEY,
  kind text NOT NULL CHECK (kind IN ('item','entity')),
  queued_at timestamptz NOT NULL DEFAULT now()
)`,
			`ALTER TABLE search_dirty SET (autovacuum_vacuum_scale_factor = 0.05)`,
			`CREATE FUNCTION search_dirty_enqueue() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  INSERT INTO search_dirty (ref_id, kind)
  VALUES ((to_jsonb(coalesce(new, old)) ->> TG_ARGV[1])::uuid, TG_ARGV[0])
  ON CONFLICT (ref_id) DO UPDATE SET queued_at = now();
  RETURN NULL;
END $$`,
			`CREATE TRIGGER item_translations_search_dirty_trg AFTER INSERT OR UPDATE OR DELETE ON item_translations
  FOR EACH ROW EXECUTE FUNCTION search_dirty_enqueue('item', 'item_id')`,
			`CREATE TRIGGER item_alt_titles_search_dirty_trg AFTER INSERT OR UPDATE OR DELETE ON item_alt_titles
  FOR EACH ROW EXECUTE FUNCTION search_dirty_enqueue('item', 'item_id')`,
			`CREATE TRIGGER editions_search_dirty_trg AFTER INSERT OR UPDATE OR DELETE ON editions
  FOR EACH ROW EXECUTE FUNCTION search_dirty_enqueue('item', 'work_id')`,
			`CREATE TRIGGER entity_translations_search_dirty_trg AFTER INSERT OR UPDATE OR DELETE ON entity_translations
  FOR EACH ROW EXECUTE FUNCTION search_dirty_enqueue('entity', 'entity_id')`,
			`CREATE TRIGGER entities_search_dirty_trg AFTER INSERT OR UPDATE OR DELETE ON entities
  FOR EACH ROW EXECUTE FUNCTION search_dirty_enqueue('entity', 'id')`,
		},
		[]string{
			`DROP TRIGGER IF EXISTS item_translations_search_dirty_trg ON item_translations`,
			`DROP TRIGGER IF EXISTS item_alt_titles_search_dirty_trg ON item_alt_titles`,
			`DROP TRIGGER IF EXISTS editions_search_dirty_trg ON editions`,
			`DROP TRIGGER IF EXISTS entity_translations_search_dirty_trg ON entity_translations`,
			`DROP TRIGGER IF EXISTS entities_search_dirty_trg ON entities`,
			`DROP FUNCTION IF EXISTS search_dirty_enqueue()`,
			`DROP TABLE IF EXISTS search_dirty`,
		},
	)
}
