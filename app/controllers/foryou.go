package controllers

import (
	"context"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/foryou"
)

// ForYou serves the ForYou/Index page: one deferred section per type (SDD 08 section 6).
type ForYou struct {
	svc   *foryou.Service
	guard string
}

// NewForYou builds the page handler for one guard.
func NewForYou(svc *foryou.Service, guard string) *ForYou { return &ForYou{svc: svc, guard: guard} }

// Index is GET /for-you.
func (h *ForYou) Index(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	out, err := h.svc.OptedOut(c.Context(), u.ID)
	if err != nil {
		return err
	}

	props := inertia.Props{"optedOut": out}

	if !out {
		locale := i18n.Locale(c.Context())

		for _, typ := range homeTypes {
			props[typ] = inertia.Defer(func(ctx context.Context) (any, error) {
				return h.svc.Get(ctx, u.ID, typ, locale)
			})
		}
	}

	return inertia.Render(c, "ForYou/Index", props)
}
