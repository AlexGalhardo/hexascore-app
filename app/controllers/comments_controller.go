package controllers

import (
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
	"github.com/AlexGalhardo/hexascore-app/app/social"
)

// Comments serves comments and reactions on reviews, comments and lists (SDD 14 §12.11); 404 while the flag
// `social` is off.
type Comments struct {
	svc   *social.Follows
	guard string
}

// NewComments builds the handlers for the session guard.
func NewComments(svc *social.Follows, guard string) *Comments {
	return &Comments{svc: svc, guard: guard}
}

func (h *Comments) user(c *web.Ctx) (models.User, error) {
	if !h.svc.Enabled(c.Context()) {
		return models.User{}, social.ErrNotFound()
	}

	return auth.User[models.User](c, h.guard)
}

// Create posts a comment on {type} (review or list) {id}.
func (h *Comments) Create(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return pageError(c, err)
	}

	in, err := web.Parse(c, requests.CommentSchema)
	if err != nil {
		return err
	}

	if err := h.svc.AddComment(c.Context(), u.ID, c.Param("type"), c.Param("id"), in.Body, in.ParentID); err != nil {
		return pageError(c, err)
	}

	return c.Back("/")
}

// Delete removes the signed-in user's comment {id}.
func (h *Comments) Delete(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return pageError(c, err)
	}

	if err := h.svc.DeleteComment(c.Context(), u.ID, c.Param("id")); err != nil {
		return pageError(c, err)
	}

	return c.Back("/")
}

// React sets the signed-in user's reaction on {type} {id}.
func (h *Comments) React(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return pageError(c, err)
	}

	in, err := web.Parse(c, requests.ReactionSchema)
	if err != nil {
		return err
	}

	if err := h.svc.React(c.Context(), u.ID, c.Param("type"), c.Param("id"), in.Kind); err != nil {
		return pageError(c, err)
	}

	return c.Back("/")
}

// Unreact removes the signed-in user's reaction on {type} {id}.
func (h *Comments) Unreact(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return pageError(c, err)
	}

	if err := h.svc.Unreact(c.Context(), u.ID, c.Param("type"), c.Param("id")); err != nil {
		return pageError(c, err)
	}

	return c.Back("/")
}
