package config_test

import (
	"testing"

	"github.com/AlexGalhardo/hexascore-app/config"
)

type lookupEnv map[string]string

func (e lookupEnv) Lookup(k string) (string, bool) { v, ok := e[k]; return v, ok }

func TestPlanForPrice(t *testing.T) {
	t.Parallel()

	b := config.LoadBilling(
		lookupEnv{"STRIPE_PRICE_HOBBY": "price_h", "STRIPE_PRICE_PRO": "price_p", "STRIPE_PRICE_BUSINESS": "price_old"},
	)

	for price, want := range map[string]string{"price_h": "hobby", "price_p": "pro"} {
		if p, ok := b.PlanForPrice(price); !ok || p.Name != want {
			t.Errorf("PlanForPrice(%q) = %q, %v; want %q", price, p.Name, ok, want)
		}
	}

	// A retired business price id no longer maps to a plan.
	for _, price := range []string{"price_old", "", "price_unknown"} {
		if _, ok := b.PlanForPrice(price); ok {
			t.Errorf("PlanForPrice(%q) found a plan", price)
		}
	}

	if err := b.Validate(); err == nil || b.Plans[0].Name != "hobby" {
		t.Errorf("Validate without key/webhook = %v, first plan %q", err, b.Plans[0].Name)
	}
}
