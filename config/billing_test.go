package config_test

import (
	"maps"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/config"
)

type lookupEnv map[string]string

func (e lookupEnv) Lookup(k string) (string, bool) { v, ok := e[k]; return v, ok }

func TestPlanForPrice(t *testing.T) {
	t.Parallel()

	b := config.LoadBilling(
		lookupEnv{
			"STRIPE_PRICE_HOBBY_SANDBOX": "price_h",
			"STRIPE_PRICE_PRO_SANDBOX":   "price_p",
			"STRIPE_PRICE_BUSINESS":      "price_old",
		},
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

// both holds a full sandbox and a full live set side by side, as the Railway environments do.
var both = lookupEnv{
	"STRIPE_KEY_SANDBOX": "rk_test_s", "STRIPE_WEBHOOK_SECRET_SANDBOX": "whsec_s",
	"STRIPE_PRICE_HOBBY_SANDBOX": "price_hs", "STRIPE_PRICE_PRO_SANDBOX": "price_ps",
	"STRIPE_KEY_LIVE": "rk_live_l", "STRIPE_WEBHOOK_SECRET_LIVE": "whsec_l",
	"STRIPE_PRICE_HOBBY_LIVE": "price_hl", "STRIPE_PRICE_PRO_LIVE": "price_pl",
}

func with(env lookupEnv, kv ...string) lookupEnv {
	out := lookupEnv{}
	maps.Copy(out, env)

	for i := 0; i+1 < len(kv); i += 2 {
		out[kv[i]] = kv[i+1]
	}

	return out
}

func TestBillingModeSelectsTheVariableSet(t *testing.T) {
	t.Parallel()

	// Unset STRIPE_MODE reads the sandbox set (development and tests); production refuses it in Validate.
	if b := config.LoadBilling(both); b.Mode != "sandbox" || b.Key.Reveal() != "rk_test_s" {
		t.Errorf("unset mode = %q key %q, want sandbox rk_test_s", b.Mode, b.Key.Reveal())
	}

	for _, tc := range []struct {
		env              lookupEnv
		mode, key, price string
	}{
		{with(both, "STRIPE_MODE", "sandbox"), "sandbox", "rk_test_s", "price_hs"},
		{with(both, "STRIPE_MODE", " LIVE "), "live", "rk_live_l", "price_hl"},
	} {
		b := config.LoadBilling(tc.env)
		if b.Mode != tc.mode || b.Key.Reveal() != tc.key || b.Plans[0].Price != tc.price {
			t.Errorf(
				"mode %q key %q price %q; want %q %q %q",
				b.Mode,
				b.Key.Reveal(),
				b.Plans[0].Price,
				tc.mode,
				tc.key,
				tc.price,
			)
		}

		if err := b.Validate(); err != nil {
			t.Errorf("Validate(%s) = %v", tc.mode, err)
		}
	}
}

// TestSecurity_A05_StripeModeAndKeyMustMatch: production never charges in the wrong mode by mistake: an unknown
// mode, a key of the other mode or a publishable key refuses to boot.
func TestSecurity_A05_StripeModeAndKeyMustMatch(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		env  lookupEnv
		want string
	}{
		"mode unset":         {both, "STRIPE_MODE"},
		"unknown mode":       {with(both, "STRIPE_MODE", "prod"), "STRIPE_MODE"},
		"test key in live":   {with(both, "STRIPE_MODE", "live", "STRIPE_KEY_LIVE", "rk_test_x"), "STRIPE_KEY_LIVE"},
		"live key in test":   {with(both, "STRIPE_MODE", "sandbox", "STRIPE_KEY_SANDBOX", "sk_live_x"), "STRIPE_KEY_SANDBOX"},
		"publishable key":    {with(both, "STRIPE_MODE", "live", "STRIPE_KEY_LIVE", "pk_live_x"), "STRIPE_KEY_LIVE"},
		"missing live price": {with(both, "STRIPE_MODE", "live", "STRIPE_PRICE_PRO_LIVE", ""), "STRIPE_PRICE_PRO_LIVE"},
	} {
		err := config.LoadBilling(tc.env).Validate()
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: Validate = %v, want an error naming %s", name, err, tc.want)
		}
	}
}
