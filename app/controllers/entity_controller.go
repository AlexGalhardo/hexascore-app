package controllers

import (
	"errors"
	"slices"
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

// Show renders Entity/Show for the entity kind; param names the path parameter holding the slug (or the
// code for countries and languages). `?type=` pages one item type from `?cursor=`.
func (h *Entity) Show(kind, param string) web.HandlerFunc {
	return func(c *web.Ctx) error {
		db, err := h.app.DB()
		if err != nil {
			return err
		}

		typ := c.Query("type")
		if typ != "" && !slices.Contains([]string{"game", "movie", "series", "book"}, typ) {
			typ = ""
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

		if !page.Indexable {
			c.Writer().Header().Set("X-Robots-Tag", "noindex, follow")
		}

		meta := inertia.ScrollMeta{PageName: "cursor"}
		if typ != "" && len(page.Groups) == 1 && page.Groups[0].Next >= 0 {
			meta.NextPage = strconv.Itoa(page.Groups[0].Next)
		}

		return inertia.Render(c, "Entity/Show", inertia.Props{
			"entity":    page.Entity,
			"roles":     page.Roles,
			"items":     inertia.Scroll(map[string]any{"data": page.Groups}, meta),
			"counts":    page.Counts,
			"avgScore":  page.AvgScore,
			"filters":   map[string]string{"type": typ},
			"indexable": page.Indexable,
		})
	}
}
