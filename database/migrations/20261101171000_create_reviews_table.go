package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261101171000CreateReviewsTable adds reviews: one optional text per rating, shown while the rating exists
// (SDD 14 §12.10). helpful_count is kept by the reactions of P7.B.1.2.
func m20261101171000CreateReviewsTable() schema.Migration {
	return postgresMigration("20261101171000_create_reviews_table",
		[]string{
			`CREATE TABLE reviews (
  id uuid PRIMARY KEY,
  user_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_id uuid NOT NULL REFERENCES items(id) ON DELETE CASCADE,
  body text NOT NULL CHECK (char_length(body) BETWEEN 1 AND 5000),
  spoiler boolean NOT NULL DEFAULT false,
  helpful_count integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, item_id))`,
			`CREATE INDEX reviews_item_idx ON reviews (item_id, helpful_count DESC, created_at DESC)`,
		},
		[]string{`DROP TABLE IF EXISTS reviews`},
	)
}
