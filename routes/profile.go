package routes

import (
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/controllers/profile"
)

// Profile registers the /profile page and its section forms (SDD 03 section 6.1).
func Profile(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	h := profile.New(acc)
	required, verified := auth.Required(acc.Web), auth.Verified(acc.Web)

	r.Get("/profile", h.Show).Name("profile").Use(required)
	r.Patch("/profile/details", h.UpdateDetails).Name("profile.details").Use(required, verified)
	r.Put("/profile/password", h.UpdatePassword).Name("profile.password").Use(required, verified)
	r.Patch("/profile/privacy", h.UpdatePrivacy).Name("profile.privacy").Use(required, verified)
	r.Patch("/profile/notifications", h.UpdateNotifications).Name("profile.notifications").Use(required, verified)
	r.Delete("/profile/sessions/{handle}", h.RevokeSession).Name("profile.sessions.revoke").Use(required)
	r.Post("/profile/sessions/revoke-all", h.RevokeAllSessions).Name("profile.sessions.revoke-all").Use(required)

	return nil
}
