package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/AlexGalhardo/vukem"
	vapi "github.com/AlexGalhardo/vukem/api"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/controllers/publicapi"
	appmiddleware "github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/config"
)

// APIV1 mounts the /api/v1 group with the SDD 05 §6.1 stack: IP limit 120/min, key guard (401/503), ban (403),
// per-plan throttle (429) and the quota meter (402). Read endpoints are typed routes (P1.H.2.1). The meter flushes
// every minute and on shutdown as an app runner, so it lives in the serving process (ADR 0003). The tables
// exist only on PostgreSQL (VF-033).
func APIV1(r *router.Router, app *vukem.App) error {
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

	guard := &api.Guard{Keys: &api.Keys{DB: acc.DB, Now: app.Now}, User: acc.Users.ByID}
	meter := &api.Meter{Store: api.PGQuotaStore{DB: acc.DB}, Now: app.Now, Log: app.Log}
	app.AddRunner("api-usage", meter)

	plans := api.PlanResolver(api.FreeForAll)

	key, err := app.DeriveKey("catalog.cursor", 32)
	if err != nil {
		return err
	}
	codec, err := catalog.NewCursorCodec(key)
	if err != nil {
		return err
	}
	items := &publicapi.Controller{
		DB:     acc.DB,
		Policy: catalog.NewPolicy(catalog.AudienceAPI, config.LoadCatalog(app.Source.Get).APISources),
		Codec:  codec,
		Meter:  meter,
		Plans:  plans,
		Redirects: func(ctx context.Context, typ, id string) (catalog.Target, bool, error) {
			return catalog.ResolveID(ctx, acc.DB, catalog.ScopeItem, typ, id)
		},
	}
	throttles := map[string]func(http.Handler) http.Handler{}

	for _, id := range []string{api.PlanFree, api.PlanPro, api.PlanBusiness} {
		p, _ := api.PlanByID(id)
		throttles[id] = ratelimit.Middleware(limiters("api-"+id, p.RPS, time.Second), ratelimit.ByUser(api.GuardName))
	}

	throttle := func(next http.Handler) http.Handler {
		byPlan := map[string]http.Handler{}
		for id, mw := range throttles {
			byPlan[id] = mw(next)
		}

		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			c := web.NewCtx(w, req, web.ServicesFrom(req.Context()))

			u, err := auth.User[models.User](c, api.GuardName)
			if err != nil {
				c.HandleError(err)
				return
			}

			p, err := plans(req.Context(), u.ID)
			if err != nil {
				c.HandleError(api.Unavailable(err))
				return
			}

			byPlan[p.ID].ServeHTTP(w, req)
		})
	}

	r.Group("/api/v1", func(g *router.Group) {
		g.Use(
			middleware.JSONOnly(),
			ratelimit.Middleware(limiters("api-ip", 120, time.Minute), ratelimit.ByIP),
			auth.Required(guard),
			appmiddleware.NotBanned(func(c *web.Ctx) (models.User, error) {
				return auth.User[models.User](c, api.GuardName)
			}, acc.ProblemBase, app.Now),
			throttle,
			api.Quota(meter, plans, acc.ProblemBase),
		)

		read := func(name string) vapi.Op {
			return vapi.Op{Name: name, Tags: []string{"catalog"}, Security: []string{"bearer"}, ValidateResponse: true}
		}
		vapi.Register(g, "GET /me/usage", read("getUsage"), items.GetUsage)
		vapi.Register(g, "GET /books/{id}/editions", read("listEditions"), items.ListEditions)
		vapi.Register(g, "GET /entities/{id}/items", read("listEntityItems"), items.ListEntityItems)
		vapi.Register(g, "GET /entities/{id}", read("getEntity"), items.GetEntity)
		vapi.Register(g, "GET /people/{id}", read("getPerson"), items.GetPerson)
		vapi.Register(g, "GET /organizations/{id}", read("getOrganization"), items.GetOrganization)
		vapi.Register(g, "GET /franchises/{id}", read("getFranchise"), items.GetFranchise)
		vapi.Register(g, "GET /platforms/{id}", read("getPlatform"), items.GetPlatform)
		// Literal sections: "/{type}/genres" overlaps "/entities/{id}" and the ServeMux rejects the pair.
		for section, op := range map[string]string{
			"games": "listGameGenres", "movies": "listMovieGenres",
			"series": "listSeriesGenres", "books": "listBookGenres",
		} {
			vapi.Register(g, "GET /"+section+"/genres", read(op), items.ListGenres(section))
		}
		vapi.Register(g, "GET /{type}/{id}", read("getItem"), items.GetItem)
		vapi.Register(g, "GET /{type}", read("listItems"), items.ListItems)
	})

	return nil
}
