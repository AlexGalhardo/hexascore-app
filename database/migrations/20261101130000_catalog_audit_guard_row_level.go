package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261101130000CatalogAuditGuardRowLevel makes the append-only guard a row trigger. The statement trigger
// also fired for the ON DELETE SET NULL foreign key on actor_id with zero rows, so every hard DELETE FROM users
// failed. The one write the guard lets through is that foreign key nulling actor_id (account erasure); every
// other UPDATE and every DELETE still raises.
func m20261101130000CatalogAuditGuardRowLevel() schema.Migration {
	return postgresMigration("20261101130000_catalog_audit_guard_row_level",
		[]string{
			`DROP TRIGGER catalog_audit_append_only ON catalog_audit`,
			`CREATE OR REPLACE FUNCTION catalog_audit_append_only() RETURNS trigger AS $$
BEGIN
  IF TG_OP = 'UPDATE' AND NEW.actor_id IS NULL AND OLD.actor_id IS NOT NULL
     AND to_jsonb(NEW) - 'actor_id' = to_jsonb(OLD) - 'actor_id' THEN
    RETURN NEW;
  END IF;
  RAISE EXCEPTION 'catalog_audit is append-only';
END $$ LANGUAGE plpgsql`,
			`CREATE TRIGGER catalog_audit_append_only
  BEFORE UPDATE OR DELETE ON catalog_audit
  FOR EACH ROW EXECUTE FUNCTION catalog_audit_append_only()`,
		},
		[]string{
			`DROP TRIGGER catalog_audit_append_only ON catalog_audit`,
			`CREATE OR REPLACE FUNCTION catalog_audit_append_only() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'catalog_audit is append-only';
END $$ LANGUAGE plpgsql`,
			`CREATE TRIGGER catalog_audit_append_only
  BEFORE UPDATE OR DELETE ON catalog_audit
  FOR EACH STATEMENT EXECUTE FUNCTION catalog_audit_append_only()`,
		},
	)
}
