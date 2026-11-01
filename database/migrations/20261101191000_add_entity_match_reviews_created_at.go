package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261101191000AddEntityMatchReviewsCreatedAt gives the review queue an age, which DQ-I-11 (open > 7 days) reads.
func m20261101191000AddEntityMatchReviewsCreatedAt() schema.Migration {
	return postgresMigration(
		"20261101191000_add_entity_match_reviews_created_at",
		[]string{`ALTER TABLE entity_match_reviews ADD COLUMN created_at timestamptz NOT NULL DEFAULT now()`},
		[]string{`ALTER TABLE entity_match_reviews DROP COLUMN IF EXISTS created_at`},
	)
}
