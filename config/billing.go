// Package config holds the application's own settings, read from the
// environment (.env).
package config

import (
	"errors"
	"slices"
	"strings"

	vconfig "github.com/AlexGalhardo/vukem/config"
)

// Plan is a subscription plan sold with Stripe Checkout.
type Plan struct {
	// Name is the plan's key in URLs and in OnPlan ("pro").
	Name string `json:"name"`
	// Label is shown on the billing page.
	Label string `json:"label"`
	// Price is the Stripe price id (price_...), from STRIPE_PRICE_<NAME>. The
	// Stripe price carries one currency_options entry per supported currency.
	Price string `json:"-"`
	// Amounts are display-only minor units per currency; Stripe charges the
	// price's own currency_options amount.
	Amounts map[string]int `json:"amounts"`
}

// Currencies are the checkout currencies (SDD 05 §6.3); the first is the default.
var Currencies = []string{"usd", "brl", "eur"}

// Currency returns c lower-cased when it is supported, else ok is false.
func Currency(c string) (string, bool) {
	c = strings.ToLower(strings.TrimSpace(c))
	return c, slices.Contains(Currencies, c)
}

// plans lists the plans on sale, in display order. Add yours here and set
// STRIPE_PRICE_<NAME> to a recurring price id from the Stripe Dashboard.
func plans() []Plan {
	return []Plan{
		{Name: "pro", Label: "Pro", Amounts: map[string]int{"brl": 2900, "usd": 600, "eur": 600}},
		{Name: "business", Label: "Business", Amounts: map[string]int{"brl": 14900, "usd": 2900, "eur": 2700}},
	}
}

// Catalogue lists the paid plans with their display amounts; it needs no Stripe settings, so public pages can show it.
func Catalogue() []Plan { return plans() }

// Billing configures Stripe (vukem install:billing).
type Billing struct {
	// Key is the Stripe API key (STRIPE_KEY). Use a restricted key (rk_...)
	// with write access to Customers, Checkout Sessions and Customer portal
	// and read access to Subscriptions, nothing else.
	Key vconfig.Secret
	// WebhookSecret verifies webhooks (STRIPE_WEBHOOK_SECRET, whsec_...).
	// `stripe listen --forward-to localhost:3000/stripe/webhook` prints one
	// for development.
	WebhookSecret vconfig.Secret
	// Plans are the plans on sale.
	Plans []Plan
}

// LoadBilling reads the billing settings from src (app.Source).
func LoadBilling(src vconfig.Lookuper) Billing {
	get := func(key string) string {
		v, _ := src.Lookup(key)
		return strings.TrimSpace(v)
	}

	b := Billing{
		Key:           vconfig.NewSecret(get("STRIPE_KEY")),
		WebhookSecret: vconfig.NewSecret(get("STRIPE_WEBHOOK_SECRET")),
	}

	for _, p := range plans() {
		p.Price = get("STRIPE_PRICE_" + strings.ToUpper(p.Name))
		b.Plans = append(b.Plans, p)
	}

	return b
}

// Validate reports the missing settings; production refuses to boot
// without them.
func (b Billing) Validate() error {
	var errs []error

	if b.Key.IsZero() {
		errs = append(errs, errors.New("STRIPE_KEY is not set"))
	}

	if b.WebhookSecret.IsZero() {
		errs = append(errs, errors.New("STRIPE_WEBHOOK_SECRET is not set"))
	}

	for _, p := range b.Plans {
		if p.Price == "" {
			errs = append(errs, errors.New("STRIPE_PRICE_"+strings.ToUpper(p.Name)+" is not set"))
		}
	}

	return errors.Join(errs...)
}

// Plan returns the plan named name.
func (b Billing) Plan(name string) (Plan, bool) {
	for _, p := range b.Plans {
		if p.Name == name {
			return p, true
		}
	}

	return Plan{}, false
}

// PlanForPrice returns the plan that sells price.
func (b Billing) PlanForPrice(price string) (Plan, bool) {
	for _, p := range b.Plans {
		if p.Price != "" && p.Price == price {
			return p, true
		}
	}

	return Plan{}, false
}
