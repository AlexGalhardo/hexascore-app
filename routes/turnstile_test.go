package routes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/web"
)

// Cloudflare's published test keys (always pass); no real secret in tests.
var testTurnstile = func() Turnstile {
	var t Turnstile
	t.SiteKey, t.SecretKey = "1x00000000000000000000AA", config.NewSecret("1x0000000000000000000000000000000AA")

	return t
}()

// A fake siteverify accepts any token, so production must never boot with one.
func TestSecurity_A05_TurnstileFakeVerifyRefusedInProduction(t *testing.T) {
	t.Parallel()

	if err := checkVerifyURL("http://127.0.0.1:9/siteverify", true); err == nil {
		t.Error("production accepted TURNSTILE_VERIFY_URL")
	}

	for _, c := range []struct {
		url        string
		production bool
	}{{"", true}, {"", false}, {"http://127.0.0.1:9/siteverify", false}} {
		if err := checkVerifyURL(c.url, c.production); err != nil {
			t.Errorf("%+v: %v", c, err)
		}
	}
}

// The step-up interstitial posts to /turnstile (the app owns /challenge for proof of work) and is never cached or
// indexed; with Turnstile off it never challenges.
func TestTurnstileStepUpPath(t *testing.T) {
	t.Parallel()

	key := bytes.Repeat([]byte{3}, 32)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/games/zelda", nil)

	if testTurnstile.stepUp(key, nil).Require(web.NewCtx(rec, req, &web.Services{})) {
		t.Fatal("an uncleared client passed the step-up")
	}

	body := rec.Body.String()
	if rec.Code != http.StatusForbidden || !strings.Contains(body, `action="/turnstile"`) {
		t.Fatalf("status %d, interstitial:\n%s", rec.Code, body)
	}

	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("Cache-Control = %q", rec.Header().Get("Cache-Control"))
	}

	skip := func(*web.Ctx) bool { return true }
	if !testTurnstile.stepUp(key, skip).Require(web.NewCtx(httptest.NewRecorder(), req, &web.Services{})) {
		t.Error("a skipped client (verified crawler) was challenged")
	}

	if !(Turnstile{}).stepUp(key, nil).Require(web.NewCtx(httptest.NewRecorder(), req, &web.Services{})) {
		t.Error("Turnstile off still challenged")
	}
}
