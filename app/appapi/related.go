package appapi

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/related"
)

// Related serves GET /app/v1/items/{id}/related (SDD 08 section 6). Similar items stay off /api/v1 (SDD 08 section 2).
type Related struct{ db *database.DB }

// NewRelated builds the handler.
func NewRelated(db *database.DB) *Related { return &Related{db: db} }

// Show is GET /app/v1/items/{id}/related: {similar, crossMedia}, 404 for an unknown or non-public item.
func (h *Related) Show(c *web.Ctx) error {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return errs.New(http.StatusNotFound, "Unknown item.")
	}

	var typ string

	err := h.db.QueryRowContext(c.Context(), `SELECT type FROM items WHERE id = $1 AND visibility = 'public'`, id).
		Scan(&typ)
	if errors.Is(err, sql.ErrNoRows) {
		return errs.New(http.StatusNotFound, "Unknown item.")
	}
	if err != nil {
		return err
	}

	out, err := related.Load(c.Context(), h.db, id, typ, i18n.Locale(c.Context()))
	if err != nil {
		return err
	}

	c.SetHeader("Cache-Control", "private, max-age=300")

	return c.JSON(http.StatusOK, out)
}
