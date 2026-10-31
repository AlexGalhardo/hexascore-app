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
// Trending is not here: P3.A.1.4 waits on P6.C, and an empty list would look
// like a real ranking (ROADMAP P3.C.2.1).
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
	vapi.Register(g, "GET /{type}/{id}", read("getItem"), items.GetItem)
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
			web.Fail(w, r, errs.ErrNotFound)
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
