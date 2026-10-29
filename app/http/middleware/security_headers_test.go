package middleware_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	vukemmw "github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

var nonceRe = regexp.MustCompile(`'nonce-([A-Za-z0-9+/_=-]+)'`)

func directive(csp, name string) string {
	for part := range strings.SplitSeq(csp, ";") {
		if p := strings.TrimSpace(part); p == name || strings.HasPrefix(p, name+" ") {
			return p
		}
	}

	return ""
}

func TestSecurity_A05_HeadersOnHTMLAndJSON(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{})
	c := app.Client(t)

	for _, path := range []string{"/", "/up"} {
		res := c.Get(path)
		h := func(k string) string { return res.Header(k) }
		csp := h("Content-Security-Policy")

		scripts := directive(csp, "script-src")
		if !strings.Contains(scripts, "'strict-dynamic'") || !nonceRe.MatchString(scripts) {
			t.Errorf("%s: script-src = %q, want nonce and strict-dynamic", path, scripts)
		}

		if strings.Contains(csp, "unsafe-inline") || strings.Contains(csp, "unsafe-eval") {
			t.Errorf("%s: CSP allows unsafe sources: %s", path, csp)
		}

		for dir, want := range map[string]string{
			"object-src":      "'none'",
			"base-uri":        "'self'",
			"frame-ancestors": "'none'",
			"frame-src":       "https://www.youtube-nocookie.com",
			"form-action":     "https://checkout.stripe.com",
			"worker-src":      "'self'",
		} {
			if got := directive(csp, dir); !strings.Contains(got, want) {
				t.Errorf("%s: %s = %q, want %q", path, dir, got, want)
			}
		}

		if m := nonceRe.FindStringSubmatch(directive(csp, "style-src")); m == nil {
			t.Errorf("%s: style-src has no nonce: %s", path, csp)
		}

		want := map[string]string{
			"X-Content-Type-Options":       "nosniff",
			"X-Frame-Options":              "DENY",
			"Referrer-Policy":              "strict-origin-when-cross-origin",
			"Cross-Origin-Opener-Policy":   "same-origin",
			"Cross-Origin-Resource-Policy": "same-site",
		}
		for k, v := range want {
			if got := h(k); got != v {
				t.Errorf("%s: %s = %q, want %q", path, k, got, v)
			}
		}

		pp := h("Permissions-Policy")
		for _, f := range []string{"camera=()", "microphone=()", "geolocation=()", "payment=()", "browsing-topics=()"} {
			if !strings.Contains(pp, f) {
				t.Errorf("%s: Permissions-Policy %q lacks %s", path, pp, f)
			}
		}
	}
}

func TestSecurity_A05_NonceIsPerRequestAndInPage(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{})
	c := app.Client(t)

	nonce := func() (string, string) {
		res := c.Get("/")
		m := nonceRe.FindStringSubmatch(res.Header("Content-Security-Policy"))
		if m == nil {
			t.Fatal("no nonce in CSP")
		}

		return m[1], res.Body()
	}

	n1, body := nonce()
	n2, _ := nonce()

	if n1 == n2 {
		t.Errorf("nonce reused across requests: %s", n1)
	}

	if !strings.Contains(body, `nonce="`+n1+`"`) {
		t.Errorf("page does not carry the response nonce %q", n1)
	}
}

func TestSecurity_A05_HSTSOnHTTPSOnly(t *testing.T) {
	cfg := vukemmw.SecureHeadersConfig{}
	middleware.SecureHeaders(true)(&cfg)

	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(cfg.CSP, "upgrade-insecure-requests") {
		t.Errorf("production CSP lacks upgrade-insecure-requests: %s", cfg.CSP)
	}

	handler := vukemmw.SecureHeaders(cfg)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	plain := httptest.NewRecorder()
	handler.ServeHTTP(plain, httptest.NewRequest(http.MethodGet, "http://hexascore.test/", nil))

	if got := plain.Header().Get("Strict-Transport-Security"); got != "" {
		t.Errorf("HSTS over plain HTTP = %q, want none", got)
	}

	secure := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://hexascore.test/", nil)
	req.TLS = &tls.ConnectionState{}
	handler.ServeHTTP(secure, req)

	got := secure.Header().Get("Strict-Transport-Security")
	if got != "max-age=63072000; includeSubDomains" {
		t.Errorf("HSTS = %q, want two years with includeSubDomains and no preload", got)
	}
}

func TestSecurity_A05_DevPolicyStaysPlainHTTP(t *testing.T) {
	cfg := vukemmw.SecureHeadersConfig{}
	middleware.SecureHeaders(false)(&cfg)

	if strings.Contains(cfg.CSP, "upgrade-insecure-requests") {
		t.Errorf("dev CSP must not upgrade requests: %s", cfg.CSP)
	}
}
