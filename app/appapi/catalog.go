// Package appapi is the /app/v1 HTTP surface.
package appapi

import (
	"net/http"
	"strings"

	vapi "github.com/AlexGalhardo/vukem/api"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/controllers/publicapi"
)

// Catalog registers the site-audience catalogue reads on an /app/v1 group.
// Trending is the same snapshot as /api/v1 (SDD 08 section 6); item activity is business-only and stays off this audience.
func Catalog(g *router.Group, items *publicapi.Controller) {
	read := func(name string) vapi.Op {
		return vapi.Op{
			Name: name, Tags: []string{"catalog"}, Security: []string{"bearer"}, ValidateResponse: true,
		}
	}

	vapi.Register(g, "GET /books/{id}/editions", read("listEditions"), items.ListEditions)
	vapi.Register(g, "GET /entities/{id}/items", read("listEntityItems"), items.ListEntityItems)
	vapi.Register(g, "GET /entities/{id}", read("getEntity"), items.GetEntity)
	// Literal sections: "/{type}/genres" overlaps "/entities/{id}" on ServeMux.
	for section, op := range map[string]string{
		"games": "listGameGenres", "movies": "listMovieGenres",
		"series": "listSeriesGenres", "books": "listBookGenres",
	} {
		vapi.Register(g, "GET /"+section+"/genres", read(op), items.ListGenres(section))
	}
	for section, op := range map[string]string{
		"games": "listGameTrending", "movies": "listMovieTrending",
		"series": "listSeriesTrending", "books": "listBookTrending",
	} {
		vapi.Register(g, "GET /"+section+"/trending", read(op), items.ListTrending(section))
	}
	getItem := read("getItem")
	getItem.Description = "`id` is the item uuid or its slug (web App Links, SDD 10 FR-10-7). " +
		"A renamed slug or a merged id answers 301 to the live id."
	vapi.Register(g, "GET /{type}/{id}", getItem, items.GetItem)
	vapi.Register(g, "GET /{type}", read("listItems"), items.ListItems)
}

// KnownType rejects a first segment that is not a catalogue collection.
// GET /{type} would otherwise match /app/v1/nope and /app/v1/openapi.json and
// the group auth would answer 401. Those paths are not the catalogue: they
// are the same 404 the router uses for an unknown URL, before any credential
// check. Install it ahead of the auth stack on this group only.
func KnownType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch catalogSegment(r.URL.Path) {
		case "games", "movies", "series", "books", "entities":
			next.ServeHTTP(w, r)
		default:
			// web.Fail renders the HTML status page when Accept is empty.
			// Schemathesis omits Accept, and these 404s are documented as
			// application/problem+json.
			nr := r.Clone(r.Context())
			nr.Header.Set("Accept", "application/problem+json")
			web.Fail(w, nr, errs.ErrNotFound)
		}
	})
}

func catalogSegment(path string) string {
	rest, ok := strings.CutPrefix(path, "/app/v1/")
	if !ok {
		return ""
	}

	seg, _, _ := strings.Cut(rest, "/")

	return seg
}
