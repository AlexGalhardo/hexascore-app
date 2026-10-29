package controllers

import (
	"errors"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// Item serves the item pages Catalog/Show and Catalog/Edition (SDD 09 §6.2).
type Item struct {
	app    *vukem.App
	writer *scoring.Writer
	user   func(c *web.Ctx) (models.User, error)
}

// NewItem builds the handlers; writer reads the viewer's own rating, status and time to beat.
func NewItem(app *vukem.App, writer *scoring.Writer, user func(c *web.Ctx) (models.User, error)) *Item {
	return &Item{app: app, writer: writer, user: user}
}

// Show renders Catalog/Show for the item type typ; `me` and `myLists` exist only when logged in.
func (h *Item) Show(typ string) web.HandlerFunc {
	return func(c *web.Ctx) error {
		db, err := h.app.DB()
		if err != nil {
			return err
		}

		page, err := catalog.LoadItemPage(c.Context(), db, typ, c.Param("slug"), i18n.Locale(c.Context()))
		if errors.Is(err, catalog.ErrItemNotFound) {
			return NotFound(c)
		}
		if err != nil {
			return err
		}

		props := inertia.Props{
			"item":          page.Item,
			"ratingSummary": page.RatingSummary,
			"entities":      page.Entities,
			"editions":      page.Editions,
			"editionCount":  page.EditionCount,
			"timeToBeat":    page.TimeToBeat,
		}

		if u, err := h.user(c); err == nil {
			me, err := h.writer.Me(c.Context(), u.ID, page.Item.ID)
			if err != nil {
				return err
			}
			lists, err := catalog.ListsFor(c.Context(), db, u.ID, typ, page.Item.ID)
			if err != nil {
				return err
			}
			props["me"], props["myLists"] = me, lists
		}

		return inertia.Render(c, "Catalog/Show", props)
	}
}

// Edition renders Catalog/Edition for one edition of the book in {slug}.
func (h *Item) Edition(c *web.Ctx) error {
	db, err := h.app.DB()
	if err != nil {
		return err
	}

	page, err := catalog.LoadItemPage(c.Context(), db, "book", c.Param("slug"), i18n.Locale(c.Context()))
	if errors.Is(err, catalog.ErrItemNotFound) {
		return NotFound(c)
	}
	if err != nil {
		return err
	}

	ed, err := catalog.LoadEdition(c.Context(), db, page.Item.ID, c.Param("id"))
	if errors.Is(err, catalog.ErrItemNotFound) {
		return NotFound(c)
	}
	if err != nil {
		return err
	}

	return inertia.Render(c, "Catalog/Edition", inertia.Props{
		"work":    map[string]string{"id": page.Item.ID, "slug": page.Item.Slug, "title": page.Item.Title},
		"edition": ed,
	})
}
