package controllers

import (
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
	"github.com/AlexGalhardo/hexascore-app/app/social"
)

// Reviews serves writing and deleting a review on the web (SDD 14 §12.10); 404 while the flag `social` is off.
type Reviews struct {
	svc   *social.Follows
	guard string
}

// NewReviews builds the handlers for the session guard.
func NewReviews(svc *social.Follows, guard string) *Reviews {
	return &Reviews{svc: svc, guard: guard}
}

func (h *Reviews) user(c *web.Ctx) (models.User, error) {
	if !h.svc.Enabled(c.Context()) {
		return models.User{}, social.ErrNotFound()
	}

	return auth.User[models.User](c, h.guard)
}

// Put creates or edits the signed-in user's review of item {id}.
func (h *Reviews) Put(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return pageError(c, err)
	}

	in, err := web.Parse(c, requests.ReviewSchema)
	if err != nil {
		return err
	}

	if err := h.svc.SaveReview(c.Context(), u.ID, c.Param("id"), in.Body, in.Spoiler); err != nil {
		return pageError(c, err)
	}

	return c.Back("/")
}

// Delete removes the signed-in user's review of item {id}.
func (h *Reviews) Delete(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return pageError(c, err)
	}

	if err := h.svc.DeleteReview(c.Context(), u.ID, c.Param("id")); err != nil {
		return pageError(c, err)
	}

	return c.Back("/")
}
