package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261101164000CreateActivityEvents adds the feed opt-in (default off while Q-A5 is open) and the activity
// events the feed reads (SDD 14 §12.4, ADR 0020).
func m20261101164000CreateActivityEvents() schema.Migration {
	return postgresMigration("20261101164000_create_activity_events",
		[]string{
			`ALTER TABLE users ADD COLUMN feed_opt_in boolean NOT NULL DEFAULT false`,
			`CREATE TABLE activity_events (
  id uuid PRIMARY KEY,
  actor_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  kind text NOT NULL CHECK (kind IN ('rating_saved','status_changed','list_created','list_item_added',
    'review_published')),
  item_id uuid REFERENCES items(id) ON DELETE CASCADE,
  list_id uuid REFERENCES lists(id) ON DELETE CASCADE,
  payload jsonb NOT NULL DEFAULT '{}',
  created_at timestamptz NOT NULL DEFAULT now())`,
			`CREATE INDEX activity_events_actor_idx ON activity_events (actor_id, id DESC)`,
			`CREATE INDEX activity_events_created_idx ON activity_events (created_at)`,
		},
		[]string{
			`DROP TABLE IF EXISTS activity_events`,
			`ALTER TABLE users DROP COLUMN IF EXISTS feed_opt_in`,
		},
	)
}
