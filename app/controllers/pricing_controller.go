package controllers

import (
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/config"
)

type pricingPlan struct {
	ID          string         `json:"id"`
	Amounts     map[string]int `json:"amounts"`
	Quota       int            `json:"quota"`
	Lifetime    bool           `json:"lifetime"`
	MaxPageSize int            `json:"maxPageSize"`
	RPS         int            `json:"rps"`
}

// currencyForLocale is the default currency of a locale before the visitor picks one.
func currencyForLocale(locale string) string {
	switch locale {
	case "pt-BR":
		return "brl"
	case "es":
		return "eur"
	}

	return "usd"
}

// Pricing renders the public plan table (SDD 05 §6.5, SDD 09 §6.2): limits come from api.PlanByID so the
// page cannot drift from the quota the API enforces, amounts from the display catalogue. The currency in
// the query is validated; anything else falls back to the locale default.
func Pricing(c *web.Ctx) error {
	currency, ok := config.Currency(c.Query("currency"))
	if !ok {
		currency = currencyForLocale(i18n.Locale(c.Context()))
	}

	amounts := map[string]map[string]int{api.PlanFree: {}}
	for _, p := range config.Catalogue() {
		amounts[p.Name] = p.Amounts
	}

	plans := make([]pricingPlan, 0, 3)

	for _, id := range []string{api.PlanFree, api.PlanPro, api.PlanBusiness} {
		p, _ := api.PlanByID(id)
		plans = append(plans, pricingPlan{
			ID: id, Amounts: amounts[id], Quota: p.Quota, Lifetime: p.Lifetime, MaxPageSize: p.MaxPageSize, RPS: p.RPS,
		})
	}

	return inertia.Render(c, "Pricing/Index", inertia.Props{
		"plans":      plans,
		"currency":   currency,
		"currencies": config.Currencies,
		// Percentages are published after qa/sql/api_coverage.sql runs on a production copy (open issue 05-11).
		"coverage": map[string]any{"measured": false, "docsPath": "/docs/coverage"},
	})
}
