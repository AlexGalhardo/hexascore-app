package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261101140000AddIsSeedToUsers flags the fake legacy seed users (owner decision 2026-11-01) so they can never
// sign in and `seed:purge` can remove them in one command. The partial index only serves purge and is PostgreSQL-only.
func m20261101140000AddIsSeedToUsers() schema.Migration {
	return schema.Migration{
		ID: "20261101140000_add_is_seed_to_users",
		Up: func(s *schema.Schema) error {
			if err := s.Table("users", func(t *schema.Table) {
				t.Boolean("is_seed").Default(false)
			}); err != nil {
				return err
			}

			if s.Dialect() != "postgres" {
				return nil
			}

			return s.Raw(`CREATE INDEX users_is_seed_idx ON users (id) WHERE is_seed`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() == "postgres" {
				if err := s.Raw(`DROP INDEX IF EXISTS users_is_seed_idx`); err != nil {
					return err
				}
			}

			return s.Table("users", func(t *schema.Table) { t.DropColumn("is_seed") })
		},
	}
}
