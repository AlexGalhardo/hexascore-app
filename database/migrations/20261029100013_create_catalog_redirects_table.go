package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029100013CreateCatalogRedirectsTable is SDD 01 §5.2 `catalog_redirects`: old slugs and ids of renamed or merged items and entities. `target_id` is polymorphic by `scope`, so I-2 (live target, flattened chains) is enforced by the merge code, not a foreign key.
func m20261029100013CreateCatalogRedirectsTable() schema.Migration {
	return schema.Migration{
		ID: "20261029100013_create_catalog_redirects_table",
		Up: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil // scaffold tests run on SQLite (VF-033)
			}
			return s.Raw(`CREATE TABLE catalog_redirects (
  scope text NOT NULL CHECK (scope IN ('item','entity')),
  namespace text NOT NULL,
  old_slug text,
  old_id uuid,
  target_id uuid NOT NULL,
  reason text NOT NULL CHECK (reason IN ('rename','merge')),
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (old_slug IS NOT NULL OR old_id IS NOT NULL)
);
CREATE UNIQUE INDEX catalog_redirects_slug_uq ON catalog_redirects (scope, namespace, old_slug)
  WHERE old_slug IS NOT NULL;
CREATE UNIQUE INDEX catalog_redirects_id_uq ON catalog_redirects (scope, old_id)
  WHERE old_id IS NOT NULL`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() != "postgres" {
				return nil
			}
			return s.Raw(`DROP TABLE IF EXISTS catalog_redirects`)
		},
	}
}
