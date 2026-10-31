package community

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

const mineLimit = 100

type actionForm struct {
	Version int `json:"version"`
}

// load reads the suggestion in {id} and runs the policy action on it; a policy denial is the gate's 404.
func (h *Suggest) load(c *web.Ctx, action string) (community.StatusView, error) {
	v, err := community.LoadStatus(c.Context(), h.db, c.Param("id"))
	if errors.Is(err, community.ErrDraftNotFound) {
		return v, errs.ErrNotFound
	}
	if err != nil {
		return v, err
	}

	owned := models.Suggestion{ID: v.ID, AuthorID: v.AuthorID, Status: v.Status}
	if err := h.gate.Authorize(c.Context(), action, &owned); err != nil {
		return v, err
	}

	return v, nil
}

// Show renders Community/SuggestionStatus for the author or an admin (GET /suggestions/{id}).
func (h *Suggest) Show(c *web.Ctx) error {
	v, err := h.load(c, "view")
	if err != nil {
		return err
	}

	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	quota, err := community.QuotaFor(c.Context(), h.db, u.ID, h.now())
	if err != nil {
		return err
	}

	return inertia.Render(c, "Community/SuggestionStatus", inertia.Props{
		"suggestion": v,
		"canAct":     v.AuthorID.Valid && v.AuthorID.Int64 == u.ID,
		"quota":      quota,
		"open":       h.open(c),
	})
}

// Mine renders Community/MySuggestions (GET /me/suggestions).
func (h *Suggest) Mine(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	items, err := community.ListMine(c.Context(), h.db, u.ID, mineLimit)
	if err != nil {
		return err
	}

	return inertia.Render(c, "Community/MySuggestions", inertia.Props{"suggestions": items, "open": h.open(c)})
}

// SubmitDraft moves the author's draft to submitted (POST /suggestions/{id}/submit): 303 to the status page, or the
// SDD 04 §6 problem (403/409/422/423/429).
func (h *Suggest) SubmitDraft(c *web.Ctx) error {
	if !h.open(c) {
		return errs.New(http.StatusForbidden, "Suggestions are closed").
			WithExtension("code", "community-submissions-closed")
	}

	v, err := h.load(c, "submit")
	if err != nil {
		return err
	}

	f, err := decodeAction(c)
	if err != nil {
		return err
	}

	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	if _, err := community.Submit(c.Context(), h.db, v.ID, u.ID, f.Version, h.now()); err != nil {
		return problemError(err)
	}

	return c.RedirectStatus(http.StatusSeeOther, "/suggestions/"+v.ID)
}

// Withdraw pulls the author's suggestion back (POST /suggestions/{id}/withdraw). It stays available while
// submissions are closed: leaving the queue is always allowed.
func (h *Suggest) Withdraw(c *web.Ctx) error {
	v, err := h.load(c, "withdraw")
	if err != nil {
		return err
	}

	f, err := decodeAction(c)
	if err != nil {
		return err
	}

	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	if err := community.Withdraw(c.Context(), h.db, v.ID, u.ID, f.Version); err != nil {
		return problemError(err)
	}

	return c.RedirectStatus(http.StatusSeeOther, "/suggestions/"+v.ID)
}

func decodeAction(c *web.Ctx) (actionForm, error) {
	var f actionForm
	if err := json.NewDecoder(http.MaxBytesReader(c.Writer(), c.Request().Body, 1<<10)).Decode(&f); err != nil {
		return f, errs.NewValidation().Add("body", "invalid")
	}

	return f, nil
}
