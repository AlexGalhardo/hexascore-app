package routes

import (
	"net/http"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/events"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/billing"
	"github.com/AlexGalhardo/hexascore-app/app/controllers/account"
)

// Billing registers the billing page, Stripe Checkout, the Billing Portal
// and the Stripe webhook (vukem install:billing). Protect your own routes
// with b.Subscribed() or b.OnPlan("business") after auth.Required.
func Billing(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	b, err := billing.New(app)
	if err != nil {
		return err
	}

	// POST /stripe/webhook runs before the session and CSRF middleware:
	// Stripe authenticates with Stripe-Signature, not cookies.
	r.UseServer(b.Webhook())
	events.Listen(app.Events, b.HandleWebhook)

	h := account.NewBilling(b)
	required := auth.Required(acc.Web)
	verified := auth.Verified(acc.Web)

	r.Get("/billing", h.Show).Name("billing.show").Use(required, verified)
	r.Post("/billing/checkout/{plan}", h.Checkout).Name("billing.checkout").Use(required, verified)
	r.Post("/billing/portal", h.Portal).Name("billing.portal").Use(required, verified)

	// Legacy checkout page (SDD 05 section 6.3): old links and Stripe return URLs land on /billing.
	r.Get("/profile/api/checkout", func(c *web.Ctx) error {
		target := "/billing"

		switch c.Query("status") {
		case "success":
			target += "?checkout=success"
		case "canceled":
			target += "?checkout=canceled"
		}

		return c.RedirectStatus(http.StatusMovedPermanently, target)
	}).Name("billing.legacy")

	return nil
}
