package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261031130000CreateCommunityHoneypotHits counts filled community honeypot
// fields. Three hits in seven days set the contributor tier to blocked.
func m20261031130000CreateCommunityHoneypotHits() schema.Migration {
	return postgresMigration("20261031130000_create_community_honeypot_hits", honeypotUp, honeypotDown)
}

var honeypotUp = []string{
	`CREATE TABLE community_honeypot_hits (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL
)`,
	`CREATE INDEX community_honeypot_hits_user_idx ON community_honeypot_hits (user_id, created_at)`,
}

var honeypotDown = []string{
	`DROP TABLE IF EXISTS community_honeypot_hits`,
}
