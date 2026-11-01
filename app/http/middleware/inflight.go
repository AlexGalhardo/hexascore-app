package middleware

import (
	"net/http"
	"regexp"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
)

// itemPath matches an item page, with or without a locale prefix: /movies/dune, /pt-BR/games/celeste.
var itemPath = regexp.MustCompile(`^(?:/[a-z]{2}(?:-[A-Za-z]{2})?)?/(?:games|movies|series|books)/[^/]+/?$`)

// Inflight feeds the ingestion pressure gate: it counts the requests being served and times the item pages
// (SDD 06 §7.5). It is a server middleware, so it sees every request and adds no allocation beyond the counter.
func Inflight(p *ingest.Pressure) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			finish := p.Begin()
			defer func() { finish(r.Method == http.MethodGet && itemPath.MatchString(r.URL.Path)) }()

			next.ServeHTTP(w, r)
		})
	}
}
