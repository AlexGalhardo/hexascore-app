package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261101140000AddCommunityMailOptOutSQLite adds users.community_mail_opt_out on SQLite, where the profile tests run.
// On PostgreSQL 20261031203100 already added it and this migration does nothing; that migration is on main, so it is
// not edited.
func m20261101140000AddCommunityMailOptOutSQLite() schema.Migration {
	skip := func(s *schema.Schema) bool { return s.Dialect() == "postgres" }

	return schema.Migration{
		ID: "20261101140000_add_community_mail_opt_out_sqlite",
		Up: func(s *schema.Schema) error {
			if skip(s) {
				return nil
			}

			return s.Table("users", func(t *schema.Table) {
				t.Boolean("community_mail_opt_out").Default(false)
			})
		},
		Down: func(s *schema.Schema) error {
			if skip(s) {
				return nil
			}

			return s.Table("users", func(t *schema.Table) { t.DropColumn("community_mail_opt_out") })
		},
	}
}
