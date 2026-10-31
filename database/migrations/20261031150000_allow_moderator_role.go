package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261031150000AllowModeratorRole widens users.role for the trusted-moderator role of SDD 04 §6 (D-15). The role
// grants nothing unless the community_trusted_moderators flag is on.
func m20261031150000AllowModeratorRole() schema.Migration {
	return postgresMigration("20261031150000_allow_moderator_role",
		[]string{
			`ALTER TABLE users DROP CONSTRAINT users_role_valid,
  ADD CONSTRAINT users_role_valid CHECK (role IN ('user','moderator','admin'))`,
		},
		[]string{
			`UPDATE users SET role = 'user' WHERE role = 'moderator'`,
			`ALTER TABLE users DROP CONSTRAINT users_role_valid,
  ADD CONSTRAINT users_role_valid CHECK (role IN ('user','admin'))`,
		},
	)
}
