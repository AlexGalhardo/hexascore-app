// Package profile serves the /profile sections (SDD 03 section 6.1).
package profile

import (
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
)

// Handlers serve the profile page and its section forms.
type Handlers struct{ a *accounts.Accounts }

// New builds the handlers.
func New(a *accounts.Accounts) *Handlers { return &Handlers{a: a} }

func (h *Handlers) user(c *web.Ctx) (models.User, error) { return auth.User[models.User](c, "web") }

// Show renders Profile/Index with one prop per section.
func (h *Handlers) Show(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return err
	}

	p, err := h.a.Profile(c.Context(), u.ID)
	if err != nil {
		return err
	}

	tf, err := h.a.TwoFactor(c.Context(), u.ID)
	if err != nil {
		return err
	}

	sessions, err := h.a.Sessions(c.Context(), u.ID, c.Session().ID())
	if err != nil {
		return err
	}

	passkeys, err := h.a.PasskeyList(c.Context(), u.ID)
	if err != nil {
		return err
	}

	rows := make([]map[string]any, 0, len(sessions))
	for _, s := range sessions {
		rows = append(rows, map[string]any{
			"handle": s.Handle, "current": s.Current, "createdAt": s.CreatedAt, "lastSeenAt": s.LastSeenAt,
			"userAgent": s.UserAgent,
		})
	}

	return inertia.Render(c, "Profile/Index", inertia.Props{
		"details":     inertia.Props{"name": u.Name, "email": u.Email, "locale": p.Locale},
		"passwordSet": p.PasswordSet,
		"twoFactor":   inertia.Props{"enabled": tf.Enabled()},
		"passkeys":    passkeys,
		"sessions":    rows,
		"privacy": inertia.Props{
			"profilePrivate": p.ProfilePrivate, "personalizationOptOut": p.PersonalizationOptOut,
		},
		"deletionScheduledAt": u.DeletionScheduledAt,
	})
}

// UpdateDetails saves name and locale.
func (h *Handlers) UpdateDetails(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.ProfileDetailsSchema)
	if err != nil {
		return err
	}

	if err := h.a.UpdateDetails(c.Context(), u.ID, in.Name, in.Locale); err != nil {
		return err
	}

	inertia.Flash(c, "status", h.a.T(c.Context(), "auth.profile.details_saved"))

	return c.Redirect("/profile")
}

// UpdatePassword changes the password and keeps only the current session.
func (h *Handlers) UpdatePassword(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.ProfilePasswordSchema)
	if err != nil {
		return err
	}

	keep := c.Session().ID()
	if err := h.a.ChangePassword(c.Context(), u, in.CurrentPassword, in.NewPassword, keep); err != nil {
		return err
	}

	// The version bump ended the current session too: keep it.
	if err := h.a.Web.RefreshSession(c); err != nil {
		return err
	}

	inertia.Flash(c, "status", h.a.T(c.Context(), "auth.profile.password_saved"))

	return c.Redirect("/profile")
}

// UpdatePrivacy saves the privacy switches.
func (h *Handlers) UpdatePrivacy(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.ProfilePrivacySchema)
	if err != nil {
		return err
	}

	if err := h.a.UpdatePrivacy(c.Context(), u.ID, in.ProfilePrivate, in.PersonalizationOptOut); err != nil {
		return err
	}

	inertia.Flash(c, "status", h.a.T(c.Context(), "auth.profile.privacy_saved"))

	return c.Redirect("/profile")
}

// RevokeSession ends one other session.
func (h *Handlers) RevokeSession(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return err
	}

	if err := h.a.RevokeSession(c.Context(), u.ID, c.Param("handle")); err != nil {
		return err
	}

	return c.Redirect("/profile")
}

// RevokeAllSessions ends every session but this one.
func (h *Handlers) RevokeAllSessions(c *web.Ctx) error {
	u, err := h.user(c)
	if err != nil {
		return err
	}

	if err := h.a.RevokeOtherSessions(c.Context(), u.ID, c.Session().ID()); err != nil {
		return err
	}

	if err := h.a.Web.RefreshSession(c); err != nil {
		return err
	}

	return c.Redirect("/profile")
}
