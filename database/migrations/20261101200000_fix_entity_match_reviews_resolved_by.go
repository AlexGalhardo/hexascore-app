package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261101200000FixEntityMatchReviewsResolvedBy turns resolved_by into a users.id: the P6.A.1.2 table typed it uuid
// from the SDD text, but users.id is a bigint (SDD 03), so the review queue could never record who resolved a match.
func m20261101200000FixEntityMatchReviewsResolvedBy() schema.Migration {
	return postgresMigration(
		"20261101200000_fix_entity_match_reviews_resolved_by",
		[]string{
			`ALTER TABLE entity_match_reviews ALTER COLUMN resolved_by TYPE bigint USING NULL`,
			`ALTER TABLE entity_match_reviews ADD CONSTRAINT entity_match_reviews_resolved_by_fkey
  FOREIGN KEY (resolved_by) REFERENCES users(id) ON DELETE SET NULL`,
		},
		[]string{
			`ALTER TABLE entity_match_reviews DROP CONSTRAINT IF EXISTS entity_match_reviews_resolved_by_fkey`,
			`ALTER TABLE entity_match_reviews ALTER COLUMN resolved_by TYPE uuid USING NULL`,
		},
	)
}
