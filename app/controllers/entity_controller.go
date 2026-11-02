package controllers

import (
	"errors"
	"net/http"
	"net/url"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// Entity answers the old entity pages (people, companies, franchises, platforms, genres, game modes, countries and
// languages): no entity has a page of its own any more, each one is a filter of the section listings (owner
// 2026-11-02).
type Entity struct {
	app *vukem.App
}

// NewEntity builds the handlers.
func NewEntity(app *vukem.App) *Entity {
	return &Entity{app: app}
}

// listingURL is the section listing of item type typ filtered by the entity key, with the locale prefix.
func listingURL(prefix, typ, kind, key string) string {
	q := url.Values{"sort": {"-popularity"}, "per": {"25"}, catalog.ListingFilters[kind]: {key}}

	return prefix + "/" + catalog.Sections[typ] + "?" + q.Encode()
}

// Show answers /{section}/{entity path}/{key} with one 301 to that section's listing filtered by the entity.
func (h *Entity) Show(kind, param, typ string) web.HandlerFunc {
	return func(c *web.Ctx) error {
		return c.RedirectStatus(http.StatusMovedPermanently,
			listingURL(i18n.PathPrefix(c.Context()), typ, kind, c.Param(param)))
	}
}

// Home answers the old top-level entity path (/{entity path}/{key} plus suffix, "" or "/changes") with one 301:
// the bare path to the filtered listing of the entity's category, /changes to the changes page under it; `?type=` when it names an item type, else the type with the
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

		prefix := i18n.PathPrefix(c.Context())
		if suffix == "" {
			// The entity has no page: one 301 straight to the section listing filtered by it (owner 2026-11-02).
			return c.RedirectStatus(http.StatusMovedPermanently, listingURL(prefix, typ, kind, key))
		}

		u := url.URL{
			Path:     prefix + "/" + catalog.Sections[typ] + "/" + catalog.EntityPaths[kind] + "/" + key + suffix,
			RawQuery: q.Encode(),
		}

		return c.RedirectStatus(http.StatusMovedPermanently, u.String())
	}
}
