package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261102190000AddUserFksToScoringTables gives score_actions (user_id, actor), search_events.user_id and
// rating_flags.user_id a foreign key to users (audit T-05). The purge keeps the users row, so the policies only matter
// for a hard delete; the purge itself erases or anonymises these rows (app/accounts/deletion.go). Orphans are cleaned
// first and each constraint is added NOT VALID; the scan runs in m20261102190100ValidateUserFksOnScoringTables, outside
// this transaction, because a VALIDATE here would hold the ADD CONSTRAINT lock on users and the scoring tables for the
// whole scan and block writes during the pre-deploy.
func m20261102190000AddUserFksToScoringTables() schema.Migration {
	return postgresMigration("20261102190000_add_user_fks_to_scoring_tables",
		[]string{
			`UPDATE score_actions SET user_id = NULL WHERE user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = score_actions.user_id)`,
			`UPDATE score_actions SET actor = NULL WHERE actor IS NOT NULL AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = score_actions.actor)`,
			`UPDATE search_events SET user_id = NULL WHERE user_id IS NOT NULL AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = search_events.user_id)`,
			`DELETE FROM rating_flags f WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = f.user_id)`,
			`ALTER TABLE score_actions
  ADD CONSTRAINT score_actions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL NOT VALID,
  ADD CONSTRAINT score_actions_actor_fkey FOREIGN KEY (actor) REFERENCES users(id) ON DELETE SET NULL NOT VALID`,
			`ALTER TABLE search_events
  ADD CONSTRAINT search_events_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL NOT VALID`,
			`ALTER TABLE rating_flags
  ADD CONSTRAINT rating_flags_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE NOT VALID`,
		},
		[]string{
			`ALTER TABLE score_actions DROP CONSTRAINT IF EXISTS score_actions_user_id_fkey, DROP CONSTRAINT IF EXISTS score_actions_actor_fkey`,
			`ALTER TABLE search_events DROP CONSTRAINT IF EXISTS search_events_user_id_fkey`,
			`ALTER TABLE rating_flags DROP CONSTRAINT IF EXISTS rating_flags_user_id_fkey`,
		},
	)
}

// m20261102190100ValidateUserFksOnScoringTables validates the constraints added NOT VALID by the previous migration.
// Outside a transaction each VALIDATE takes only SHARE UPDATE EXCLUSIVE, so writes continue during the scan; a rerun
// after a partial failure is safe because validating a valid constraint is a no-op. Down has nothing to undo.
func m20261102190100ValidateUserFksOnScoringTables() schema.Migration {
	m := postgresMigration("20261102190100_validate_user_fks_on_scoring_tables",
		[]string{
			`ALTER TABLE score_actions VALIDATE CONSTRAINT score_actions_user_id_fkey`,
			`ALTER TABLE score_actions VALIDATE CONSTRAINT score_actions_actor_fkey`,
			`ALTER TABLE search_events VALIDATE CONSTRAINT search_events_user_id_fkey`,
			`ALTER TABLE rating_flags VALIDATE CONSTRAINT rating_flags_user_id_fkey`,
		},
		nil,
	)
	m.NoTransaction = true

	return m
}
