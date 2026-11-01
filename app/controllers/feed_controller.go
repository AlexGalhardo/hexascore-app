package controllers

import (
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/social"
)

// Feed serves /feed: activity of the users the viewer follows (SDD 14 FR-14-16). 404 while the flag is off.
type Feed struct {
	svc   *social.Follows
	guard string
}

// NewFeed builds the handler for the session guard.
func NewFeed(svc *social.Follows, guard string) *Feed {
	return &Feed{svc: svc, guard: guard}
}

// Index renders one page of the feed; ?cursor= is the last event id of the previous page.
func (h *Feed) Index(c *web.Ctx) error {
	if !h.svc.Enabled(c.Context()) {
		return NotFound(c)
	}

	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	entries, next, err := h.svc.Feed(c.Context(), u.ID, i18n.Locale(c.Context()), c.Query("cursor"))
	if err != nil {
		return err
	}

	return inertia.Render(c, "Feed/Index", inertia.Props{"events": entries, "nextCursor": next})
}
