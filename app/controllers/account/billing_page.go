package account

import (
	"strings"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Show renders the plans and the current subscription.
func (h *BillingHandlers) Show(c *web.Ctx) error {
	u, err := auth.User[models.User](c, "web")
	if err != nil {
		return err
	}

	page, err := h.page(c, u)
	if err != nil {
		return err
	}

	checkout := any(false)

	switch c.Query("checkout") {
	case "success":
		checkout = true
	case "canceled":
		checkout = "canceled"
	}

	return inertia.Render(c, "Billing/Index", inertia.Props{
		"currency":     defaultCurrency(i18n.Locale(c.Context())),
		"plans":        page.Plans,
		"subscription": page.Subscription,
		"subscribed":   page.Subscribed,
		"checkout":     checkout,
	})
}

// defaultCurrency is the checkout currency preselected for a locale (SDD 05 section 6.3).
func defaultCurrency(locale string) string {
	switch {
	case strings.HasPrefix(locale, "pt"):
		return "brl"
	case strings.HasPrefix(locale, "es"):
		return "eur"
	}

	return "usd"
}
