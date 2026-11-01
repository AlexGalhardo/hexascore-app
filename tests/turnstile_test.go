//go:build integration

package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/routes"
)

// fakeSiteVerify answers like Cloudflare siteverify: a token "good-<action>" is valid for that action, anything
// else is refused. No request leaves the machine.
func fakeSiteVerify(t *testing.T) string {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		action, ok := strings.CutPrefix(r.PostForm.Get("response"), "good-")
		res := map[string]any{"success": ok, "action": action, "hostname": "example.test"}
		if !ok {
			res["error-codes"] = []string{"invalid-input-response"}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))
	t.Cleanup(srv.Close)

	return srv.URL
}

func turnstileApp(t *testing.T) *vukemtest.App {
	t.Helper()

	return vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL": pgtest.Database(t, ""),
		// Cloudflare's published test keys; siteverify is the local fake.
		"TURNSTILE_SITE_KEY":         "1x00000000000000000000AA",
		"TURNSTILE_SECRET_KEY":       "1x0000000000000000000000000000000AA",
		routes.TurnstileVerifyURLEnv: fakeSiteVerify(t),
	}})
}

// TestSecurity_A07_TurnstileProtectsForms: sign-up, login and contact posts need a Turnstile token solved for that
// form; a missing token, a bad one or one solved on another form is refused before the handler runs.
func TestSecurity_A07_TurnstileProtectsForms(t *testing.T) {
	app := turnstileApp(t)
	form := url.Values{"identifier": {"ana@example.test"}, "password": {"Demo-password-1"}}

	post := func(path, token string) *vukemtest.Response {
		c := app.Client(t).WithHeader("Sec-Fetch-Site", "same-origin").WithHeader("Accept", "application/json")
		if token != "" {
			c = c.WithHeader("X-Challenge-Token", token)
		}

		return c.PostForm(path, form)
	}

	for _, tc := range []struct{ path, action string }{
		{"/login", "login"}, {"/signup", "register"}, {"/contact", "contact"},
	} {
		t.Run(tc.action, func(t *testing.T) {
			for name, token := range map[string]string{"missing": "", "invalid": "forged", "other form": "good-other"} {
				res := post(tc.path, token).AssertUnprocessable()
				if !strings.Contains(res.Body(), "challenge") {
					t.Errorf("%s token: %s", name, res.Body())
				}
			}

			if res := post(tc.path, "good-"+tc.action); strings.Contains(res.Body(), `"challenge"`) {
				t.Errorf("valid %s token refused: %d %s", tc.action, res.Recorder().Code, res.Body())
			}
		})
	}
}

// The login page carries the Turnstile script with the request nonce, the CSP allows its frame, and the site key
// reaches the page props; pages without a protected form do not load the script.
func TestTurnstileOnLoginPage(t *testing.T) {
	app := turnstileApp(t)

	res := app.Client(t).Get("/login").AssertOK()
	csp := res.Header("Content-Security-Policy")

	m := cspNonceRe.FindStringSubmatch(csp)
	if m == nil ||
		!strings.Contains(csp, "frame-src https://www.youtube-nocookie.com https://challenges.cloudflare.com") {
		t.Fatalf("CSP: %s", csp)
	}

	body := res.Body()
	if !strings.Contains(
		body,
		`src="https://challenges.cloudflare.com/turnstile/v0/api.js" async defer nonce="`+m[1]+`"`,
	) {
		t.Error("login page lacks the nonce'd Turnstile script")
	}

	if !strings.Contains(body, "1x00000000000000000000AA") {
		t.Error("turnstileSiteKey prop missing")
	}

	if strings.Contains(app.Client(t).Get("/terms").AssertOK().Body(), "turnstile/v0/api.js") {
		t.Error("terms page loads Turnstile")
	}
}
