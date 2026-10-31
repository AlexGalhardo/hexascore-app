package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261031190000CreateSimilarTables creates the nightly item neighbours and the 24 h "For you" cache (SDD 08 section 5).
// users.id is bigint in the real schema (SDD 03 section 5 and SDD 08 say text; follows the table, flagged in the report).
func m20261031190000CreateSimilarTables() schema.Migration {
	return postgresMigration("20261031190000_create_similar_tables",
		[]string{
			`CREATE TABLE item_similar (
  item_id uuid PRIMARY KEY REFERENCES items(id) ON DELETE CASCADE,
  similar_ids uuid[] NOT NULL,
  scores real[] NOT NULL,
  cooccur int[] NOT NULL,
  feature_hash bytea NOT NULL,
  computed_at timestamptz NOT NULL
)`,
			`CREATE TABLE user_recommendations (
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_type text NOT NULL,
  item_ids uuid[] NOT NULL,
  reason_item_ids uuid[] NOT NULL,
  events_at_compute bigint NOT NULL,
  computed_at timestamptz NOT NULL,
  PRIMARY KEY (user_id, item_type)
)`,
		},
		[]string{`DROP TABLE IF EXISTS user_recommendations`, `DROP TABLE IF EXISTS item_similar`},
	)
}
