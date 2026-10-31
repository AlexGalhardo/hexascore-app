package appapi

import (
	"net/http"
	"slices"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/foryou"
)

// ForYou serves GET /app/v1/me/for-you?type= (SDD 08 section 6).
type ForYou struct {
	svc   *foryou.Service
	guard string
}

// NewForYou builds the handler for one guard.
func NewForYou(svc *foryou.Service, guard string) *ForYou { return &ForYou{svc: svc, guard: guard} }

// Show is GET /app/v1/me/for-you. type is required: game, movie, series or book.
func (h *ForYou) Show(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	typ := c.Query("type")
	if !slices.Contains([]string{"game", "movie", "series", "book"}, typ) {
		return errs.NewValidation().Add("type", "must be game, movie, series or book")
	}

	list, err := h.svc.Get(c.Context(), u.ID, typ, i18n.Locale(c.Context()))
	if err != nil {
		return err
	}

	c.SetHeader("Cache-Control", "private, no-store")

	return c.JSON(http.StatusOK, list)
}
