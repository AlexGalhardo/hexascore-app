package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261101190000ModerateReviewsAndComments lets members report reviews and comments and admins hide or delete
// them (SDD 14 §12.11, FR-14-9): a `hidden` flag on both tables, the new report targets and outcomes, and the
// author's notification kinds.
func m20261101190000ModerateReviewsAndComments() schema.Migration {
	return postgresMigration("20261101190000_moderate_reviews_and_comments",
		[]string{
			`ALTER TABLE reviews ADD COLUMN hidden boolean NOT NULL DEFAULT false`,
			`ALTER TABLE comments ADD COLUMN hidden boolean NOT NULL DEFAULT false`,
			`ALTER TABLE reports DROP CONSTRAINT reports_target_type_check,
  ADD CONSTRAINT reports_target_type_check CHECK (target_type IN ('list','user','suggestion','review','comment')),
  DROP CONSTRAINT reports_outcome_check,
  ADD CONSTRAINT reports_outcome_check CHECK (outcome IN ('dismissed','list_hidden','user_banned',
    'suggestion_rejected','no_action','content_hidden','content_deleted','content_unhidden'))`,
			`ALTER TABLE user_notifications DROP CONSTRAINT user_notifications_kind_check,
  ADD CONSTRAINT user_notifications_kind_check CHECK (kind IN ('suggestion_decided','suggestion_reverted',
    'suggestion_expired','report_resolved','rating_flagged','list_hidden','system','new_follower','follow_request',
    'new_comment','new_reaction','content_hidden','content_deleted'))`,
		},
		[]string{
			`DELETE FROM user_notifications WHERE kind IN ('content_hidden','content_deleted')`,
			`ALTER TABLE user_notifications DROP CONSTRAINT user_notifications_kind_check,
  ADD CONSTRAINT user_notifications_kind_check CHECK (kind IN ('suggestion_decided','suggestion_reverted',
    'suggestion_expired','report_resolved','rating_flagged','list_hidden','system','new_follower','follow_request',
    'new_comment','new_reaction'))`,
			`DELETE FROM reports WHERE target_type IN ('review','comment')`,
			`ALTER TABLE reports DROP CONSTRAINT reports_target_type_check,
  ADD CONSTRAINT reports_target_type_check CHECK (target_type IN ('list','user','suggestion')),
  DROP CONSTRAINT reports_outcome_check,
  ADD CONSTRAINT reports_outcome_check CHECK (outcome IN ('dismissed','list_hidden','user_banned',
    'suggestion_rejected','no_action'))`,
			`ALTER TABLE comments DROP COLUMN IF EXISTS hidden`,
			`ALTER TABLE reviews DROP COLUMN IF EXISTS hidden`,
		},
	)
}
