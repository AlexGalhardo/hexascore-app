package routes

import (
	"context"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/appapi"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/controllers/publicapi"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
)

// AppCatalog mounts the P3.A catalogue reads on /app/v1 with audience site
// (SDD 05 §6.4). The group name is app.v1. so these route names do not collide
// with /api/v1. Trending stays unmounted until P3.A.1.4.
func AppCatalog(r *router.Router, app *vukem.App) error {
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

	key, err := app.DeriveKey("catalog.cursor", 32)
	if err != nil {
		return err
	}
	codec, err := catalog.NewCursorCodec(key)
	if err != nil {
		return err
	}

	items := &publicapi.Controller{
		DB:      acc.DB,
		Policy:  catalog.NewPolicy(catalog.AudienceSite, nil),
		Codec:   codec,
		Mount:   "/app/v1",
		PageCap: catalog.MaxLimit,
		Redirects: func(ctx context.Context, typ, id string) (catalog.Target, bool, error) {
			return catalog.ResolveID(ctx, acc.DB, catalog.ScopeItem, typ, id)
		},
		Now: app.Now,
	}

	stack := AppAuthStack(app, acc, limiters, auth.Ability(acc.API, "catalog:read"))
	r.Group("/app/v1", func(g *router.Group) {
		g.Use(appapi.KnownType)
		g.Use(stack...)
		appapi.Catalog(g, items, catalog.CoversBase(app.Source.Get("COVERS_BASE_URL")))
	}).Name("app.v1.")

	return nil
}
