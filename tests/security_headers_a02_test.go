package tests

import (
	"net/http"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestSecurity_A02_HexaHeadersWeb(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{})
	res := app.Client(t).Get("/")
	csp := res.Header("Content-Security-Policy")

	if !strings.Contains(csp, "'strict-dynamic'") || !strings.Contains(csp, "'nonce-") {
		t.Errorf("web CSP = %q, want a nonce and strict-dynamic", csp)
	}

	if strings.Contains(csp, "unsafe-inline") || strings.Contains(csp, "unsafe-eval") {
		t.Errorf("web CSP allows an unsafe source: %s", csp)
	}

	if !strings.Contains(csp, "report-uri /csp-reports") {
		t.Errorf("web CSP = %q, want report-uri /csp-reports", csp)
	}

	if got := res.Header("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
		t.Errorf("Referrer-Policy = %q", got)
	}

	if got := res.Header("Cross-Origin-Resource-Policy"); got != "same-site" {
		t.Errorf("CORP = %q, want same-site", got)
	}

	for _, feature := range []string{
		"camera=()", "microphone=()", "geolocation=()", "payment=()",
		"publickey-credentials-get=(self)", "publickey-credentials-create=(self)",
	} {
		if !strings.Contains(res.Header("Permissions-Policy"), feature) {
			t.Errorf("Permissions-Policy %q lacks %s", res.Header("Permissions-Policy"), feature)
		}
	}

	if strings.Contains(res.Header("Strict-Transport-Security"), "preload") {
		t.Errorf("HSTS must not preload: %s", res.Header("Strict-Transport-Security"))
	}
}

func TestSecurity_A02_HexaHeadersAPI(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{})
	res := app.Client(t).Get("/api/v1/games")
	csp := res.Header("Content-Security-Policy")
	want := "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'"

	if csp != want {
		t.Errorf("API CSP = %q, want %q", csp, want)
	}

	if got := res.Header("Referrer-Policy"); got != "no-referrer" {
		t.Errorf("API Referrer-Policy = %q", got)
	}

	if got := res.Header("Cross-Origin-Resource-Policy"); got != "cross-origin" {
		t.Errorf("API CORP = %q", got)
	}

	if got := res.Header("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("API CORS origin = %q", got)
	}

	if got := res.Header("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("API CORS credentials = %q, want none", got)
	}
}

func TestSecurity_A02_HexaCSPHasStrictDynamic(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{})
	csp := app.Client(t).Get("/").Header("Content-Security-Policy")
	script := ""

	for part := range strings.SplitSeq(csp, ";") {
		if p := strings.TrimSpace(part); strings.HasPrefix(p, "script-src ") {
			script = p
		}
	}

	if !strings.Contains(script, "'strict-dynamic'") || !strings.Contains(script, "'nonce-") {
		t.Errorf("script-src = %q, want nonce and strict-dynamic", script)
	}
}

func TestSecurity_A02_HexaCSPReports(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{})
	body := strings.NewReader(`{"csp-report":{"document-uri":"http://127.0.0.1/","violated-directive":"script-src"}}`)
	res := app.Client(t).Do(http.MethodPost, "/csp-reports", body, "application/csp-report")

	if res.Status() != http.StatusNoContent {
		t.Fatalf("csp report status = %d, body %s", res.Status(), res.Body())
	}

	for range 29 {
		app.Client(t).Do(http.MethodPost, "/csp-reports", strings.NewReader("{}"), "application/csp-report")
	}

	over := app.Client(t).Do(http.MethodPost, "/csp-reports", strings.NewReader("{}"), "application/csp-report")
	if over.Status() != http.StatusTooManyRequests {
		t.Fatalf("over budget status = %d, want 429", over.Status())
	}
}
