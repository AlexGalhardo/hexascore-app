package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261102090000AddIndexNowCursor stores how far the IndexNow sweep (jobs.IndexNowChanges) has read: the last
// api_changes id for items and the last entities.updated_at. It starts at the current end, so the first run does not
// resubmit the whole catalogue.
func m20261102090000AddIndexNowCursor() schema.Migration {
	return postgresMigration("20261102090000_add_indexnow_cursor",
		[]string{
			`CREATE TABLE indexnow_cursor (
  id smallint PRIMARY KEY CHECK (id = 1),
  last_change_id bigint NOT NULL,
  entities_since timestamptz NOT NULL
)`,
			`INSERT INTO indexnow_cursor (id, last_change_id, entities_since)
  SELECT 1, coalesce(max(id), 0), now() FROM api_changes`,
		},
		[]string{`DROP TABLE IF EXISTS indexnow_cursor`},
	)
}
