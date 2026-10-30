package routes

import (
	"net/http"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/controllers"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	"github.com/AlexGalhardo/hexascore-app/app/library"
)

// Exports registers the data export request on the web (POST /profile/export) and on /app/v1 (ability
// profile:write), and serves the ZIPs only through signed 24 h links under /exports/ (SDD 14 §7.4). The table
// exists only on PostgreSQL (VF-033).
func Exports(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	if acc.DB.Dialect().Name() != "postgres" {
		return nil
	}

	svc, err := library.NewExports(app, acc)
	if err != nil {
		return err
	}

	limiters, err := limits.Factory(app)
	if err != nil {
		return err
	}

	// The 7-day cooldown is the real limit; this only stops request floods hitting the cooldown query.
	limit := limits.Exports.From(limiters)

	web := controllers.NewExports(svc, acc.Web.Name(), false, acc.ProblemBase)
	r.Post("/profile/export", web.Create).Name("profile.export").
		Use(auth.Required(acc.Web), auth.Verified(acc.Web), ratelimit.Middleware(limit, ratelimit.ByUser(acc.Web.Name())))

	api := controllers.NewExports(svc, acc.API.Name(), true, acc.ProblemBase)
	r.Group("/app/v1", func(g *router.Group) {
		g.Use(
			middleware.JSONOnly(),
			auth.Required(acc.API),
			auth.Verified(acc.API),
			auth.Ability(acc.API, "profile:write"),
			ratelimit.Middleware(limit, ratelimit.ByUser(acc.API.Name())),
		)
		g.Post("/me/export", api.Create).Name("me.export")
	}).Name("app.v1.")

	r.Handle(
		"GET "+library.ExportsPrefix+"/",
		http.StripPrefix(library.ExportsPrefix, storage.Handler(svc.Disk, svc.Signer)),
	)

	return nil
}
