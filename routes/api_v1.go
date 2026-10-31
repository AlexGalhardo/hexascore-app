package routes

import (
	"context"
	"net/http"

	"github.com/AlexGalhardo/vukem"
	vapi "github.com/AlexGalhardo/vukem/api"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/api"
	apiv1 "github.com/AlexGalhardo/hexascore-app/app/api/v1"
	"github.com/AlexGalhardo/hexascore-app/app/billing"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/controllers/publicapi"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	appmiddleware "github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/config"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
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

	limiters, err := limits.Factory(app)
	if err != nil {
		return err
	}

	guard := &api.Guard{Keys: &api.Keys{DB: acc.DB, Now: app.Now}, User: acc.Users.ByID}
	meter := &api.Meter{Store: api.PGQuotaStore{DB: acc.DB}, Now: app.Now, Log: app.Log}
	app.AddRunner("api-usage", meter)

	plans := billing.PlanResolver(acc.DB, app.Now)

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
		Search: search.New(acc.DB, config.LoadCatalog(app.Source.Get).APISources),
		Redirects: func(ctx context.Context, typ, id string) (catalog.Target, bool, error) {
			return catalog.ResolveID(ctx, acc.DB, catalog.ScopeItem, typ, id)
		},
		Now: app.Now,
	}
	throttles := map[string]func(http.Handler) http.Handler{}

	for _, id := range []string{api.PlanFree, api.PlanPro, api.PlanBusiness} {
		p, _ := api.PlanByID(id)
		rule := limits.APIPlan(id, p.RPS, p.Burst).From(limiters)
		throttles[id] = ratelimit.Middleware(rule, ratelimit.ByUser(api.GuardName))
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
			ratelimit.Middleware(limits.APIIP.From(limiters), ratelimit.ByIP),
			auth.Required(guard),
			appmiddleware.NotBanned(func(c *web.Ctx) (models.User, error) {
				return auth.User[models.User](c, api.GuardName)
			}, acc.ProblemBase, app.Now),
			throttle,
			api.Quota(meter, plans, acc.ProblemBase),
			publicapi.TypeGuard,
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
		for section, op := range map[string]string{
			"games": "listGameTrending", "movies": "listMovieTrending",
			"series": "listSeriesTrending", "books": "listBookTrending",
		} {
			vapi.Register(g, "GET /"+section+"/trending", read(op), items.ListTrending(section))
		}
		vapi.Register(g, "GET /{type}/{id}/activity", read("getItemActivity"), items.GetItemActivity)
		vapi.Register(g, "GET /search/entities", read("searchEntities"), items.SearchEntities)
		vapi.Register(g, "GET /search", read("searchItems"), items.SearchItems)
		vapi.Register(g, "GET /lookup", read("lookupItem"), items.LookupItem)
		vapi.Register(g, "GET /changes", read("listChanges"), items.ListChanges)
		// Literal /webhooks before /{type}, same reason as /{type}/genres.
		hooks := &apiv1.Webhooks{App: app, DB: acc.DB, Crypt: app.Crypt, Plans: plans}
		hook := func(name string) vapi.Op {
			return vapi.Op{Name: name, Tags: []string{"webhooks"}, Security: []string{"bearer"}, ValidateResponse: true}
		}
		vapi.Register(g, "GET /webhooks", hook("listWebhooks"), hooks.List)
		vapi.Register(g, "POST /webhooks", hook("createWebhook"), hooks.Create)
		vapi.Register(g, "DELETE /webhooks/{id}", hook("deleteWebhook"), hooks.Delete)
		// A test ping does not create a resource. POST otherwise defaults to 201.
		testHook := hook("testWebhook")
		testHook.Status = http.StatusOK
		vapi.Register(g, "POST /webhooks/{id}/test", testHook, hooks.Test)
		vapi.Register(g, "GET /{type}/{id}", read("getItem"), items.GetItem)
		vapi.Register(g, "GET /{type}", read("listItems"), items.ListItems)
	})

	return nil
}
