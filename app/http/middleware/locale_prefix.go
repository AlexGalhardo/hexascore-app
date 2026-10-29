// Package middleware holds the app's HTTP middleware.
package middleware

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

type originalURLKey struct{}

// LocalePrefixConfig configures LocalePrefix.
type LocalePrefixConfig struct {
	// Prefixes maps a lower-case first path segment to a locale, for
	// example {"pt-br": "pt-BR"}. The default locale has no prefix.
	Prefixes map[string]string
	// Query is the reserved query parameter handed to i18n.Middleware
	// (MiddlewareConfig.Query). A client-supplied value is always dropped.
	Query string
	// Cookie names the cookie holding a visitor's preferred locale.
	Cookie string
}

// LocalePrefix is the workaround(vukem) VF-024, to drop when v0.33.0 ships
// i18n.PathPrefixes: vukem's i18n.Middleware has no
// path-prefix source. Install it with Router.UseServer, so it runs before
// routing and before i18n.Middleware: it strips "/pt-br" from the path and
// passes the locale as cfg.Query, which i18n.Middleware reads
// (Order: SourceQuery). A GET/HEAD without a prefix from a visitor whose
// cookie names a prefixed locale gets a 307 to the prefixed URL.
func LocalePrefix(cfg LocalePrefixConfig) func(http.Handler) http.Handler {
	byLocale := make(map[string]string, len(cfg.Prefixes))
	for prefix, locale := range cfg.Prefixes {
		byLocale[locale] = prefix
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			q.Del(cfg.Query)

			first, rest, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/"), "/")
			if locale, ok := cfg.Prefixes[strings.ToLower(first)]; ok {
				q.Set(cfg.Query, locale)

				r2 := r.Clone(context.WithValue(r.Context(), originalURLKey{}, r.URL))
				r2.URL.Path = "/" + rest
				r2.URL.RawPath = ""
				r2.URL.RawQuery = q.Encode()
				next.ServeHTTP(w, r2)

				return
			}

			if prefix, ok := redirectPrefix(r, cfg, byLocale); ok {
				target := url.URL{Path: "/" + prefix + r.URL.Path, RawQuery: r.URL.RawQuery}
				w.Header().Add("Vary", "Cookie")
				http.Redirect(w, r, target.String(), http.StatusTemporaryRedirect)

				return
			}

			r2 := r.Clone(r.Context())
			r2.URL.RawQuery = q.Encode()
			next.ServeHTTP(w, r2)
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

// RestoreURL puts the original, prefixed URL back on the request after
// routing, so Inertia's page object url and redirects keep the prefix.
// Install it first in Router.Use (route-level), before the Inertia
// middleware; without a stripped prefix it does nothing.
func RestoreURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, ok := r.Context().Value(originalURLKey{}).(*url.URL); ok {
			r2 := r.Clone(r.Context())
			*r2.URL = *u
			r = r2
		}

		next.ServeHTTP(w, r)
	})
}
