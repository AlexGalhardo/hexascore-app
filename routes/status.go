package routes

import (
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/controllers"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// Status registers the item status and time-to-beat write routes on the web (session) and on /app/v1
// (token, ability status:write) with legacy-parity limits: 60 status and 30 time-to-beat writes per minute.
func Status(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	limiters, err := app.RateLimiters()
	if err != nil {
		return err
	}

	statusLimit, ttbLimit := limiters("status", 60, time.Minute), limiters("time-to-beat", 30, time.Minute)
	w := &scoring.Writer{DB: acc.DB, Events: app.Events, Now: app.Now}

	web := controllers.NewStatus(w, acc.Web.Name(), false)
	required, verified := auth.Required(acc.Web), auth.Verified(acc.Web)
	byUser := func(l ratelimit.Limiter) router.Middleware {
		return ratelimit.Middleware(l, ratelimit.ByUser(acc.Web.Name()))
	}
	r.Put("/items/{id}/status", web.PutStatus).Name("status.put").Use(required, verified, byUser(statusLimit))
	r.Delete("/items/{id}/status", web.DeleteStatus).Name("status.delete").Use(required, verified, byUser(statusLimit))
	r.Put("/items/{id}/time-to-beat", web.PutTimeToBeat).
		Name("time-to-beat.put").
		Use(required, verified, byUser(ttbLimit))
	r.Delete("/items/{id}/time-to-beat", web.DeleteTimeToBeat).
		Name("time-to-beat.delete").
		Use(required, verified, byUser(ttbLimit))

	api := controllers.NewStatus(w, acc.API.Name(), true)
	byAPIUser := func(l ratelimit.Limiter) router.Middleware {
		return ratelimit.Middleware(l, ratelimit.ByUser(acc.API.Name()))
	}
	r.Group("/app/v1", func(g *router.Group) {
		g.Use(middleware.JSONOnly(), auth.Required(acc.API))
		g.Get("/items/{id}/me", api.Me).Name("me")
		g.Group("", func(s *router.Group) {
			s.Use(auth.Ability(acc.API, "status:write"))
			s.Put("/items/{id}/status", api.PutStatus).Name("status.put").Use(byAPIUser(statusLimit))
			s.Delete("/items/{id}/status", api.DeleteStatus).Name("status.delete").Use(byAPIUser(statusLimit))
			s.Put("/items/{id}/time-to-beat", api.PutTimeToBeat).Name("time-to-beat.put").Use(byAPIUser(ttbLimit))
			s.Delete("/items/{id}/time-to-beat", api.DeleteTimeToBeat).
				Name("time-to-beat.delete").
				Use(byAPIUser(ttbLimit))
		})
	}).Name("app.v1.")

	return nil
}
