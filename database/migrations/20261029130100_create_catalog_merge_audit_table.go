package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029130100CreateCatalogMergeAuditTable holds the jsonb snapshot of both
// items taken by the merge service (SDD 01 §7.5, open issue 04-9). It is an
// interim append-only table: SDD 04's catalog_audit (P4.A.1.1) does not exist
// yet; its migration folds these rows in as `source=merge` events.
func m20261029130100CreateCatalogMergeAuditTable() schema.Migration {
	return schema.Migration{
		ID: "20261029130100_create_catalog_merge_audit_table",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE TABLE catalog_merge_audit (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  occurred_at timestamptz NOT NULL DEFAULT now(),
  actor_id bigint REFERENCES users(id) ON DELETE SET NULL,
  item_type text NOT NULL CHECK (item_type IN ('game','movie','series','book')),
  source_item_id uuid NOT NULL,
  target_item_id uuid NOT NULL,
  snapshot jsonb NOT NULL
);
CREATE INDEX catalog_merge_audit_target_idx ON catalog_merge_audit (target_item_id, occurred_at DESC);
CREATE FUNCTION catalog_merge_audit_append_only() RETURNS trigger AS $$
BEGIN RAISE EXCEPTION 'catalog_merge_audit is append-only'; END $$ LANGUAGE plpgsql;
CREATE TRIGGER catalog_merge_audit_append_only BEFORE UPDATE OR DELETE ON catalog_merge_audit
  FOR EACH ROW EXECUTE FUNCTION catalog_merge_audit_append_only()`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DROP TABLE IF EXISTS catalog_merge_audit;
DROP FUNCTION IF EXISTS catalog_merge_audit_append_only()`)
		},
	}
}
