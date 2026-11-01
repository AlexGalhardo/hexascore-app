package appapi

import (
	vapi "github.com/AlexGalhardo/vukem/api"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/controllers"
)

// SearchInput documents the /app/v1/search query. The controller parses the raw query (searchParams), so these
// fields exist for the OpenAPI document and the generated client; strings keep a bad page a default, not a 422.
type SearchInput struct {
	Q      string `json:"q"      doc:"search text, at most 100 characters"`
	Type   string `json:"type"   doc:"comma-separated item types: game, movie, series, book"`
	Locale string `json:"locale" doc:"en, pt-BR or es; empty searches every locale; absent means the interface locale"`
	Page   string `json:"page"   doc:"1-based page, default 1"`
}

// SuggestInput documents the /app/v1/search/suggest query.
type SuggestInput struct {
	Q      string `json:"q"      doc:"search prefix, at most 100 characters"`
	Type   string `json:"type"   doc:"comma-separated kinds: game, movie, series, book, person, organization, franchise"`
	Locale string `json:"locale" doc:"en, pt-BR or es; empty searches every locale; absent means the interface locale"`
}

// Search registers the site search routes. The caller puts optional auth on
// the group: a missing token is anonymous, and q longer than 100 runes is 400
// from the searcher (SDD 07, SDD 10). suggest, clicks and search are the
// per-route IP limits.
func Search(
	g *router.Group, h *controllers.Search, suggest, clicks, search router.Middleware,
) {
	read := func(name string) vapi.Op {
		return vapi.Op{Name: name, Tags: []string{"search"}, ValidateResponse: true}
	}

	vapi.Register(g, "GET /search/suggest", read("search.suggest"),
		func(c *web.Ctx, _ SuggestInput) (controllers.SuggestJSON, error) { return h.SuggestRead(c) }).Use(suggest)
	g.Post("/search/events", h.Event).Name("search.events").Use(clicks)
	vapi.Register(g, "GET /search", read("search.index"),
		func(c *web.Ctx, _ SearchInput) (controllers.AppSearchJSON, error) { return h.AppSearch(c) }).Use(search)
}
