package middleware

import (
	"net/http"
	"strings"

	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
)

// PageNamespaces is the per-page namespace table of PageI18n: a namespace listed under a path prefix ships only
// on pages under that prefix. A namespace used outside the listed prefixes (a layout, Nav, Footer) must stay in
// the base list.
var PageNamespaces = map[string][]string{
	"/about":           {"about."},
	"/docs":            {"docs.", "coverage."},
	"/terms":           {"terms."},
	"/privacy":         {"policy."},
	"/help":            {"help."},
	"/how-score-works": {"help."},
	"/billing":         {"billing."},
	"/pricing":         {"billing.", "coverage."},
	"/profile/api":     {"api."},
	"/contact":         {"contact."},
	"/dashboard":       {"dashboard."},
}

// PageI18n replaces inertia.I18n (same locale, locales and messages props) with a page-scoped prefix list: base
// plus the PageNamespaces entry for the request path (locale prefixes ignored). The once-prop key carries the
// page group, because a once prop the client already holds is not sent again and would keep the previous
// page's smaller catalogue. localePrefixes maps URL prefixes ("pt-br") to locales (SDD 09 §6.1).
func PageI18n(base []string, localePrefixes map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if b := i18n.BundleFrom(ctx); b != nil {
				locale := i18n.Locale(ctx)
				if locale == "" {
					locale = b.Default()
				}

				group := i18nGroup(r.URL.Path, localePrefixes)
				prefixes := append(append([]string{}, base...), PageNamespaces[group]...)
				key := "i18n:" + locale
				if group != "" {
					key += ":" + strings.TrimPrefix(group, "/")
				}

				inertia.Share(ctx, "locale", locale)
				inertia.Share(ctx, "locales", b.Locales())
				inertia.Share(ctx, "messages", inertia.Once(b.Messages(locale, prefixes...)).As(key))
			}

			next.ServeHTTP(w, r)
		})
	}
}

func i18nGroup(path string, localePrefixes map[string]string) string {
	if seg, rest, ok := strings.Cut(strings.TrimPrefix(path, "/"), "/"); ok || seg != "" {
		if _, isLocale := localePrefixes[strings.ToLower(seg)]; isLocale {
			path = "/" + rest
		}
	}

	for key := range PageNamespaces {
		if path == key || strings.HasPrefix(path, key+"/") {
			return key
		}
	}

	return ""
}
