package controllers

import (
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/flash"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// Ratings serves the rating write routes for one guard: the session guard
// answers 303 back (SDD 09 §6.2), the token guard answers JSON (SDD 02 §6).
type Ratings struct {
	w     *scoring.Writer
	guard string
	api   bool
}

// NewRatings builds the handlers; api selects JSON answers.
func NewRatings(w *scoring.Writer, guard string, api bool) *Ratings {
	return &Ratings{w: w, guard: guard, api: api}
}

// Put saves the signed-in user's rating of the item in {id}.
func (h *Ratings) Put(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.RatingSchema)
	if err != nil {
		return err
	}

	sum, err := h.w.Save(c.Context(), u.ID, c.Param("id"), in.Scores)
	if err != nil {
		return err
	}

	if !h.api {
		flash.Status(c, "flash.ratingSaved")

		return c.Back("/")
	}

	return c.JSON(200, map[string]any{"rating": in.Scores, "item": sum})
}

// Delete removes the signed-in user's rating of the item in {id}.
func (h *Ratings) Delete(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	if _, err := h.w.Delete(c.Context(), u.ID, c.Param("id")); err != nil {
		return err
	}

	if !h.api {
		flash.Status(c, "flash.ratingRemoved")

		return c.Back("/")
	}

	return c.NoContent()
}
