package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261031173100CreateCatalogChangedDone records the catalog_audit ids whose `catalog.changed` side effects ran,
// so a retried or duplicated job does them once (SDD 04 section 6 "Jobs": idempotent by audit_ids).
func m20261031173100CreateCatalogChangedDone() schema.Migration {
	return postgresMigration("20261031173100_create_catalog_changed_done",
		[]string{
			`CREATE TABLE catalog_changed_done (
  audit_id bigint PRIMARY KEY REFERENCES catalog_audit(id),
  done_at timestamptz NOT NULL DEFAULT now()
)`,
		},
		[]string{`DROP TABLE IF EXISTS catalog_changed_done`},
	)
}
