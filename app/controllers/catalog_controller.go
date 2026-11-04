package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// CatalogPageSize is the default item count of one Catalog/Index page; ?per= picks 40, 80 or 120
// (owner 2026-11-02).
const CatalogPageSize = 40

// catalogPageSize reads ?per=; any value but the three offered sizes is the default.
func catalogPageSize(raw string) int {
	switch n, _ := strconv.Atoi(raw); n {
	case 40, 80, 120:
		return n
	default:
		return CatalogPageSize
	}
}

// catalogPage reads ?page=; anything but a positive integer is page 1, and a page past catalog.MaxPage is the last
// one allowed (an old link or a crawler gets an empty page, not a 400).
func catalogPage(raw string) int {
	if n, err := strconv.Atoi(raw); err == nil && n > 0 {
		return min(n, catalog.MaxPage)
	}

	return 1
}

// pageLinks are the prev/next page numbers of a numbered listing page ("" at either end).
func pageLinks(page int, hasMore bool) map[string]string {
	links := map[string]string{"prev": "", "next": ""}
	if page > 1 {
		links["prev"] = strconv.Itoa(page - 1)
	}
	if hasMore {
		links["next"] = strconv.Itoa(page + 1)
	}

	return links
}

// Catalog serves the section listings Catalog/Index.
type Catalog struct {
	app   *vukem.App
	codec *catalog.CursorCodec
}

// NewCatalog builds the handlers; the cursor key derives from APP_KEY.
func NewCatalog(app *vukem.App) (*Catalog, error) {
	key, err := app.DeriveKey("catalog.cursor", 32)
	if err != nil {
		return nil, err
	}
	codec, err := catalog.NewCursorCodec(key)
	if err != nil {
		return nil, err
	}

	return &Catalog{app: app, codec: codec}, nil
}

// Index renders Catalog/Index for the item type typ: one page of items plus the opaque signed cursors of the
// pages before and after it ("" at either end), which the page turns into buttons and rel=prev/next links.
// Facets are in the first paint.
func (h *Catalog) Index(typ string) web.HandlerFunc {
	return func(c *web.Ctx) error {
		db, err := h.app.DB()
		if err != nil {
			return err
		}

		params, err := catalog.ParseListQuery(c.Query, typ)
		params.Limit = catalogPageSize(c.Query("per"))
		if err != nil {
			return badListing(err)
		}
		// Numbered pages replace the opaque cursors on the web listings (owner 2026-11-02); an old ?cursor= link
		// lands on page 1.
		params.Cursor, params.Page = "", catalogPage(c.Query("page"))
		if params.Entities, err = entityFilters(c.Query); err != nil {
			return badListing(err)
		}
		if ff, ferr := h.app.Flags(); ferr == nil {
			params.Showcase = catalog.ShowcaseFrom(c.Context(), ff)
		}
		page, err := catalog.List(c.Context(), db, h.codec, params)
		if err != nil {
			return badListing(err)
		}

		ids := make([]string, len(page.Items))
		for i, it := range page.Items {
			ids[i] = it.ID
		}
		cards, err := catalog.Summaries(c.Context(), db, ids, i18n.Locale(c.Context()))
		if err != nil {
			return err
		}

		// Facets ride the first response. A follow-up reload used to widen the
		// selects after paint and shift the list (CLS on /games).
		genres, platforms, err := catalog.Facets(c.Context(), db, typ, i18n.Locale(c.Context()), nil)
		if err != nil {
			return err
		}
		facets := map[string]any{"genres": genres, "platforms": platforms}
		if typ == "movie" || typ == "series" {
			facets["watch"], err = catalog.WatchFacets(c.Context(), db, typ)
			if err != nil {
				return err
			}
		}
		if typ == "game" {
			if facets["modes"], err = catalog.FacetsOf(
				c.Context(),
				db,
				typ,
				"game_mode",
				i18n.Locale(c.Context()),
				nil,
			); err != nil {
				return err
			}
		}
		active, err := catalog.ActiveEntityFilters(c.Context(), db, params.Entities, i18n.Locale(c.Context()))
		if err != nil {
			return err
		}

		props := inertia.Props{
			"type":       typ,
			"items":      map[string]any{"data": cards},
			"pagination": pageLinks(params.Page, page.HasMore),
			"filters": map[string]any{
				"genre":    params.Genre,
				"platform": params.Platform,
				"watch":    params.Watch,
				"year":     params.Year,
				"minScore": params.MinScore,
				"sort":     params.Sort,
				"per":      params.Limit,
				"mode":     params.Entities["game_mode"],
			},
			"facets": facets,
			// The entity filters other than genre, platform and game mode come from links (the old entity pages):
			// they show as removable chips with the entity's name.
			"active": active,
		}
		if page.Total != nil {
			props["total"] = *page.Total
			// A page past the end (an old link, a filter that shrank) shows a link back instead of an empty grid.
			props["lastPage"] = max(1, (*page.Total+params.Limit-1)/params.Limit)
		}

		return inertia.Render(c, "Catalog/Index", props)
	}
}

// entityFilters reads the entity filters of the listings beyond genre and platform (catalog.ListingFilters).
func entityFilters(get func(string) string) (map[string]string, error) {
	out := map[string]string{}
	for kind, param := range catalog.ListingFilters {
		if kind == "genre" || kind == "platform" {
			continue
		}
		v := strings.TrimSpace(get(param))
		if v == "" {
			continue
		}
		// PostgreSQL rejects NUL and invalid UTF-8 (a 500); a slug or code is short.
		if !utf8.ValidString(v) || strings.ContainsRune(v, 0) || len(v) > 200 {
			return nil, catalog.ErrInvalidQuery
		}
		out[kind] = v
	}

	return out, nil
}

// badListing maps bad filters and forged cursors to 400.
func badListing(err error) error {
	if errors.Is(err, catalog.ErrInvalidQuery) || errors.Is(err, catalog.ErrInvalidCursor) {
		return errs.New(http.StatusBadRequest, err.Error())
	}

	return err
}
