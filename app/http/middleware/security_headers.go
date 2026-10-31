package middleware

import (
	"net/http"
	"strings"

	vukemmw "github.com/AlexGalhardo/vukem/middleware"
)

// permissionsPolicy is the legacy HexaScore list (SDD 13 §6.2), wider than vukem's three-feature default.
const permissionsPolicy = "accelerometer=(), autoplay=(), camera=(), display-capture=(), geolocation=(), " +
	"gyroscope=(), magnetometer=(), microphone=(), midi=(), payment=(), usb=(), browsing-topics=(), " +
	"publickey-credentials-get=(self), publickey-credentials-create=(self)"

// UpgradeInsecure reports whether the CSP should carry upgrade-insecure-requests: only in production and when
// the public URL is https, so the built container reached over plain HTTP (local QA gate) keeps working.
func UpgradeInsecure(production bool, appURL string) bool {
	return production && strings.HasPrefix(appURL, "https://")
}

// SecureHeaders returns the adjustment passed as MiddlewareOptions.SecureHeaders (SDD 09 §7.8, SDD 13 §6.2).
// vukem's default policy keeps its per-request nonce on scripts and styles; this adds strict-dynamic, the
// YouTube, Stripe and Google origins, and, when upgrade is requested, upgrade-insecure-requests (a plain-HTTP dev
// server must keep working). extraImgHosts lists additional image origins such as a cover CDN.
func SecureHeaders(upgradeInsecure bool, extraImgHosts ...string) func(*vukemmw.SecureHeadersConfig) {
	return func(cfg *vukemmw.SecureHeadersConfig) {
		// frame-src and worker-src are written out: through CSPSources they would start from script-src and
		// inherit its nonce and 'strict-dynamic'.
		cfg.CSP = vukemmw.DefaultCSP + "; frame-src https://www.youtube-nocookie.com; worker-src 'self'"
		if upgradeInsecure {
			cfg.CSP += "; upgrade-insecure-requests"
		}

		cfg.PermissionsPolicy = permissionsPolicy
		cfg.ReportURI = "/csp-reports"
		cfg.CSPSources = map[string][]string{
			"script-src": {"'strict-dynamic'"},
			"img-src":    extraImgHosts,
			"form-action": {
				"https://checkout.stripe.com", "https://billing.stripe.com", "https://accounts.google.com",
			},
		}
	}
}

// APICSP is the JSON API policy (SDD 13 §6.2): no scripts, frames, or forms.
const APICSP = "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'"

// APIHeaders replaces the web header set on /api/v1. SecureHeaders runs first and
// stamps every response; this runs later and overwrites only the API paths.
func APIHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1" || strings.HasPrefix(r.URL.Path, "/api/v1/") {
			h := w.Header()
			h.Set("Content-Security-Policy", APICSP)
			h.Set("Referrer-Policy", "no-referrer")
			h.Set("Cross-Origin-Resource-Policy", "cross-origin")
			h.Set("Access-Control-Allow-Origin", "*")
			h.Del("Access-Control-Allow-Credentials")
			h.Del("Reporting-Endpoints")
		}

		next.ServeHTTP(w, r)
	})
}

// CrossOriginResourcePolicy sets Cross-Origin-Resource-Policy, which vukem's SecureHeaders does not send.
// Install it with Router.UseServer.
func CrossOriginResourcePolicy(value string) func(http.Handler) http.Handler {
	value = strings.TrimSpace(value)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cross-Origin-Resource-Policy", value)
			next.ServeHTTP(w, r)
		})
	}
}
