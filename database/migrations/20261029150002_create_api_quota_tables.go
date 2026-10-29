package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029150002CreateAPIQuotaTables is SDD 05 §5 metering with users.id as bigint: legacy credit balances
// (D-23), counters per period ('life' or 'YYYYMM', ADR 0003) and daily usage kept 400 days.
func m20261029150002CreateAPIQuotaTables() schema.Migration {
	return postgresMigration("20261029150002_create_api_quota_tables",
		[]string{
			`CREATE TABLE api_credit_balances (
  user_id bigint PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  remaining integer NOT NULL CHECK (remaining >= 0),
  migrated_at timestamptz NOT NULL DEFAULT now()
)`,
			`CREATE TABLE api_quota_counters (
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  period text NOT NULL CHECK (period = 'life' OR period ~ '^[0-9]{6}$'),
  used integer NOT NULL CHECK (used >= 0),
  PRIMARY KEY (user_id, period)
)`,
			`CREATE TABLE api_usage_daily (
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  day date NOT NULL,
  count integer NOT NULL CHECK (count >= 0),
  refunded integer NOT NULL DEFAULT 0 CHECK (refunded >= 0),
  PRIMARY KEY (user_id, day)
)`,
		},
		[]string{
			`DROP TABLE IF EXISTS api_usage_daily`,
			`DROP TABLE IF EXISTS api_quota_counters`,
			`DROP TABLE IF EXISTS api_credit_balances`,
		},
	)
}
