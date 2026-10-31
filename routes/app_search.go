package routes

import (
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/appapi"
	"github.com/AlexGalhardo/hexascore-app/app/controllers"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	appmiddleware "github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// AppSearch mounts /search, /search/suggest and POST /search/events on
// /app/v1. A token is optional (SDD 10 §6.1); the audience is site.
func AppSearch(r *router.Router, app *vukem.App) error {
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

	h := controllers.NewSearch(app, search.New(acc.DB, nil), acc.API.User)
	r.Group("/app/v1", func(g *router.Group) {
		g.Use(
			middleware.JSONOnly(),
			auth.Optional(acc.API),
			appmiddleware.NotBanned(acc.API.User, acc.ProblemBase, app.Now),
		)
		appapi.Search(g, h,
			ratelimit.Middleware(limits.Suggest.From(limiters), ratelimit.ByIP),
			ratelimit.Middleware(limits.SearchEvents.From(limiters), ratelimit.ByIP),
			ratelimit.Middleware(limits.Search.From(limiters), ratelimit.ByIP),
		)
	}).Name("app.v1.")

	return nil
}
