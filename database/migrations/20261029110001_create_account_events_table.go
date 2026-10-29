package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261029110001CreateAccountEventsTable is the user-facing security history
// and admin action log (SDD 03 §5). data holds ids and reasons only, never
// emails or IPs; rows older than two years are deleted by a monthly job.
func m20261029110001CreateAccountEventsTable() schema.Migration {
	return schema.Migration{
		ID: "20261029110001_create_account_events_table",
		Up: func(s *schema.Schema) error {
			return s.Create("account_events", func(t *schema.Table) {
				t.ID()
				t.ForeignID("user_id").Constrained("users").CascadeOnDelete()
				t.ForeignID("actor_id").Nullable().Constrained("users").NullOnDelete()
				t.String("kind")
				t.JSON("data").Default("{}")
				t.Timestamp("created_at")
				t.Index("user_id", "created_at")
			})
		},
		Down: func(s *schema.Schema) error { return s.DropIfExists("account_events") },
	}
}
