package migrations

import "github.com/AlexGalhardo/vukem/schema"

// m20261030150000AddWebhookFailureStreak counts deliveries that used up their
// retries. Three in a row disable the endpoint (SDD 05 §7.3).
func m20261030150000AddWebhookFailureStreak() schema.Migration {
	return postgresMigration("20261030150000_add_webhook_failure_streak",
		[]string{`ALTER TABLE webhook_endpoints ADD COLUMN failure_streak integer NOT NULL DEFAULT 0`},
		[]string{`ALTER TABLE webhook_endpoints DROP COLUMN failure_streak`},
	)
}
