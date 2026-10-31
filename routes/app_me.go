package routes

import (
	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/appapi"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
)

// AppMe registers GET/PATCH /app/v1/me and POST/DELETE /app/v1/me/deletion.
// POST /app/v1/me/export stays in Exports. GET requires catalog:read; writes
// require profile:write (SDD 03 §6.2).
func AppMe(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	limiters, err := limits.Factory(app)
	if err != nil {
		return err
	}

	h := appapi.NewMe(acc, acc.API.Name())
	read := AppAuthStack(app, acc, limiters, auth.Ability(acc.API, "catalog:read"))
	write, err := AppWriteStack(app, acc, limiters, auth.Ability(acc.API, "profile:write"))
	if err != nil {
		return err
	}

	schedule, err := AppWriteStack(app, acc, limiters,
		auth.Ability(acc.API, "profile:write"),
		ratelimit.Middleware(limits.Deletion.From(limiters), ratelimit.ByUser(acc.API.Name())),
	)
	if err != nil {
		return err
	}

	r.Group("/app/v1", func(g *router.Group) {
		g.Get("/me", h.Show).Name("me.show").Use(read...)
		g.Patch("/me", h.Update).Name("me.update").Use(write...)
		g.Post("/me/deletion", h.Schedule).Name("me.deletion.schedule").Use(schedule...)
		g.Delete("/me/deletion", h.Cancel).Name("me.deletion.cancel").Use(write...)
	}).Name("app.v1.")

	return nil
}
