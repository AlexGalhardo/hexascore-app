package controllers

import (
	"strings"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/social"
)

// Follows serves follow, unfollow and the follow-request answers on the web (SDD 14 §12.6). Every handler answers
// 404 while the feature flag `social` is off.
type Follows struct {
	svc   *social.Follows
	guard string
}

// NewFollows builds the handlers for the session guard.
func NewFollows(svc *social.Follows, guard string) *Follows {
	return &Follows{svc: svc, guard: guard}
}

func (h *Follows) user(c *web.Ctx) (models.User, error) {
	if !h.svc.Enabled(c.Context()) {
		return models.User{}, social.ErrNotFound()
	}

	return auth.User[models.User](c, h.guard)
}

func (h *Follows) back(c *web.Ctx) error {
	return c.Back("/u/" + strings.ToLower(c.Param("nickname")))
}

// Follow follows {nickname}, or sends a request when the profile is private.
func (h *Follows) Follow(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return pageError(c, err)
	}

	if _, err := h.svc.Follow(c.Context(), u.ID, c.Param("nickname")); err != nil {
		return pageError(c, err)
	}

	return h.back(c)
}

// Unfollow removes the follow or cancels the pending request.
func (h *Follows) Unfollow(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return pageError(c, err)
	}

	if err := h.svc.Unfollow(c.Context(), u.ID, c.Param("nickname")); err != nil {
		return pageError(c, err)
	}

	return h.back(c)
}

// Accept accepts the request {nickname} sent to the signed-in user.
func (h *Follows) Accept(c *web.Ctx) error { return h.answer(c, true) }

// Decline deletes the request {nickname} sent to the signed-in user.
func (h *Follows) Decline(c *web.Ctx) error { return h.answer(c, false) }

func (h *Follows) answer(c *web.Ctx, accept bool) error {
	u, err := h.user(c)
	if err != nil {
		return pageError(c, err)
	}

	if err := h.svc.Answer(c.Context(), u.ID, c.Param("nickname"), accept); err != nil {
		return pageError(c, err)
	}

	return c.Back("/notifications")
}
