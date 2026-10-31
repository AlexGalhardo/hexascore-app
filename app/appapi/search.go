package appapi

import (
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/controllers"
)

// Search registers the site search routes. The caller puts optional auth on
// the group: a missing token is anonymous, and q longer than 100 runes is 400
// from the searcher (SDD 07, SDD 10). suggest, clicks and search are the
// per-route IP limits.
func Search(
	g *router.Group, h *controllers.Search, suggest, clicks, search router.Middleware,
) {
	g.Get("/search/suggest", h.Suggest).Name("search.suggest").Use(suggest)
	g.Post("/search/events", h.Event).Name("search.events").Use(clicks)
	g.Get("/search", h.AppIndex).Name("search.index").Use(search)
}
