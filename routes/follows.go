package routes

import (
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/controllers"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	"github.com/AlexGalhardo/hexascore-app/app/social"
)

// Follows registers follow, unfollow, the follow-request answers and /feed (SDD 14 §12.6). The routes always exist; the
// handlers answer 404 while the feature flag `social` is off. The table exists only on Postgres (VF-033).
func Follows(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	if acc.DB.Dialect().Name() != "postgres" {
		return nil
	}

	svc, err := SocialService(app, acc)
	if err != nil {
		return err
	}

	limiters, err := limits.Factory(app)
	if err != nil {
		return err
	}

	h := controllers.NewFollows(svc, acc.Web.Name())
	required, verified := auth.Required(acc.Web), auth.Verified(acc.Web)
	byUser := ratelimit.Middleware(limits.Follows.From(limiters), ratelimit.ByUser(acc.Web.Name()))
	r.Put("/u/{nickname}/follow", h.Follow).Name("follows.put").Use(required, verified, byUser)
	r.Delete("/u/{nickname}/follow", h.Unfollow).Name("follows.delete").Use(required, verified, byUser)
	r.Post("/me/follow-requests/{nickname}/accept", h.Accept).Name("follows.accept").Use(required, verified)
	r.Post("/me/follow-requests/{nickname}/decline", h.Decline).Name("follows.decline").Use(required, verified)
	r.Get("/feed", controllers.NewFeed(svc, acc.Web.Name()).Index).Name("feed").Use(required, verified)

	return nil
}

// SocialService builds the follow service shared by these routes and the public profile.
func SocialService(app *vukem.App, acc *accounts.Accounts) (*social.Follows, error) {
	ff, err := app.Flags()
	if err != nil {
		return nil, err
	}

	return &social.Follows{DB: acc.DB, Flags: ff}, nil
}
