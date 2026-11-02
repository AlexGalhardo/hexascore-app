// Package config holds the application's own settings, read from the
// environment (.env).
package config

import (
	"cmp"
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
	// Price is the Stripe price id (price_...), from STRIPE_PRICE_<NAME>_<MODE>. The
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
// STRIPE_PRICE_<NAME>_SANDBOX and STRIPE_PRICE_<NAME>_LIVE to recurring price ids from the Stripe Dashboard.
func plans() []Plan {
	return []Plan{
		{Name: "hobby", Label: "Hobby", Amounts: map[string]int{"brl": 2000, "usd": 500, "eur": 500}},
		{Name: "pro", Label: "PRO", Amounts: map[string]int{"brl": 5000, "usd": 1000, "eur": 1000}},
	}
}

// Catalogue lists the paid plans with their display amounts; it needs no Stripe settings, so public pages can show it.
func Catalogue() []Plan { return plans() }

// Stripe modes (STRIPE_MODE): which variable set the app reads, STRIPE_*_SANDBOX or STRIPE_*_LIVE (owner, 2026-11-02).
const (
	ModeSandbox = "sandbox"
	ModeLive    = "live"
)

// Billing configures Stripe (vukem install:billing).
type Billing struct {
	// Mode is "sandbox" or "live" (STRIPE_MODE; unset is sandbox). It picks the suffix of every variable below.
	Mode string
	// modeSet is false when STRIPE_MODE is absent; production requires it, so it never runs on a default.
	modeSet bool
	// Key is the Stripe API key (STRIPE_KEY_<MODE>). Use a restricted key (rk_...)
	// with write access to Customers, Checkout Sessions and Customer portal
	// and read access to Subscriptions, nothing else.
	Key vconfig.Secret
	// WebhookSecret verifies webhooks (STRIPE_WEBHOOK_SECRET_<MODE>, whsec_...).
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

	mode := strings.ToLower(get("STRIPE_MODE"))
	b := Billing{Mode: cmp.Or(mode, ModeSandbox), modeSet: mode != ""}

	b.Key = vconfig.NewSecret(get(b.Var("STRIPE_KEY")))
	b.WebhookSecret = vconfig.NewSecret(get(b.Var("STRIPE_WEBHOOK_SECRET")))

	for _, p := range plans() {
		p.Price = get(b.Var("STRIPE_PRICE_" + strings.ToUpper(p.Name)))
		b.Plans = append(b.Plans, p)
	}

	return b
}

// Var is the name of the variable base for the current mode, e.g. STRIPE_KEY_LIVE.
func (b Billing) Var(base string) string { return base + "_" + strings.ToUpper(b.Mode) }

// Validate reports the missing settings; production refuses to boot
// without them.
func (b Billing) Validate() error {
	var errs []error

	switch {
	case !b.modeSet:
		errs = append(errs, errors.New("STRIPE_MODE is not set (sandbox or live)"))
	case b.Mode != ModeSandbox && b.Mode != ModeLive:
		errs = append(errs, errors.New("STRIPE_MODE must be sandbox or live"))
	}

	key := b.Key.Reveal()
	switch {
	case key == "":
		errs = append(errs, errors.New(b.Var("STRIPE_KEY")+" is not set"))
	case !strings.HasPrefix(key, "sk_"+b.keyMode()+"_") && !strings.HasPrefix(key, "rk_"+b.keyMode()+"_"):
		// A publishable key (pk_) or a key of the other mode must never reach Stripe calls.
		errs = append(errs, errors.New(b.Var("STRIPE_KEY")+" must be a secret or restricted "+b.Mode+" key (sk_"+
			b.keyMode()+"_ or rk_"+b.keyMode()+"_)"))
	}

	if b.WebhookSecret.IsZero() {
		errs = append(errs, errors.New(b.Var("STRIPE_WEBHOOK_SECRET")+" is not set"))
	}

	for _, p := range b.Plans {
		if p.Price == "" {
			errs = append(errs, errors.New(b.Var("STRIPE_PRICE_"+strings.ToUpper(p.Name))+" is not set"))
		}
	}

	return errors.Join(errs...)
}

// keyMode is the mode word inside Stripe key prefixes: live keys are sk_live_/rk_live_, sandbox keys sk_test_/rk_test_.
func (b Billing) keyMode() string {
	if b.Mode == ModeLive {
		return "live"
	}

	return "test"
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
