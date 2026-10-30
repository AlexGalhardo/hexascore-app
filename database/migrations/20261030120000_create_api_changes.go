package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261030120000CreateAPIChanges is SDD 05 §5. item_id is uuid, matching the catalogue primary key.
func m20261030120000CreateAPIChanges() schema.Migration {
	return postgresMigration("20261030120000_create_api_changes",
		[]string{
			`CREATE TABLE api_changes (
  id bigserial PRIMARY KEY,
  item_id uuid NOT NULL,
  item_type text NOT NULL CHECK (item_type IN ('game','movie','series','book')),
  action text NOT NULL CHECK (action IN ('created','updated','deleted')),
  fields text[] NOT NULL DEFAULT '{}',
  merged_into uuid,
  changed_at timestamptz NOT NULL DEFAULT now()
)`,
			`CREATE INDEX api_changes_changed_at_id_idx ON api_changes (changed_at, id)`,
		},
		[]string{`DROP TABLE IF EXISTS api_changes`},
	)
}
