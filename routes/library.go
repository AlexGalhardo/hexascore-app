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
)

// Library registers the library, the edition choice and the public profile on the web and on /app/v1
// (SDD 14 §6.2, §6.3). The tables exist only on Postgres (VF-033).
func Library(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	if acc.DB.Dialect().Name() != "postgres" {
		return nil
	}

	limiters, err := app.RateLimiters()
	if err != nil {
		return err
	}

	limit := limiters("edition", 60, time.Minute)

	web := controllers.NewLibrary(acc.DB, app.Now, acc.Web.Name(), false)
	required, verified := auth.Required(acc.Web), auth.Verified(acc.Web)
	byUser := ratelimit.Middleware(limit, ratelimit.ByUser(acc.Web.Name()))
	r.Get("/library", web.Index).Name("library").Use(required, verified)
	r.Put("/items/{id}/edition", web.PutEdition).Name("edition.put").Use(required, verified, byUser)
	r.Delete("/items/{id}/edition", web.DeleteEdition).Name("edition.delete").Use(required, verified, byUser)

	profile := controllers.NewPublicProfile(acc.DB, false, acc.Web.Name())
	r.Get("/u/{nickname}", profile.Show).Name("profile.public").Use(auth.Optional(acc.Web))

	h := controllers.NewLibrary(acc.DB, app.Now, acc.API.Name(), true)
	pub := controllers.NewPublicProfile(acc.DB, true, acc.API.Name())
	r.Group("/app/v1", func(g *router.Group) {
		g.Use(
			middleware.JSONOnly(),
			auth.Required(acc.API),
			ratelimit.Middleware(limit, ratelimit.ByUser(acc.API.Name())),
		)
		g.Get("/me/library", h.Index).Name("library").Use(auth.Ability(acc.API, "catalog:read"))
		g.Get("/users/{nickname}", pub.Show).Name("users.show").Use(auth.Ability(acc.API, "catalog:read"))
		g.Put("/items/{id}/edition", h.PutEdition).Name("edition.put").Use(auth.Ability(acc.API, "status:write"))
		g.Delete("/items/{id}/edition", h.DeleteEdition).
			Name("edition.delete").
			Use(auth.Ability(acc.API, "status:write"))
	}).Name("app.v1.")

	return nil
}
