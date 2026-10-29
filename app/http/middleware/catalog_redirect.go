package middleware

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

type catalogSection struct {
	scope     catalog.Scope
	namespace string
}

// catalogSections maps a web path section to its redirect scope and namespace (SDD 01 §6.2).
var catalogSections = map[string]catalogSection{
	"games":      {catalog.ScopeItem, "game"},
	"movies":     {catalog.ScopeItem, "movie"},
	"series":     {catalog.ScopeItem, "series"},
	"books":      {catalog.ScopeItem, "book"},
	"people":     {catalog.ScopeEntity, "person"},
	"companies":  {catalog.ScopeEntity, "organization"},
	"franchises": {catalog.ScopeEntity, "franchise"},
	"platforms":  {catalog.ScopeEntity, "platform"},
	"genres":     {catalog.ScopeEntity, "genre"},
	"game-modes": {catalog.ScopeEntity, "game_mode"},
}

// CatalogRedirect answers 301 for `/{section}/{old-slug-or-merged-id}` with
// the live slug (SDD 01 §7.1). A live slug can never also be a redirect
// (invariant I-1), so one indexed lookup in catalog_redirects is enough and
// misses fall through to routing and its 404. Install it with Router.UseServer
// right after LocalePrefix: it sees the stripped path and keeps the prefix.
// A failed lookup is logged and the request continues: a redirect is a
// convenience, not an access check.
func CatalogRedirect(ex database.Executor, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path, ok := lookupRedirect(r, ex, log)
			if !ok {
				next.ServeHTTP(w, r)

				return
			}

			// The stripped request carries LocalePrefix's reserved query; the visitor's own is the original.
			query := r.URL.RawQuery
			prefix := ""

			if orig, ok := r.Context().Value(originalURLKey{}).(*url.URL); ok {
				query = orig.RawQuery
				prefix = strings.ToLower(strings.TrimSuffix(orig.Path, r.URL.Path))
			}

			loc := url.URL{Path: prefix + path, RawQuery: query}
			http.Redirect(w, r, loc.String(), http.StatusMovedPermanently)
		})
	}
}

func lookupRedirect(r *http.Request, ex database.Executor, log *slog.Logger) (string, bool) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return "", false
	}

	section, key, ok := strings.Cut(strings.TrimPrefix(r.URL.Path, "/"), "/")
	sec, known := catalogSections[section]

	if !ok || !known || key == "" || strings.Contains(key, "/") {
		return "", false
	}

	resolve := catalog.ResolveSlug
	if _, err := uuid.Parse(key); err == nil {
		resolve = catalog.ResolveID
	}

	t, found, err := resolve(r.Context(), ex, sec.scope, sec.namespace, key)
	if err != nil {
		log.ErrorContext(r.Context(), "catalog redirect lookup failed", "path", r.URL.Path, "error", err)

		return "", false
	}

	return "/" + section + "/" + t.Slug, found
}
