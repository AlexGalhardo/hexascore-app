package controllers

import (
	"strings"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Index renders Lists/Index with the signed-in user's lists (SDD 14 §6.2).
func (h *Lists) Index(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	lists, err := h.svc.Summaries(c.Context(), u.ID)
	if err != nil {
		return err
	}

	return inertia.Render(c, "Lists/Index", inertia.Props{
		"lists":  lists,
		"limits": map[string]int{"perType": library.MaxListsPerType, "items": library.MaxItemsPerList},
	})
}

// Edit renders Lists/Edit for the owner; anyone else gets a 404.
func (h *Lists) Edit(c *web.Ctx) error {
	l, err := h.owned(c, "update")
	if err != nil {
		return err
	}

	items, err := h.svc.Entries(c.Context(), l.ID, i18n.Locale(c.Context()))
	if err != nil {
		return err
	}

	return inertia.Render(c, "Lists/Edit", inertia.Props{
		"list": listJSON(l), "items": items, "version": l.Version,
		"limits": map[string]int{"items": library.MaxItemsPerList},
	})
}

// Show renders User/List; the list must belong to {nickname} and pass ListPolicy.View.
func (h *Lists) Show(c *web.Ctx) error {
	ctx := c.Context()

	l, err := h.svc.Get(ctx, c.Param("id"))
	if err != nil {
		return err
	}

	if err := h.gate.Authorize(ctx, "view", &l); err != nil {
		return err
	}

	nick, banned, err := h.svc.OwnerNickname(ctx, l.UserID)
	if err != nil {
		return err
	}

	if nick == "" || !strings.EqualFold(nick, c.Param("nickname")) {
		return library.ErrNotFound()
	}

	items, err := h.svc.Entries(ctx, l.ID, i18n.Locale(ctx))
	if err != nil {
		return err
	}

	viewer, err := auth.UserFrom[models.User](ctx, h.guard)
	canEdit := err == nil && viewer.ID == l.UserID

	indexable := l.IsPublic && !l.Hidden && !l.OwnerPrivate && !banned
	if !indexable {
		c.Writer().Header().Set("X-Robots-Tag", "noindex")
	}

	return inertia.Render(c, "User/List", inertia.Props{
		"list": listJSON(l), "owner": map[string]string{"nickname": nick}, "items": items, "canEdit": canEdit,
		"indexable": indexable,
	})
}
