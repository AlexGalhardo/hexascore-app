package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029140011CreateUserNotificationsTable is SDD 14 §5 with users.id as bigint (the SDD's text is superseded by SDD 03): the in-app inbox, written only by notifications.Send; (user_id, dedupe_key) makes a retried job insert once.
func m20261029140011CreateUserNotificationsTable() schema.Migration {
	return postgresMigration("20261029140011_create_user_notifications_table",
		[]string{
			`CREATE TABLE user_notifications (
  id uuid PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  kind text NOT NULL CHECK (kind IN ('suggestion_decided','suggestion_reverted','suggestion_expired',
    'report_resolved','rating_flagged','list_hidden','system')),
  payload jsonb NOT NULL,
  dedupe_key text NOT NULL CHECK (char_length(dedupe_key) BETWEEN 1 AND 200),
  read_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, dedupe_key)
)`,
			`CREATE INDEX user_notifications_unread ON user_notifications (user_id, created_at DESC) WHERE read_at IS NULL;
CREATE INDEX user_notifications_user ON user_notifications (user_id, id DESC)`,
		},
		[]string{`DROP TABLE IF EXISTS user_notifications`},
	)
}
