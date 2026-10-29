package controllers

import (
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/authz"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
)

// Lists serves the list write routes for one guard like Ratings: the session guard answers 303 back, the token
// guard JSON. Every handler loads the row before authorizing it (SDD 14 §6.1).
type Lists struct {
	svc   *library.Lists
	gate  *authz.Gate[*models.User]
	guard string
	api   bool
}

// NewLists builds the handlers; api selects JSON answers.
func NewLists(svc *library.Lists, gate *authz.Gate[*models.User], guard string, api bool) *Lists {
	return &Lists{svc: svc, gate: gate, guard: guard, api: api}
}

// Create adds a list for the signed-in user.
func (h *Lists) Create(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}
	if err := h.gate.Authorize(c.Context(), "create", (*models.List)(nil)); err != nil {
		return err
	}

	in, err := web.Parse(c, requests.ListCreateSchema)
	if err != nil {
		return err
	}

	l, err := h.svc.Create(c.Context(), u.ID, library.CreateInput{
		Type: in.Type, Title: in.Title, Description: in.Description, IsPublic: in.IsPublic, FirstItemID: in.FirstItemID,
	})
	if err != nil {
		return err
	}

	if !h.api {
		return c.Back("/lists")
	}

	return c.JSON(201, listJSON(l))
}

// Update patches the list in {id} when the signed-in user owns it.
func (h *Lists) Update(c *web.Ctx) error {
	l, err := h.owned(c, "update")
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.ListUpdateSchema)
	if err != nil {
		return err
	}

	out, err := h.svc.Update(
		c.Context(),
		l,
		library.UpdateInput{Title: in.Title, Description: in.Description, IsPublic: in.IsPublic},
	)
	if err != nil {
		return err
	}

	if !h.api {
		return c.Back("/lists")
	}

	return c.JSON(200, listJSON(out))
}

// Delete removes the list in {id} when the signed-in user owns it.
func (h *Lists) Delete(c *web.Ctx) error {
	l, err := h.owned(c, "delete")
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Context(), l); err != nil {
		return err
	}

	if !h.api {
		return c.Back("/lists")
	}

	return c.NoContent()
}

func (h *Lists) owned(c *web.Ctx, ability string) (models.List, error) {
	if _, err := auth.User[models.User](c, h.guard); err != nil {
		return models.List{}, err
	}

	l, err := h.svc.Get(c.Context(), c.Param("id"))
	if err != nil {
		return l, err
	}

	return l, h.gate.Authorize(c.Context(), ability, &l)
}

func listJSON(l models.List) map[string]any {
	return map[string]any{
		"id":          l.ID,
		"type":        l.Type,
		"title":       l.Title,
		"description": l.Description,
		"isPublic":    l.IsPublic,
		"hidden":      l.Hidden,
		"itemCount":   l.ItemCount,
		"version":     l.Version,
		"createdAt":   l.CreatedAt,
		"updatedAt":   l.UpdatedAt,
	}
}

// AddItem puts the item in {itemId} at the end of the list in {id}.
func (h *Lists) AddItem(c *web.Ctx) error {
	l, err := h.owned(c, "addItem")
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.ListItemSchema)
	if err != nil {
		return err
	}

	out, err := h.svc.AddItem(c.Context(), l, c.Param("itemId"), in.Note)
	if err != nil {
		return err
	}

	if !h.api {
		return c.Back("/lists")
	}

	return c.JSON(200, listJSON(out))
}

// RemoveItem takes the item in {itemId} out of the list in {id}.
func (h *Lists) RemoveItem(c *web.Ctx) error {
	l, err := h.owned(c, "addItem")
	if err != nil {
		return err
	}
	if err := h.svc.RemoveItem(c.Context(), l, c.Param("itemId")); err != nil {
		return err
	}

	if !h.api {
		return c.Back("/lists")
	}

	return c.NoContent()
}

// Reorder replaces the order of the list in {id}.
func (h *Lists) Reorder(c *web.Ctx) error {
	l, err := h.owned(c, "reorder")
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.ListOrderSchema)
	if err != nil {
		return err
	}

	version, err := h.svc.Reorder(c.Context(), l, in.Version, in.ItemIDs)
	if err != nil {
		return err
	}

	if !h.api {
		return c.Back("/lists")
	}

	return c.JSON(200, map[string]any{"version": version})
}
