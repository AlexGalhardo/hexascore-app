//go:build integration

package legacyimport_test

import (
	"io"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

func TestBillingStage(t *testing.T) {
	t.Parallel()

	app, db := fixtureDB(t)
	ctx := t.Context()

	stages, err := legacyimport.Select(legacyimport.Stages(), []string{"users", "billing"})
	if err != nil {
		t.Fatal(err)
	}

	r := &legacyimport.Runner{DB: db, Batch: 2, Out: io.Discard, Now: time.Now}
	if err := r.Run(ctx, stages); err != nil {
		t.Fatalf("run: %v", err)
	}

	checks := []struct{ name, query, want string }{
		{"stripe customer", `SELECT stripe_customer_id FROM users WHERE legacy_id = 'u01'`, "cus_legacy1"},
		{
			"key hash verbatim",
			`SELECT k.hash = l.hash FROM api_keys k JOIN legacy_import.api_keys l ON l.id = k.id`,
			"true",
		},
		{
			"subscriptions",
			`SELECT string_agg(u.legacy_id || ':' || s.plan || '/' || s.provider || '/' || s.state || '/' ||
  s.stripe_subscription_id, ',') FROM subscriptions s JOIN users u ON u.id = s.user_id`,
			"u01:pro/stripe/active/sub_legacy1",
		},
		{"usage", `SELECT sum(count)::text FROM api_usage_daily`, "200"},
		{"credit balances", `SELECT string_agg(u.legacy_id || '=' || b.remaining, ',' ORDER BY u.legacy_id)
FROM api_credit_balances b JOIN users u ON u.id = b.user_id`, "u01=23800,u05=0"},
		{"ledger verbatim", `SELECT count(*)::text FROM credit_ledger c JOIN legacy_import.credit_ledger l
  ON l.id = c.id AND l.delta = c.delta AND l.reason = c.reason AND l.ref = c.ref AND l.created_at = c.created_at`, "4"},
		{
			"pix verbatim", `SELECT p.provider_ref || '/' || p.status || '/' || p.amount_cents FROM pix_charges p`,
			"pix_legacy1/paid/1990",
		},
	}

	for _, c := range checks {
		var got string
		if err := db.QueryRowContext(ctx, c.query).Scan(&got); err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}

		if got != c.want {
			t.Errorf("%s = %q, want %q", c.name, got, c.want)
		}
	}

	// The legacy key authenticates on the public API after the import.
	app.Client(t).WithHeader("Authorization", "Bearer hx_live_Ab3dEf6hIj9kLm2nOp5qRs8tUv1wXy4z").
		GetJSON("/api/v1/games").AssertOK()
	app.Client(t).WithHeader("Authorization", "Bearer hx_live_Ab3dEf6hIj9kLm2nOp5qRs8tUv1wXy4Z").
		GetJSON("/api/v1/games").AssertStatus(401)
}
