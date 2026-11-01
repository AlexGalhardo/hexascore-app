package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261031203100AddCommunityMailOptOut stores the email opt-out of community notices (SDD 04 FR-04-14); the
// in-app notification is always written.
func m20261031203100AddCommunityMailOptOut() schema.Migration {
	return postgresMigration("20261031203100_add_community_mail_opt_out",
		[]string{`ALTER TABLE users ADD COLUMN community_mail_opt_out boolean NOT NULL DEFAULT false`},
		[]string{`ALTER TABLE users DROP COLUMN IF EXISTS community_mail_opt_out`},
	)
}
