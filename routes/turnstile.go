package routes

import (
	"errors"
	"net/http"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/challenge"
	"github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/web"
)

// TurnstileVerifyURLEnv points siteverify at a fake in tests and local e2e; production refuses to boot with it.
const TurnstileVerifyURLEnv = "TURNSTILE_VERIFY_URL"

// Turnstile is the Cloudflare Turnstile setup (https://vukem.dev/docs/security/bot-challenge): off until both
// TURNSTILE_SITE_KEY and TURNSTILE_SECRET_KEY are set.
type Turnstile struct {
	challenge.TurnstileConfig

	verifyURL string
}

// LoadTurnstile reads TURNSTILE_* and refuses a half configuration or a fake siteverify in production.
func LoadTurnstile(app *vukem.App) (Turnstile, error) {
	cfg, err := config.Load[challenge.TurnstileConfig](app.Source)
	if err != nil {
		return Turnstile{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Turnstile{}, err
	}

	t := Turnstile{TurnstileConfig: cfg, verifyURL: app.Source.Get(TurnstileVerifyURLEnv)}
	if err := checkVerifyURL(t.verifyURL, app.Config.IsProduction()); err != nil {
		return Turnstile{}, err
	}

	return t, nil
}

// checkVerifyURL keeps the test siteverify out of production: a fake there would accept any token.
func checkVerifyURL(u string, production bool) error {
	if u != "" && production {
		return errors.New(TurnstileVerifyURLEnv + " is for tests only and refuses production")
	}

	return nil
}

// verifier is the siteverify client for action, nil while Turnstile is off.
func (t Turnstile) verifier(action string) *challenge.SiteVerify {
	v := t.Verifier(action)
	if v != nil && t.verifyURL != "" {
		// httpclient.Safe refuses loopback, where the test fake listens.
		v.URL, v.Client = t.verifyURL, &http.Client{Timeout: 5 * time.Second}
	}

	return v
}

// protect checks the X-Challenge-Token of a form POST against action; a pass-through while Turnstile is off.
func (t Turnstile) protect(action string) func(http.Handler) http.Handler {
	if !t.Enabled() {
		return func(next http.Handler) http.Handler { return next }
	}

	return challenge.Middleware(t.verifier(action), nil)
}

// stepUp is the scrape guard's interstitial, served at POST /turnstile because /challenge is the app's
// proof-of-work route. key signs the clearance cookie.
func (t Turnstile) stepUp(key []byte, skip func(*web.Ctx) bool) *challenge.StepUp {
	s := challenge.NewStepUp(t.TurnstileConfig, key)
	s.Path = "/turnstile"
	s.Skip = skip

	if v := t.verifier(s.Action); v != nil {
		s.Verifier = v
	}

	return s
}
