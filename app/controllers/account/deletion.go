package account

import (
	"errors"
	"net/http"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"
	"github.com/AlexGalhardo/vukem/z"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Deletion is the schedule form; the client shows the 10 s countdown (SDD 03 FR-03-15).
type Deletion struct {
	Confirm bool `json:"confirm" form:"confirm"`
}

// DeletionSchema requires an explicit confirmation.
var DeletionSchema = z.Struct[Deletion](z.Shape{"confirm": z.Bool()}).Refine(func(d Deletion) error {
	if !d.Confirm {
		return z.FieldError("confirm", "Confirm the deletion.")
	}

	return nil
})

// ScheduleDeletion starts the grace period and signs the user out everywhere.
func (h *Handlers) ScheduleDeletion(c *web.Ctx) error {
	u, err := auth.User[models.User](c, "web")
	if err != nil {
		return err
	}

	if _, err := web.Parse(c, DeletionSchema); err != nil {
		return err
	}

	if _, err := h.a.ScheduleDeletion(c.Context(), u.ID); err != nil {
		return err
	}

	if err := h.a.Web.Logout(c); err != nil {
		return err
	}

	inertia.ClearHistory(c)

	return c.Redirect("/login")
}

// CancelDeletion ends the grace period.
func (h *Handlers) CancelDeletion(c *web.Ctx) error {
	u, err := auth.User[models.User](c, "web")
	if err != nil {
		return err
	}

	if err := h.a.CancelDeletion(c.Context(), u.ID); err != nil {
		if errors.Is(err, accounts.ErrNotScheduled) {
			return errs.New(http.StatusConflict, h.a.T(c.Context(), "auth.deletion.not_scheduled"))
		}

		return err
	}

	return c.Redirect("/dashboard")
}

// HelpDeleteAccount is the public page Google Play requires.
func (*Handlers) HelpDeleteAccount(c *web.Ctx) error {
	return inertia.Render(c, "Help/DeleteAccount", nil)
}
