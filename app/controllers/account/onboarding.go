package account

import (
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
)

// ShowOnboarding asks a signed-in, verified user without a nickname for one;
// users who have it go to the dashboard.
func (h *Handlers) ShowOnboarding(c *web.Ctx) error {
	u, err := auth.User[models.User](c, "web")
	if err != nil {
		return err
	}

	if has, err := h.a.HasNickname(c.Context(), u.ID); err != nil || has {
		if err != nil {
			return err
		}

		return c.Redirect("/dashboard")
	}

	return inertia.Render(c, "Auth/Onboarding", nil)
}

// Onboarding stores the nickname once (it never changes afterwards, SDD 03).
// A nickname someone else holds answers 422 `nickname-taken`, also when the
// unique index decides a race.
func (h *Handlers) Onboarding(c *web.Ctx) error {
	u, err := auth.User[models.User](c, "web")
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.OnboardingSchema)
	if err != nil {
		return err
	}

	if err := h.a.SetNickname(c.Context(), u.ID, in.Nickname, "nickname"); err != nil {
		return err
	}

	return c.Redirect("/dashboard")
}
