package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261031190000CreateSuggestionReviews stores moderator recommendations (SDD 04 FR-04-17, owner Q-A15): moderators
// review, only admins apply or reject.
func m20261031190000CreateSuggestionReviews() schema.Migration {
	return postgresMigration("20261031190000_create_suggestion_reviews",
		[]string{
			`CREATE TABLE suggestion_reviews (
  suggestion_id uuid NOT NULL REFERENCES suggestions(id),
  reviewer_id bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  recommendation text NOT NULL CHECK (recommendation IN ('approve','reject')),
  note text CHECK (char_length(note) <= 1000),
  suggestion_version integer NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (suggestion_id, reviewer_id)
)`,
			`CREATE INDEX suggestion_reviews_reviewer_idx ON suggestion_reviews (reviewer_id)`,
		},
		[]string{`DROP TABLE IF EXISTS suggestion_reviews`},
	)
}
