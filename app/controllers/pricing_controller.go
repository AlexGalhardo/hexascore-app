package controllers

import (
	"encoding/json"
	"errors"
	"fmt"

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

// APICoverage is resources/data/api_coverage.json: per (type, field) percentages from qa/sql/api_coverage.sql.
// MeasuredAt is nil until the first staging run; the pages then show the statement without numbers.
type APICoverage struct {
	MeasuredAt *string `json:"measuredAt"`
	Rows       []struct {
		Type  string  `json:"type"`
		Field string  `json:"field"`
		Items int     `json:"items"`
		Pct   float64 `json:"pct"`
	} `json:"rows"`
}

// ParseAPICoverage decodes the embedded coverage file; a broken file stops the boot instead of a wrong page.
func ParseAPICoverage(raw []byte) (APICoverage, error) {
	var c APICoverage
	if err := json.Unmarshal(raw, &c); err != nil {
		return c, fmt.Errorf("api coverage: %w", err)
	}
	if c.Rows == nil {
		return c, errors.New("api coverage: rows missing")
	}

	return c, nil
}

// Pricing renders the public plan table (SDD 05 §6.5, SDD 09 §6.2): limits come from api.PlanByID so the
// page cannot drift from the quota the API enforces, amounts from the display catalogue. The currency in
// the query is validated; anything else falls back to the locale default.
func Pricing(coverage APICoverage) web.HandlerFunc {
	return func(c *web.Ctx) error { return pricing(c, coverage) }
}

func pricing(c *web.Ctx, coverage APICoverage) error {
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
		"coverage": map[string]any{
			"measuredAt": coverage.MeasuredAt, "rows": coverage.Rows, "docsPath": "/docs/coverage",
		},
	})
}
