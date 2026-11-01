package appapi

import (
	"net/http"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/authz"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Lists serves GET /app/v1/me/lists and GET /app/v1/lists/{id}. Writes stay on
// routes.Lists behind lists:write. Edition choice stays on status:write
// (SDD 05 §6.4 and SDD 14 §6.3). The caller puts catalog:read on these reads.
type Lists struct {
	svc   *library.Lists
	gate  *authz.Gate[*models.User]
	guard string
}

// NewLists builds the read handlers. gate must already register ListPolicy.
func NewLists(svc *library.Lists, gate *authz.Gate[*models.User], guard string) *Lists {
	return &Lists{svc: svc, gate: gate, guard: guard}
}

// Index is GET /app/v1/me/lists. Summaries already returns every list the owner
// has, and the per-type cap is small, so this page is never partial.
func (h *Lists) Index(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	lists, err := h.svc.Summaries(c.Context(), u.ID)
	if err != nil {
		return err
	}

	if typ := c.Query("type"); typ != "" {
		filtered := make([]library.ListSummary, 0, len(lists))
		for _, row := range lists {
			if row.Type == typ {
				filtered = append(filtered, row)
			}
		}
		lists = filtered
	}

	return c.JSON(http.StatusOK, map[string]any{
		"data": lists,
		"meta": map[string]any{"next_cursor": nil},
	})
}

// Show is GET /app/v1/lists/{id}. ListPolicy.View answers 404 for another
// user's private list (DenyAsNotFound).
func (h *Lists) Show(c *web.Ctx) error {
	ctx := c.Context()
	if _, err := auth.User[models.User](c, h.guard); err != nil {
		return err
	}

	row, err := h.svc.Get(ctx, c.Param("id"))
	if err != nil {
		return err
	}
	if err := h.gate.Authorize(ctx, "view", &row); err != nil {
		return err
	}

	items, err := h.svc.Entries(ctx, row.ID, i18n.Locale(ctx))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]any{
		"id":          row.ID,
		"type":        row.Type,
		"title":       row.Title,
		"description": row.Description,
		"isPublic":    row.IsPublic,
		"hidden":      row.Hidden,
		"itemCount":   row.ItemCount,
		"version":     row.Version,
		"createdAt":   row.CreatedAt,
		"updatedAt":   row.UpdatedAt,
		"items":       items,
	})
}
