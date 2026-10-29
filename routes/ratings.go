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
	appmiddleware "github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// Ratings registers the rating write routes on the web (session) and on
// /app/v1 (token, ability ratings:write); both share one handler set per
// guard and the 30 per minute per-user limit (SDD 02 §6).
func Ratings(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	limiters, err := app.RateLimiters()
	if err != nil {
		return err
	}

	rating := limiters("rating", 30, time.Minute)
	w := &scoring.Writer{DB: acc.DB, Events: app.Events, Now: app.Now}

	web := controllers.NewRatings(w, acc.Web.Name(), false)
	required, verified := auth.Required(acc.Web), auth.Verified(acc.Web)
	byUser := ratelimit.Middleware(rating, ratelimit.ByUser(acc.Web.Name()))
	r.Put("/items/{id}/rating", appmiddleware.CodedProblem(acc.ProblemBase, web.Put)).
		Name("ratings.put").
		Use(required, verified, byUser)
	r.Delete("/items/{id}/rating", appmiddleware.CodedProblem(acc.ProblemBase, web.Delete)).
		Name("ratings.delete").
		Use(required, verified, byUser)

	api := controllers.NewRatings(w, acc.API.Name(), true)
	r.Group("/app/v1", func(g *router.Group) {
		g.Use(middleware.JSONOnly(), auth.Required(acc.API), auth.Ability(acc.API, "ratings:write"),
			ratelimit.Middleware(rating, ratelimit.ByUser(acc.API.Name())))
		g.Put("/items/{id}/rating", appmiddleware.CodedProblem(acc.ProblemBase, api.Put)).Name("ratings.put")
		g.Delete("/items/{id}/rating", appmiddleware.CodedProblem(acc.ProblemBase, api.Delete)).Name("ratings.delete")
	}).Name("app.v1.")

	return nil
}
