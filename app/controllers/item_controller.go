package controllers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
	"github.com/AlexGalhardo/hexascore-app/app/social"
)

// Item serves the item pages Catalog/Show and Catalog/Edition (SDD 09 §6.2).
type Item struct {
	app    *vukem.App
	writer *scoring.Writer
	user   func(c *web.Ctx) (models.User, error)
	// Social adds the deferred `reviews` prop while the feature flag `social` is on (SDD 14 §12.10); nil on SQLite.
	Social *social.Follows
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
			"trailers":      page.Trailers,
			"cover":         page.Cover,
		}

		var viewerID int64
		if u, err := h.user(c); err == nil {
			viewerID = u.ID
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

		if h.Social.Enabled(c.Context()) {
			props["ogImage"] = ogImage("items", page.Item.ID)
			locale := i18n.Locale(c.Context())
			props["reviews"] = inertia.Defer(func(ctx context.Context) (any, error) {
				data, mine, err := h.Social.ItemReviews(ctx, viewerID, page.Item.ID, locale)

				return map[string]any{"data": data, "mine": mine}, err
			}).Group("reviews")
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

// Ratings serves the public users' scores page used by the item dialog.
func (h *Item) Ratings(c *web.Ctx) error {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return errs.New(http.StatusNotFound, "Unknown item.")
	}
	page, err := publicRatingsPage(c)
	if err != nil {
		return err
	}
	db, err := h.app.DB()
	if err != nil {
		return err
	}
	out, err := catalog.PublicItemRatings(c.Context(), db, id, page)
	if errors.Is(err, catalog.ErrItemNotFound) {
		return errs.New(http.StatusNotFound, "Unknown item.")
	}
	if err != nil {
		return err
	}
	c.SetHeader("Cache-Control", "public, max-age=60")
	return c.JSON(http.StatusOK, out)
}

func publicRatingsPage(c *web.Ctx) (int, error) {
	raw := c.Request().URL.Query().Get("page")
	if raw == "" {
		return 1, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > catalog.RatingsMaxPages {
		return 0, errs.New(http.StatusBadRequest, "page must be an integer between 1 and 100.")
	}
	return n, nil
}
