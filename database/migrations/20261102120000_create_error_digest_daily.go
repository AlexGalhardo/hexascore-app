package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261102120000CreateErrorDigestDaily stores the error and alert-line counts per fingerprint and hour that the
// 08:00 Telegram digest sums over 24 h (SDD 12 §7.2, FR-12-6); rows are pruned after 90 days. message is the scrubbed
// log message, request_id one sample.
func m20261102120000CreateErrorDigestDaily() schema.Migration {
	return postgresMigration("20261102120000_create_error_digest_daily",
		[]string{
			`CREATE TABLE error_digest_daily (
  hour timestamptz NOT NULL,
  fingerprint text NOT NULL,
  message text NOT NULL,
  alert text NOT NULL DEFAULT '',
  route text NOT NULL DEFAULT '',
  request_id text NOT NULL DEFAULT '',
  count bigint NOT NULL,
  PRIMARY KEY (hour, fingerprint)
)`,
		},
		[]string{`DROP TABLE IF EXISTS error_digest_daily`},
	)
}
