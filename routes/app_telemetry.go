package routes

import (
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/appapi"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
)

// AppTelemetry mounts anonymous crash and web-vital intake on /app/v1.
// There is no bearer check. CSRF and session bypass live in Web.
func AppTelemetry(r *router.Router, app *vukem.App) error {
	db, err := app.DB()
	if err != nil {
		return err
	}

	if db.Dialect().Name() != "postgres" {
		return nil
	}

	limiters, err := limits.Factory(app)
	if err != nil {
		return err
	}

	r.Group("/app/v1", func(g *router.Group) {
		g.Use(middleware.JSONOnly())
		g.Post("/client-errors", appapi.ClientError(db, app.Now)).
			Use(ratelimit.Middleware(limits.ClientErrors.From(limiters), ratelimit.ByIP)).
			Name("client-errors")
		g.Post("/client-metrics", appapi.ClientMetric(db, app.Now)).
			Use(ratelimit.Middleware(limits.ClientMetrics.From(limiters), ratelimit.ByIP)).
			Name("client-metrics")
	}).Name("app.v1.")

	return nil
}
