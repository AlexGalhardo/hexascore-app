package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261030140000CreateWebhookEndpoints is SDD 05 §5. user_id is bigint because
// users.id is bigint in this app (the SDD text column is the legacy shape).
// webhook_cursor starts at 0 so the fan-out job has a row to advance.
func m20261030140000CreateWebhookEndpoints() schema.Migration {
	return postgresMigration("20261030140000_create_webhook_endpoints",
		[]string{
			`CREATE TABLE webhook_endpoints (
  id uuid PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  url text NOT NULL CHECK (url LIKE 'https://%'),
  events text[] NOT NULL CHECK (cardinality(events) BETWEEN 1 AND 3),
  item_types text[] NOT NULL CHECK (cardinality(item_types) BETWEEN 1 AND 4),
  secret_enc text NOT NULL,
  disabled_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
)`,
			`CREATE INDEX webhook_endpoints_user_idx ON webhook_endpoints (user_id)`,
			`CREATE TABLE webhook_cursor (
  id smallint PRIMARY KEY CHECK (id = 1),
  last_change_id bigint NOT NULL
)`,
			`INSERT INTO webhook_cursor (id, last_change_id) VALUES (1, 0)`,
		},
		[]string{
			`DROP TABLE IF EXISTS webhook_cursor`,
			`DROP TABLE IF EXISTS webhook_endpoints`,
		},
	)
}
