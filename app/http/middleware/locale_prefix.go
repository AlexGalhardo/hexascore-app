// Package middleware holds the app's HTTP middleware.
package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/AlexGalhardo/vukem/i18n"
)

// LocalePrefixConfig configures LocaleCookieRedirect.
type LocalePrefixConfig struct {
	// Prefixes maps a first path segment to a locale, as in
	// i18n.MiddlewareConfig.PathPrefixes ({"pt-br": "pt-BR"}).
	Prefixes map[string]string
	// Cookie names the cookie holding a visitor's preferred locale.
	Cookie string
}

// LocaleCookieRedirect sends a GET/HEAD page navigation without a locale
// prefix from a visitor whose cookie names a prefixed locale to the prefixed
// URL with a 307 (SDD 09 §7.1); vukem has no cookie-driven redirect. It also
// puts the visitor's prefix on same-site redirect targets so a form post does
// not drop the locale. Install it with Router.UseServer after
// i18n.StripPathPrefix (vukem.DefaultMiddlewareWith).
func LocaleCookieRedirect(cfg LocalePrefixConfig) func(http.Handler) http.Handler {
	byLocale := make(map[string]string, len(cfg.Prefixes))
	for prefix, locale := range cfg.Prefixes {
		byLocale[locale] = "/" + prefix
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if i18n.PathPrefix(r.Context()) != "" {
				next.ServeHTTP(&localizedRedirects{ResponseWriter: w, r: r}, r)

				return
			}

			if prefix, ok := redirectPrefix(r, cfg, byLocale); ok {
				target := url.URL{Path: prefix + r.URL.Path, RawQuery: r.URL.RawQuery}
				w.Header().Add("Vary", "Cookie")
				http.Redirect(w, r, target.String(), http.StatusTemporaryRedirect)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// redirectPrefix returns the prefix to redirect to: only page navigations
// (not Inertia visits, assets, health checks or the API) are redirected.
func redirectPrefix(r *http.Request, cfg LocalePrefixConfig, byLocale map[string]string) (string, bool) {
	if (r.Method != http.MethodGet && r.Method != http.MethodHead) || r.Header.Get("X-Inertia") != "" {
		return "", false
	}

	p := r.URL.Path
	if p == "/up" || strings.HasPrefix(p, "/api/") || strings.HasPrefix(p, "/build/") ||
		strings.Contains(p[strings.LastIndex(p, "/"):], ".") {
		return "", false
	}

	c, err := r.Cookie(cfg.Cookie)
	if err != nil {
		return "", false
	}

	prefix, ok := byLocale[c.Value]

	return prefix, ok
}

// localizedRedirects prefixes the Location of a redirect to an absolute
// in-site page path when the request came under a locale prefix.
type localizedRedirects struct {
	http.ResponseWriter
	r *http.Request
}

func (l *localizedRedirects) WriteHeader(code int) {
	if code >= 300 && code < 400 {
		h := l.Header()
		if loc := h.Get("Location"); strings.HasPrefix(loc, "/") && !strings.HasPrefix(loc, "//") && localizable(loc) {
			h.Set("Location", i18n.LocalizePath(l.r.Context(), loc))
		}
	}

	l.ResponseWriter.WriteHeader(code)
}

func (l *localizedRedirects) Unwrap() http.ResponseWriter { return l.ResponseWriter }

func localizable(loc string) bool {
	return loc != "/up" && !strings.HasPrefix(loc, "/api/") && !strings.HasPrefix(loc, "/build/")
}
