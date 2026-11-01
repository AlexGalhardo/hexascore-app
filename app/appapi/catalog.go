// Package appapi is the /app/v1 HTTP surface.
package appapi

import (
	"fmt"
	"maps"
	"net/http"
	"strings"

	vapi "github.com/AlexGalhardo/vukem/api"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/controllers/publicapi"
)

// AppItemJSON is one item on /app/v1: the /api/v1 item plus the cover renditions (SDD 10 §7.3). The paid API
// omits covers (SDD 05, licence), so they live only on this audience. Embedded fields flatten in the schema
// (https://vukem.dev/docs/rest-apis/openapi#go-types).
type AppItemJSON struct {
	publicapi.ItemJSON

	CoverThumbURL string `json:"coverThumbUrl,omitempty"`
	CoverURL      string `json:"coverUrl,omitempty"`
}

// AppItemPageJSON is one keyset page of app items.
type AppItemPageJSON struct {
	Data  []AppItemJSON `json:"data"`
	Meta  vapi.Meta     `json:"meta"`
	Links vapi.Links    `json:"links"`
}

// AppTrendingEntryJSON is one rank of the app trending read.
type AppTrendingEntryJSON struct {
	Rank       int         `json:"rank"       validate:"required"`
	Basis      string      `json:"basis"      validate:"required,oneof=trend prior"`
	TrendScore float64     `json:"trendScore"`
	Item       AppItemJSON `json:"item"`
}

// AppTrendingJSON is the app trending read; computedAt is null before the first snapshot.
type AppTrendingJSON struct {
	ComputedAt *string                `json:"computedAt"`
	Data       []AppTrendingEntryJSON `json:"data"`
}

// covered decorates publicapi items with their covers, made absolute with COVERS_BASE_URL.
type covered struct {
	items *publicapi.Controller
	base  string
}

func (h covered) decorate(c *web.Ctx, in []publicapi.ItemJSON) ([]AppItemJSON, error) {
	ids := make([]string, len(in))
	for i, it := range in {
		ids[i] = it.ID
	}

	covers, err := catalog.CoversFor(c.Context(), h.items.DB, ids)
	if err != nil {
		return nil, fmt.Errorf("app item covers: %w", err)
	}

	out := make([]AppItemJSON, len(in))
	for i, it := range in {
		out[i] = AppItemJSON{ItemJSON: it}
		if cv := covers[it.ID].WithBase(h.base); cv != nil {
			out[i].CoverThumbURL, out[i].CoverURL = cv.Thumb, cv.Full
		}
	}

	return out, nil
}

// unconditional runs read with the conditional headers hidden: the publicapi validators hash the item without
// covers, so they must not answer 304 for the app body. The caller then sets the ETag of what it sends.
func unconditional[T any](c *web.Ctx, read func() (T, error)) (T, error) {
	hdr := c.Request().Header
	saved := map[string][]string{}
	for _, k := range []string{"If-None-Match", "If-Modified-Since"} {
		if v, ok := hdr[k]; ok {
			saved[k] = v
			hdr.Del(k)
		}
	}
	v, err := read()
	maps.Copy(hdr, saved)

	return v, err
}

func (h covered) page(c *web.Ctx, read func() (publicapi.ItemPageJSON, error)) (AppItemPageJSON, error) {
	p, err := unconditional(c, read)
	if err != nil {
		return AppItemPageJSON{}, err
	}

	data, err := h.decorate(c, p.Data)
	if err != nil {
		return AppItemPageJSON{}, err
	}

	out := AppItemPageJSON{Data: data, Meta: p.Meta, Links: p.Links}

	return out, publicapi.NotModified(c, out)
}

func (h covered) getItem(c *web.Ctx, in publicapi.GetItemInput) (AppItemJSON, error) {
	it, err := unconditional(c, func() (publicapi.ItemJSON, error) { return h.items.GetItem(c, in) })
	if err != nil {
		return AppItemJSON{}, err
	}

	out, err := h.decorate(c, []publicapi.ItemJSON{it})
	if err != nil {
		return AppItemJSON{}, err
	}

	return out[0], publicapi.NotModified(c, out[0])
}

func (h covered) listItems(c *web.Ctx, in publicapi.ListItemsInput) (AppItemPageJSON, error) {
	return h.page(c, func() (publicapi.ItemPageJSON, error) { return h.items.ListItems(c, in) })
}

func (h covered) listEntityItems(c *web.Ctx, in publicapi.ListEntityItemsInput) (AppItemPageJSON, error) {
	return h.page(c, func() (publicapi.ItemPageJSON, error) { return h.items.ListEntityItems(c, in) })
}

func (h covered) trending(section string) func(*web.Ctx, publicapi.TrendingInput) (AppTrendingJSON, error) {
	read := h.items.ListTrending(section)

	return func(c *web.Ctx, in publicapi.TrendingInput) (AppTrendingJSON, error) {
		tr, err := unconditional(c, func() (publicapi.TrendingJSON, error) { return read(c, in) })
		if err != nil {
			return AppTrendingJSON{}, err
		}

		items := make([]publicapi.ItemJSON, len(tr.Data))
		for i, e := range tr.Data {
			items[i] = e.Item
		}

		decorated, err := h.decorate(c, items)
		if err != nil {
			return AppTrendingJSON{}, err
		}

		out := AppTrendingJSON{ComputedAt: tr.ComputedAt, Data: make([]AppTrendingEntryJSON, len(tr.Data))}
		for i, e := range tr.Data {
			out.Data[i] = AppTrendingEntryJSON{
				Rank:       e.Rank,
				Basis:      e.Basis,
				TrendScore: e.TrendScore,
				Item:       decorated[i],
			}
		}

		return out, publicapi.NotModified(c, out)
	}
}

// Catalog registers the site-audience catalogue reads on an /app/v1 group. coversBase is the normalised
// COVERS_BASE_URL. Trending is the same snapshot as /api/v1 (SDD 08 section 6); item activity is business-only
// and stays off this audience.
func Catalog(g *router.Group, items *publicapi.Controller, coversBase string) {
	cov := covered{items: items, base: coversBase}
	read := func(name string) vapi.Op {
		return vapi.Op{
			Name: name, Tags: []string{"catalog"}, Security: []string{"bearer"}, ValidateResponse: true,
		}
	}

	vapi.Register(g, "GET /books/{id}/editions", read("listEditions"), items.ListEditions)
	vapi.Register(g, "GET /entities/{id}/items", read("listEntityItems"), cov.listEntityItems)
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
		vapi.Register(g, "GET /"+section+"/trending", read(op), cov.trending(section))
	}
	vapi.Register(g, "GET /{type}/{id}", read("getItem"), cov.getItem)
	vapi.Register(g, "GET /{type}", read("listItems"), cov.listItems)
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
