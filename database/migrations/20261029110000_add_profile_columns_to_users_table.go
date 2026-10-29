package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029110000AddProfileColumnsToUsersTable adds the SDD 03 §5 profile,
// role, ban and deletion columns to the scaffold's users table. Ids stay the
// scaffold's bigint; legacy Better Auth ids live in legacy_id so the data
// migration (P2.F) can map them (SDD 03 §5 amendment, 2026-10-29).
func m20261029110000AddProfileColumnsToUsersTable() schema.Migration {
	return schema.Migration{
		ID: "20261029110000_add_profile_columns_to_users_table",
		Up: func(s *schema.Schema) error {
			if err := s.Table("users", func(t *schema.Table) {
				t.String("legacy_id").Nullable()
				// A separate index: SQLite cannot ADD COLUMN ... UNIQUE.
				t.UniqueIndex("legacy_id")
				t.Timestamp("password_set_at").Nullable()
				t.String("username").Nullable()
				t.String("display_username").Nullable()
				t.Text("image_url").Nullable()
				t.String("locale").Default("en")
				t.Boolean("profile_private").Default(false)
				t.Boolean("personalization_opt_out").Default(false)
				t.String("role").Default("user")
				t.Timestamp("banned_at").Nullable()
				t.Text("ban_reason").Nullable()
				t.Timestamp("ban_expires_at").Nullable()
				t.Timestamp("deletion_scheduled_at").Nullable()
				t.Timestamp("anonymized_at").Nullable()
			}); err != nil {
				return err
			}
			if s.Dialect() != "postgres" {
				return nil // CHECKs with regex, partial and expression indexes: PostgreSQL only (scaffold tests run on SQLite, VF-033)
			}
			return s.Raw(`ALTER TABLE users
  ADD CONSTRAINT users_username_format CHECK (username IS NULL OR username ~ '^[a-z]{4,16}$'),
  ADD CONSTRAINT users_locale_valid CHECK (locale IN ('en','pt-BR','es')),
  ADD CONSTRAINT users_role_valid CHECK (role IN ('user','admin')),
  ADD CONSTRAINT users_ban_reason_length CHECK (ban_reason IS NULL OR char_length(ban_reason) <= 500),
  ADD CONSTRAINT users_ban_expiry_needs_ban CHECK (ban_expires_at IS NULL OR banned_at IS NOT NULL);
CREATE UNIQUE INDEX users_username_uq ON users (username) WHERE username IS NOT NULL;
CREATE UNIQUE INDEX users_email_lower_uq ON users (lower(email));
CREATE INDEX users_deletion_due_idx ON users (deletion_scheduled_at) WHERE deletion_scheduled_at IS NOT NULL`)
		},
		Down: func(s *schema.Schema) error {
			if s.Dialect() == "postgres" {
				if err := s.Raw(
					`ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_format, DROP CONSTRAINT IF EXISTS users_locale_valid, DROP CONSTRAINT IF EXISTS users_role_valid, DROP CONSTRAINT IF EXISTS users_ban_reason_length, DROP CONSTRAINT IF EXISTS users_ban_expiry_needs_ban; DROP INDEX IF EXISTS users_username_uq; DROP INDEX IF EXISTS users_email_lower_uq; DROP INDEX IF EXISTS users_deletion_due_idx`,
				); err != nil {
					return err
				}
			}
			return s.Table("users", func(t *schema.Table) {
				t.DropColumn("legacy_id", "password_set_at", "username", "display_username", "image_url", "locale",
					"profile_private", "personalization_opt_out", "role", "banned_at", "ban_reason", "ban_expires_at",
					"deletion_scheduled_at", "anonymized_at")
			})
		},
	}
}
