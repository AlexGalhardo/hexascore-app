package migrations

import "github.com/AlexGalhardo/vukem/schema"

func m20261029170000AddBillingTables() schema.Migration {
	return schema.Migration{
		ID: "20261029170000_add_billing_tables",
		Up: func(s *schema.Schema) error {
			err := s.Table("users", func(t *schema.Table) {
				t.String("stripe_customer_id").Nullable()
				t.UniqueIndex("stripe_customer_id")
			})
			if err != nil {
				return err
			}

			if err = s.Create("subscriptions", func(t *schema.Table) {
				t.ID()
				t.ForeignID("user_id").Constrained("users").CascadeOnDelete()
				t.String("stripe_subscription_id").Unique()
				t.String("status")
				t.String("price_id")
				t.String("plan").Default("")
				t.String("provider").Default("stripe")
				t.String("state").Default("active")
				t.Timestamp("current_period_end").Nullable()
				t.Boolean("cancel_at_period_end").Default(false)
				t.Timestamps()
				t.Index("user_id")
			}); err != nil {
				return err
			}

			// At most one live (active) row per user (SDD 05 section 7.5).
			return s.Raw(
				"CREATE UNIQUE INDEX subscriptions_user_live_unique ON subscriptions (user_id) WHERE state = 'active'",
			)
		},
		Down: func(s *schema.Schema) error {
			if err := s.DropIfExists("subscriptions"); err != nil {
				return err
			}

			return s.Table("users", func(t *schema.Table) {
				t.DropIndex("users_stripe_customer_id_unique")
				t.DropColumn("stripe_customer_id")
			})
		},
	}
}
