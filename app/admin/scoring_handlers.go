package admin

import (
	"strconv"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// ScoringHandlers serve the moderation POST routes of SDD 02 §6 behind the
// panel gate and CSRF. They are custom routes because row actions on a
// resource are not in vukem (VF-008).
type ScoringHandlers struct {
	Accounts *accounts.Accounts
	Writer   *scoring.Writer
	// Authorize is the panel gate check (404 for non-admins).
	Authorize func(c *web.Ctx) error
}

func (h ScoringHandlers) actor(c *web.Ctx) (int64, error) {
	if err := h.Authorize(c); err != nil {
		return 0, err
	}

	u, err := auth.UserFrom[models.User](c.Context(), h.Accounts.Web.Name())
	if err != nil {
		return 0, err
	}

	return u.ID, nil
}

// ApproveBurst deletes the burst flags and lifts the limits.
func (h ScoringHandlers) ApproveBurst(c *web.Ctx) error { return h.resolve(c, true) }

// ExcludeBurst sets the burst flags to shadow 0.
func (h ScoringHandlers) ExcludeBurst(c *web.Ctx) error { return h.resolve(c, false) }

func (h ScoringHandlers) resolve(c *web.Ctx, approve bool) error {
	actor, err := h.actor(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return errs.ErrNotFound
	}

	if err := h.Writer.ResolveBurst(c.Context(), actor, id, approve); err != nil {
		return err
	}

	return c.Redirect("/admin")
}

// RestoreFlag removes one rating flag.
func (h ScoringHandlers) RestoreFlag(c *web.Ctx) error {
	actor, err := h.actor(c)
	if err != nil {
		return err
	}

	userID, err := strconv.ParseInt(c.Param("user"), 10, 64)
	if err != nil {
		return errs.ErrNotFound
	}

	itemID := c.Param("item")
	if _, err := uuid.Parse(itemID); err != nil {
		return errs.ErrNotFound
	}

	if err := h.Writer.RestoreFlag(c.Context(), actor, userID, itemID); err != nil {
		return err
	}

	return c.Redirect("/admin")
}
