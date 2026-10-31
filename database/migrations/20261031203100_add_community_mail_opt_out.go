package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261031203100AddCommunityMailOptOut stores the email opt-out of community notices (SDD 04 FR-04-14); the
// in-app notification is always written. Both dialects: the profile page reads it on SQLite tests too.
func m20261031203100AddCommunityMailOptOut() schema.Migration {
	return schema.Migration{
		ID: "20261031203100_add_community_mail_opt_out",
		Up: func(s *schema.Schema) error {
			return s.Table("users", func(t *schema.Table) {
				t.Boolean("community_mail_opt_out").Default(false)
			})
		},
		Down: func(s *schema.Schema) error {
			return s.Table("users", func(t *schema.Table) { t.DropColumn("community_mail_opt_out") })
		},
	}
}
