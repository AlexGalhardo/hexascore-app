package tests

import (
	"regexp"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

var cspNonceRe = regexp.MustCompile(`'nonce-([^']+)'`)

// The root view renders the analytics and verification tags with the request's CSP nonce, and the CSP gains the
// provider origins next to the app's own sources (https://vukem.dev/docs/digging-deeper/analytics).
func TestAnalyticsHeadTags(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"UMAMI_WEBSITE_ID":         "94db1cb1-74f4-4a40-ad6c-962362670409",
		"UMAMI_DOMAINS":            "hexascore.app,www.hexascore.app",
		"GA_MEASUREMENT_ID":        "G-ABC123XYZ",
		"GOOGLE_SITE_VERIFICATION": "google-token-1",
		"BING_SITE_VERIFICATION":   "bing-token-1",
	}})
	res := app.Client(t).Get("/terms").AssertOK()

	csp := res.Header("Content-Security-Policy")
	m := cspNonceRe.FindStringSubmatch(csp)
	if m == nil {
		t.Fatalf("CSP without nonce: %s", csp)
	}

	for _, want := range []string{"https://www.googletagmanager.com", "'strict-dynamic'", "https://checkout.stripe.com"} {
		if !strings.Contains(csp, want) {
			t.Errorf("CSP lacks %s: %s", want, csp)
		}
	}

	body := res.Body()
	for _, want := range []string{
		`data-website-id="94db1cb1-74f4-4a40-ad6c-962362670409"`,
		`data-domains="hexascore.app,www.hexascore.app"`,
		`G-ABC123XYZ`,
		`name="google-site-verification" content="google-token-1"`,
		`name="msvalidate.01" content="bing-token-1"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("page lacks %s", want)
		}
	}

	head, _, _ := strings.Cut(body, "</head>")
	if strings.Count(head, `nonce="`+m[1]+`"`) < 3 {
		t.Errorf("analytics scripts do not carry the request nonce %s", m[1])
	}

	if regexp.MustCompile(`<script[^>]+src="https://www.googletagmanager.com`).MatchString(head) {
		t.Error("gtag.js is loaded before consent")
	}

	if strings.Index(head, "data-website-id") > strings.Index(head, "resources/js/app.tsx") &&
		strings.Contains(head, "resources/js/app.tsx") {
		t.Error("analytics tags come after the app scripts")
	}
}

// With every variable unset nothing is rendered and the CSP gains nothing.
func TestAnalyticsOffByDefault(t *testing.T) {
	t.Parallel()

	res := vukemtest.New(t, bootstrap.New).Client(t).Get("/terms").AssertOK()
	if strings.Contains(res.Header("Content-Security-Policy"), "googletagmanager") ||
		strings.Contains(res.Body(), "vukemAnalytics") || strings.Contains(res.Body(), "data-website-id") {
		t.Error("analytics rendered with no configuration")
	}
}
