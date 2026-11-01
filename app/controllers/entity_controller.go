package controllers

import (
	"errors"
	"net/http"
	"net/url"
	"path"
	"strconv"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/library"
)

// Entity serves Entity/Show for people, companies, franchises, platforms, genres, game modes, countries and
// languages (SDD 09 §6.2).
type Entity struct {
	app *vukem.App
}

// NewEntity builds the handlers.
func NewEntity(app *vukem.App) *Entity {
	return &Entity{app: app}
}

// Show renders Entity/Show for the entity kind under the section of item type typ
// (/{section}/{entity path}/{key}, owner round 2026-11-01 item 19); param names the path parameter holding the
// slug (or the code for countries and languages). `?cursor=` pages the type's items.
func (h *Entity) Show(kind, param, typ string) web.HandlerFunc {
	return func(c *web.Ctx) error {
		db, err := h.app.DB()
		if err != nil {
			return err
		}

		page, err := catalog.LoadEntityPage(
			c.Context(),
			db,
			kind,
			c.Param(param),
			i18n.Locale(c.Context()),
			typ,
			library.ParseOffset(c.Query("cursor")),
		)
		if errors.Is(err, catalog.ErrItemNotFound) {
			return NotFound(c)
		}
		if err != nil {
			return err
		}

		if code := page.Entity.Code; code != "" && code != c.Param(param) {
			u := *c.Request().URL
			u.Path = path.Join(path.Dir(u.Path), code)
			http.Redirect(c.Writer(), c.Request(), u.String(), http.StatusMovedPermanently)

			return nil
		}

		if !page.Indexable {
			c.Writer().Header().Set("X-Robots-Tag", "noindex, follow")
		}

		meta := inertia.ScrollMeta{PageName: "cursor"}
		if len(page.Groups) == 1 && page.Groups[0].Next >= 0 {
			meta.NextPage = strconv.Itoa(page.Groups[0].Next)
		}

		// The page does not render the Wikidata id; it stays out of the Inertia JSON (ADR 0022).
		entity := page.Entity
		entity.WikidataQID = ""

		return inertia.Render(c, "Entity/Show", inertia.Props{
			"entity":    entity,
			"roles":     page.Roles,
			"items":     inertia.Scroll(map[string]any{"data": page.Groups}, meta),
			"counts":    page.Counts,
			"avgScore":  page.AvgScore,
			"filters":   map[string]string{"type": typ},
			"indexable": page.Indexable,
		})
	}
}

// Home answers the old top-level entity path (/{entity path}/{key} plus suffix, "" or "/changes") with one 301
// to the same page under the entity's category: `?type=` when it names an item type, else the type with the
// most public items. The canonical code and the other query parameters travel along.
func (h *Entity) Home(kind, param, suffix string) web.HandlerFunc {
	return func(c *web.Ctx) error {
		db, err := h.app.DB()
		if err != nil {
			return err
		}

		typ, key, err := catalog.EntityHome(c.Context(), db, kind, c.Param(param))
		if errors.Is(err, catalog.ErrItemNotFound) {
			return NotFound(c)
		}
		if err != nil {
			return err
		}

		q := c.Request().URL.Query()
		if _, ok := catalog.Sections[q.Get("type")]; ok {
			typ = q.Get("type")
		}
		q.Del("type")

		u := url.URL{
			Path: i18n.PathPrefix(
				c.Context(),
			) + "/" + catalog.Sections[typ] + "/" + catalog.EntityPaths[kind] + "/" + key + suffix,
			RawQuery: q.Encode(),
		}

		return c.RedirectStatus(http.StatusMovedPermanently, u.String())
	}
}
