package controllers

import (
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// Status serves the item status and time-to-beat write routes for one guard, like Ratings.
type Status struct {
	w     *scoring.Writer
	guard string
	api   bool
}

// NewStatus builds the handlers; api selects JSON answers.
func NewStatus(w *scoring.Writer, guard string, api bool) *Status {
	return &Status{w: w, guard: guard, api: api}
}

// PutStatus sets the signed-in user's status of the item in {id}.
func (h *Status) PutStatus(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.StatusSchema)
	if err != nil {
		return err
	}

	sum, err := h.w.SetStatus(c.Context(), u.ID, c.Param("id"), in.Status)
	if err != nil {
		return err
	}

	if !h.api {
		return c.Back("/")
	}

	return c.JSON(200, map[string]any{"status": in.Status, "item": sum})
}

// DeleteStatus clears the signed-in user's status (and time-to-beat report) of the item in {id}.
func (h *Status) DeleteStatus(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	if _, err := h.w.SetStatus(c.Context(), u.ID, c.Param("id"), ""); err != nil {
		return err
	}

	return h.done(c)
}

// PutTimeToBeat reports the signed-in user's time to beat for the game in {id}.
func (h *Status) PutTimeToBeat(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.TimeToBeatSchema)
	if err != nil {
		return err
	}
	if in.Main == nil && in.Extras == nil && in.Completionist == nil {
		v := errs.NewValidation()
		v.Append("mainHours", "required")

		return v
	}

	hours := scoring.TimeToBeatHours{Main: in.Main, Extras: in.Extras, Completionist: in.Completionist}
	sum, err := h.w.SaveTimeToBeat(c.Context(), u.ID, c.Param("id"), hours)
	if err != nil {
		return err
	}

	if !h.api {
		return c.Back("/")
	}

	return c.JSON(200, map[string]any{"timeToBeat": hours, "item": sum})
}

// DeleteTimeToBeat removes the signed-in user's report.
func (h *Status) DeleteTimeToBeat(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	if _, err := h.w.DeleteTimeToBeat(c.Context(), u.ID, c.Param("id")); err != nil {
		return err
	}

	return h.done(c)
}

// Me answers the user's own rating, status and time to beat of the item in {id}.
func (h *Status) Me(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	me, err := h.w.Me(c.Context(), u.ID, c.Param("id"))
	if err != nil {
		return err
	}

	return c.JSON(200, me)
}

func (h *Status) done(c *web.Ctx) error {
	if !h.api {
		return c.Back("/")
	}

	return c.NoContent()
}
