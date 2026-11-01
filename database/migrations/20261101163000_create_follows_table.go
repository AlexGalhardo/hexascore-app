package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261101163000CreateFollowsTable adds the follow graph and its notification kinds (SDD 14 §12.4). A follow of a
// private profile starts pending until the followee accepts it.
func m20261101163000CreateFollowsTable() schema.Migration {
	return postgresMigration("20261101163000_create_follows_table",
		[]string{
			`CREATE TABLE follows (
  follower_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  followee_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  state text NOT NULL CHECK (state IN ('accepted','pending')),
  created_at timestamptz NOT NULL DEFAULT now(),
  accepted_at timestamptz,
  PRIMARY KEY (follower_id, followee_id),
  CHECK (follower_id <> followee_id))`,
			`CREATE INDEX follows_followee_idx ON follows (followee_id, state)`,
			`ALTER TABLE user_notifications DROP CONSTRAINT user_notifications_kind_check,
  ADD CONSTRAINT user_notifications_kind_check CHECK (kind IN ('suggestion_decided','suggestion_reverted',
    'suggestion_expired','report_resolved','rating_flagged','list_hidden','system','new_follower','follow_request'))`,
		},
		[]string{
			`DELETE FROM user_notifications WHERE kind IN ('new_follower','follow_request')`,
			`ALTER TABLE user_notifications DROP CONSTRAINT user_notifications_kind_check,
  ADD CONSTRAINT user_notifications_kind_check CHECK (kind IN ('suggestion_decided','suggestion_reverted',
    'suggestion_expired','report_resolved','rating_flagged','list_hidden','system'))`,
			`DROP TABLE IF EXISTS follows`,
		},
	)
}
