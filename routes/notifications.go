package routes

import (
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/controllers"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
)

// Notifications registers the inbox pages and /app/v1/me/notifications (SDD 14 §6.2, §6.3).
// user_notifications exists only on Postgres (VF-033).
func Notifications(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	if acc.DB.Dialect().Name() != "postgres" {
		return nil
	}

	limiters, err := limits.Factory(app)
	if err != nil {
		return err
	}

	limit := limits.Notify.From(limiters)

	web := controllers.NewNotifications(acc.DB, acc.Web.Name(), false)
	required, verified := auth.Required(acc.Web), auth.Verified(acc.Web)
	byUser := ratelimit.Middleware(limit, ratelimit.ByUser(acc.Web.Name()))
	r.Get("/me/notifications", web.Index).Name("notifications").Use(required, verified)
	r.Post("/me/notifications/read-all", web.ReadAll).Name("notifications.read-all").Use(required, verified, byUser)
	r.Post("/me/notifications/{id}/read", web.Read).Name("notifications.read").Use(required, verified, byUser)

	h := controllers.NewNotifications(acc.DB, acc.API.Name(), true)
	r.Group("/app/v1", func(g *router.Group) {
		g.Use(middleware.JSONOnly(), auth.Required(acc.API), auth.Verified(acc.API),
			ratelimit.Middleware(limit, ratelimit.ByUser(acc.API.Name())))
		g.Get("/me/notifications", h.Index).Name("notifications.index").Use(auth.Ability(acc.API, "catalog:read"))
		g.Post("/me/notifications/read", h.ReadMany).
			Name("notifications.read").
			Use(auth.Ability(acc.API, "profile:write"))
	}).Name("app.v1.me.")

	return nil
}
