package tests

import (
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/orm"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/app/billing"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

func TestPlanStateMachine(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	future := orm.NullTime{}
	future.V, future.Valid = now.Add(time.Hour), true
	past := orm.NullTime{}
	past.V, past.Valid = now.Add(-time.Hour), true

	for status, want := range map[string]string{
		"active": "active", "trialing": "active", "past_due": "past_due", "unpaid": "past_due",
		"incomplete": "past_due", "paused": "past_due", "canceled": "canceled", "incomplete_expired": "canceled",
	} {
		if got := billing.StateOf(status); got != want {
			t.Errorf("StateOf(%q) = %q, want %q", status, got, want)
		}
	}

	cases := []struct {
		name string
		sub  models.Subscription
		want string
	}{
		{"active in period", models.Subscription{Plan: "pro", State: "active", CurrentPeriodEnd: future}, api.PlanPro},
		{"active no end", models.Subscription{Plan: "business", State: "active"}, api.PlanBusiness},
		{"period ended", models.Subscription{Plan: "pro", State: "active", CurrentPeriodEnd: past}, api.PlanFree},
		{"past due", models.Subscription{Plan: "pro", State: "past_due", CurrentPeriodEnd: future}, api.PlanFree},
		{"canceled", models.Subscription{Plan: "pro", State: "canceled"}, api.PlanFree},
		{"expired", models.Subscription{Plan: "pro", State: "expired"}, api.PlanFree},
		{"unknown plan", models.Subscription{Plan: "", State: "active"}, api.PlanFree},
	}
	for _, tc := range cases {
		if got := billing.EffectivePlan(tc.sub, now); got != tc.want {
			t.Errorf("%s: EffectivePlan = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestPlanResolverFollowsWebhooks(t *testing.T) {
	t.Parallel()

	f := newFakeStripe(t)
	app := billingApp(t, f)
	u := newUser(t, app, "plan@example.com")
	setCustomer(t, app, u)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	resolve := billing.PlanResolver(db, time.Now)
	plan := func() string {
		p, err := resolve(t.Context(), u.ID)
		if err != nil {
			t.Fatal(err)
		}

		return p.ID
	}

	if got := plan(); got != api.PlanFree {
		t.Fatalf("no subscription: plan = %q", got)
	}

	f.setSubscription("active", "price_business")
	postWebhook(
		t,
		app,
		"evt_p1",
		"customer.subscription.created",
		map[string]any{"id": "sub_1", "customer": "cus_1"},
	).AssertOK()

	if got := plan(); got != api.PlanBusiness {
		t.Fatalf("active business: plan = %q", got)
	}

	f.setSubscription("past_due", "price_business")
	postWebhook(
		t,
		app,
		"evt_p2",
		"customer.subscription.updated",
		map[string]any{"id": "sub_1", "customer": "cus_1"},
	).AssertOK()

	if got := plan(); got != api.PlanFree {
		t.Fatalf("past_due: plan = %q", got)
	}
}

func TestLegacyCheckoutRedirect(t *testing.T) {
	t.Parallel()

	app := billingApp(t, newFakeStripe(t))
	c := app.Client(t)

	c.Get("/profile/api/checkout?plan=pro").AssertRedirect("/billing")
	c.Get("/profile/api/checkout?status=success").AssertRedirect("/billing?checkout=success")
}

func TestCancelLiveOnDeletion(t *testing.T) {
	t.Parallel()

	f := newFakeStripe(t)
	app := billingApp(t, f)
	u := newUser(t, app, "bye@example.com")
	setCustomer(t, app, u)
	f.setSubscription("active", "price_pro")
	postWebhook(
		t,
		app,
		"evt_d1",
		"customer.subscription.created",
		map[string]any{"id": "sub_1", "customer": "cus_1"},
	).AssertOK()

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	f.mu.Lock()
	f.failCancel = true
	f.mu.Unlock()

	if err := acc.Subscriptions.CancelLive(t.Context(), u.ID); err == nil {
		t.Fatal("a Stripe failure must be returned so the deletion stays scheduled")
	}

	f.mu.Lock()
	f.failCancel = false
	f.mu.Unlock()

	if err := acc.Subscriptions.CancelLive(t.Context(), u.ID); err != nil {
		t.Fatal(err)
	}

	// The client retries 5xx, so the failed run may have called Stripe several times.
	f.mu.Lock()
	calls := len(f.cancels)
	f.mu.Unlock()

	if err := acc.Subscriptions.CancelLive(t.Context(), u.ID); err != nil {
		t.Fatalf("second run must be a no-op: %v", err)
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.cancels) != calls {
		t.Fatalf("cancelled subscription cancelled again: %v", f.cancels)
	}
}
