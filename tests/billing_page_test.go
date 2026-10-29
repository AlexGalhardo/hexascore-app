package tests

import (
	"net/http"
	"testing"
)

func TestBillingPage(t *testing.T) {
	t.Parallel()

	f := newFakeStripe(t)
	app := billingApp(t, f)
	u := newUser(t, app, "page@example.com")
	c := app.Client(t).ActingAs("web", u)

	c.Get("/billing").AssertInertia("Billing/Index").
		HasProp("plans", 2).
		PropEquals("plans.1.name", "business").
		PropEquals("subscribed", false).
		PropEquals("subscription", nil)

	setCustomer(t, app, u)
	f.setSubscription("active", "price_business")
	postWebhook(t, app, "evt_1", "customer.subscription.created", map[string]any{"id": "sub_1"}).AssertOK()

	c.Get("/billing").AssertInertia("Billing/Index").
		PropEquals("subscribed", true).
		PropEquals("subscription.plan", "Business").
		PropEquals("subscription.status", "active").
		PropEquals("plans.1.current", true)

	// Inertia visits leave for Stripe with a full page load.
	c.Inertia("").PostJSON("/billing/portal", nil).
		AssertStatus(http.StatusConflict).
		AssertHeader("X-Inertia-Location", "https://billing.stripe.com/p/session/test_1")
}
