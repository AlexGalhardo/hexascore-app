package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029150001CreateAPIKeysTable is SDD 05 §5 with users.id as bigint (SDD 03). Own table, not vukem access_tokens:
// legacy keys are sha256(full key) and migrate unchanged (A-04), so the hash is the lookup key.
func m20261029150001CreateAPIKeysTable() schema.Migration {
	return postgresMigration("20261029150001_create_api_keys_table",
		[]string{
			`CREATE TABLE api_keys (
  id uuid PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 40),
  prefix text NOT NULL,
  hash char(64) NOT NULL UNIQUE,
  last_used_at timestamptz,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
)`,
			`CREATE INDEX api_keys_user_live_idx ON api_keys (user_id) WHERE revoked_at IS NULL`,
		},
		[]string{`DROP TABLE IF EXISTS api_keys`},
	)
}
