package routes

import (
	"context"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/api"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/authz"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/controllers"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/policies"
)

// Lists registers the list write routes on the web (session) and on /app/v1 (token, ability lists:write), with
// the 60 per minute per-user limit (SDD 14 NFR-14-4) and Idempotency-Key on the API (SDD 14 §6.3).
func Lists(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	limiters, err := limits.Factory(app)
	if err != nil {
		return err
	}

	c, err := app.Cache()
	if err != nil {
		return err
	}

	// A caller with neither guard is anonymous (nil user): policies decide what that may see.
	gate := authz.New(func(ctx context.Context) (*models.User, error) {
		for _, g := range []string{acc.Web.Name(), acc.API.Name()} {
			if u, err := auth.UserFrom[models.User](ctx, g); err == nil {
				return &u, nil
			}
		}

		return nil, nil //nolint:nilnil // anonymous
	})
	if err := gate.Register(authz.For[*models.List](policies.ListPolicy{})); err != nil {
		return err
	}

	svc := &library.Lists{DB: acc.DB, Events: app.Events, Now: app.Now}
	limit := limits.Lists.From(limiters)

	web := controllers.NewLists(svc, gate, acc.Web.Name(), false)
	required, verified := auth.Required(acc.Web), auth.Verified(acc.Web)
	byUser := ratelimit.Middleware(limit, ratelimit.ByUser(acc.Web.Name()))
	r.Get("/lists", web.Index).Name("lists.index").Use(required, verified)
	r.Get("/lists/{id}/edit", web.Edit).Name("lists.edit").Use(required, verified)
	r.Get("/u/{nickname}/lists/{id}", web.Show).Name("lists.show").Use(auth.Optional(acc.Web))
	r.Post("/lists", web.Create).Name("lists.create").Use(required, verified, byUser)
	r.Patch("/lists/{id}", web.Update).Name("lists.update").Use(required, verified, byUser)
	r.Delete("/lists/{id}", web.Delete).Name("lists.delete").Use(required, verified, byUser)
	r.Put("/lists/{id}/items/{itemId}", web.AddItem).Name("lists.items.put").Use(required, verified, byUser)
	r.Delete("/lists/{id}/items/{itemId}", web.RemoveItem).Name("lists.items.delete").Use(required, verified, byUser)
	r.Put("/lists/{id}/order", web.Reorder).Name("lists.order").Use(required, verified, byUser)

	h := controllers.NewLists(svc, gate, acc.API.Name(), true)
	idem := api.Idempotency(c.Store(), api.IdempotencyOptions{Methods: []string{"POST", "PATCH", "PUT", "DELETE"}})
	r.Group("/app/v1", func(g *router.Group) {
		g.Use(middleware.JSONOnly(), auth.Required(acc.API), auth.Ability(acc.API, "lists:write"),
			ratelimit.Middleware(limit, ratelimit.ByUser(acc.API.Name())), idem)
		g.Post("/lists", h.Create).Name("lists.create")
		g.Patch("/lists/{id}", h.Update).Name("lists.update")
		g.Delete("/lists/{id}", h.Delete).Name("lists.delete")
		g.Put("/lists/{id}/items/{itemId}", h.AddItem).Name("lists.items.put")
		g.Delete("/lists/{id}/items/{itemId}", h.RemoveItem).Name("lists.items.delete")
		g.Put("/lists/{id}/order", h.Reorder).Name("lists.order")
	}).Name("app.v1.")

	return nil
}
