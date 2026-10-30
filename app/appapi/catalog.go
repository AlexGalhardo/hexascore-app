package appapi

import (
	vapi "github.com/AlexGalhardo/vukem/api"
	"github.com/AlexGalhardo/vukem/router"

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
