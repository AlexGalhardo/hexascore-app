package controllers

import (
	"context"
	"errors"
	"strconv"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/community"
)

// RecentChanges serves the public Catalog/RecentChanges tab of items, editions and entities (SDD 04 FR-04-13).
type RecentChanges struct {
	app *vukem.App
}

// NewRecentChanges builds the handlers.
func NewRecentChanges(app *vukem.App) *RecentChanges { return &RecentChanges{app: app} }

// Item serves /{section}/{slug}/changes for item type typ.
func (h *RecentChanges) Item(typ string) web.HandlerFunc {
	return h.page(func(ctx context.Context, db *database.DB, c *web.Ctx) (community.RecentTarget, error) {
		return community.ResolveItem(ctx, db, typ, c.Param("slug"))
	})
}

// Edition serves /books/{slug}/editions/{id}/changes.
func (h *RecentChanges) Edition() web.HandlerFunc {
	return h.page(func(ctx context.Context, db *database.DB, c *web.Ctx) (community.RecentTarget, error) {
		return community.ResolveEdition(ctx, db, c.Param("slug"), c.Param("id"))
	})
}

// Entity serves /{entity path}/{param}/changes for kind; param "code" matches by code.
func (h *RecentChanges) Entity(kind, param string) web.HandlerFunc {
	return h.page(func(ctx context.Context, db *database.DB, c *web.Ctx) (community.RecentTarget, error) {
		return community.ResolveEntity(ctx, db, kind, c.Param(param), param == "code")
	})
}

type resolver func(ctx context.Context, db *database.DB, c *web.Ctx) (community.RecentTarget, error)

func (h *RecentChanges) page(find resolver) web.HandlerFunc {
	return func(c *web.Ctx) error {
		db, err := h.app.DB()
		if err != nil {
			return err
		}

		target, err := find(c.Context(), db, c)
		if errors.Is(err, community.ErrTargetMissing) {
			return NotFound(c)
		}

		if err != nil {
			return err
		}

		// shortcut: OFFSET paging capped at 500 pages (10,000 rows); switch to an id cursor if a record gets busier.
		page, _ := strconv.Atoi(c.Query("page"))
		page = min(max(page, 1), 500)

		entries, more, err := community.RecentChanges(c.Context(), db, target, page)
		if err != nil {
			return err
		}

		var next any
		if more {
			next = page + 1
		}

		return inertia.Render(c, "Catalog/RecentChanges", inertia.Props{
			"target": target, "entries": entries, "page": page, "nextPage": next,
		})
	}
}
