package routes

import (
	"cmp"
	"context"
	"net/http"

	"github.com/AlexGalhardo/vukem"
	vapi "github.com/AlexGalhardo/vukem/api"
	"github.com/AlexGalhardo/vukem/auth"
	vconfig "github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/drivers/s3"
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

	for _, id := range []string{api.PlanFree, api.PlanHobby, api.PlanPro} {
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

	bucket, err := exportsBucket(app)
	if err != nil {
		return err
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
		)

		read := func(name string) vapi.Op {
			return vapi.Op{
				Name: name, Summary: v1Summaries[name], Description: v1Descriptions[name],
				Tags: []string{cmp.Or(v1Tags[name], "catalog")}, Security: []string{"bearer"},
				ValidateResponse: true,
			}
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
		for section, op := range map[string]string{
			"games": "getGameActivity", "movies": "getMovieActivity",
			"series": "getSeriesActivity", "books": "getBookActivity",
		} {
			vapi.Register(g, "GET /"+section+"/{id}/activity", read(op), items.GetItemActivity(section))
		}
		vapi.Register(g, "GET /search/entities", read("searchEntities"), items.SearchEntities)
		vapi.Register(g, "GET /search", read("searchItems"), items.SearchItems)
		vapi.Register(g, "GET /lookup", read("lookupItem"), items.LookupItem)
		vapi.Register(g, "GET /changes", read("listChanges"), items.ListChanges)
		// Literal /webhooks before /{type}, same reason as /{type}/genres.
		hooks := &apiv1.Webhooks{App: app, DB: acc.DB, Crypt: app.Crypt, Plans: plans}
		hook := func(name string) vapi.Op {
			return vapi.Op{
				Name: name, Summary: v1Summaries[name], Description: v1Descriptions[name],
				Tags: []string{"webhooks"}, Security: []string{"bearer"},
				ValidateResponse: true,
			}
		}
		vapi.Register(g, "GET /webhooks", hook("listWebhooks"), hooks.List)
		vapi.Register(g, "POST /webhooks", hook("createWebhook"), hooks.Create)
		vapi.Register(g, "DELETE /webhooks/{id}", hook("deleteWebhook"), hooks.Delete)
		// A test ping does not create a resource. POST otherwise defaults to 201.
		testHook := hook("testWebhook")
		testHook.Status = http.StatusOK
		vapi.Register(g, "POST /webhooks/{id}/test", testHook, hooks.Test)
		exports := &apiv1.Exports{Plans: plans, Bucket: bucket}
		exp := hook("listExports")
		exp.Tags = []string{"exports"}
		vapi.Register(g, "GET /exports", exp, exports.List)
		get := hook("getExport")
		get.Tags = []string{"exports"}
		vapi.Register(g, "GET /exports/{type}/{day}", get, exports.Get)
		vapi.Register(g, "GET /{type}/{id}", read("getItem"), items.GetItem)
		vapi.Register(g, "GET /{type}", read("listItems"), items.ListItems)
	})

	return nil
}

// exportsBucket is the S3 disk the daily ID exports live on, or nil without S3_BUCKET (development writes them to
// a local directory, which cannot presign).
func exportsBucket(app *vukem.App) (apiv1.Bucket, error) {
	if app.Source.Get("S3_BUCKET") == "" {
		return nil, nil
	}

	cfg, err := vconfig.Load[s3.Config](app.Source)
	if err != nil {
		return nil, err
	}

	return s3.New(cfg)
}

// v1Tags groups the operations of one category under its own section of the API reference (owner 2026-11-02);
// operations shared by the four categories stay under "catalog".
var v1Tags = map[string]string{
	"listGameGenres": "games", "listGameTrending": "games", "getGameActivity": "games",
	"listMovieGenres": "movies", "listMovieTrending": "movies", "getMovieActivity": "movies",
	"listSeriesGenres": "series", "listSeriesTrending": "series", "getSeriesActivity": "series",
	"listBookGenres": "books", "listBookTrending": "books", "getBookActivity": "books", "listEditions": "books",
}

// v1Descriptions say in one sentence what each operation does; the API reference shows it under the title.
var v1Descriptions = map[string]string{
	"getUsage":           "Returns the requests your API key used this month, its monthly quota and when the quota resets.",
	"listItems":          "Lists the items of one category (games, movies, series or books), with filters, sorting and pagination.",
	"getItem":            "Returns one item of a category by its ID or slug, with its score, details and credits.",
	"lookupItem":         "Finds an item by an external ID, such as a Wikidata QID, an IGDB or TMDB ID, or an ISBN.",
	"searchItems":        "Searches item titles across the four categories and returns the best matches first.",
	"searchEntities":     "Searches people, organizations, franchises, platforms and genres by name.",
	"getEntity":          "Returns one entity (person, organization, franchise, platform, genre and so on) by its ID.",
	"listEntityItems":    "Lists the items linked to an entity, for example the movies of a director or the games of a platform.",
	"getPerson":          "Returns one person (director, writer, author, actor and so on) by ID.",
	"getOrganization":    "Returns one organization (studio, publisher, network and so on) by ID.",
	"getFranchise":       "Returns one franchise by ID.",
	"getPlatform":        "Returns one game platform by ID.",
	"listChanges":        "Lists catalogue changes (created, updated and removed items) since a cursor, for keeping a copy in sync.",
	"listGameGenres":     "Lists the game genres, for filtering the game listing.",
	"listMovieGenres":    "Lists the movie genres, for filtering the movie listing.",
	"listSeriesGenres":   "Lists the series genres, for filtering the series listing.",
	"listBookGenres":     "Lists the book genres, for filtering the book listing.",
	"listGameTrending":   "Lists the games trending now, ranked by recent community activity.",
	"listMovieTrending":  "Lists the movies trending now, ranked by recent community activity.",
	"listSeriesTrending": "Lists the series trending now, ranked by recent community activity.",
	"listBookTrending":   "Lists the books trending now, ranked by recent community activity.",
	"getGameActivity":    "Returns the recent community activity of one game: ratings, reviews and list additions over time.",
	"getMovieActivity":   "Returns the recent community activity of one movie: ratings, reviews and list additions over time.",
	"getSeriesActivity":  "Returns the recent community activity of one series: ratings, reviews and list additions over time.",
	"getBookActivity":    "Returns the recent community activity of one book: ratings, reviews and list additions over time.",
	"listEditions":       "Lists the editions of one book (formats, languages, publishers and ISBNs).",
	"listWebhooks":       "Lists the webhooks of your API key.",
	"createWebhook":      "Creates a webhook that receives catalogue change events at your URL, signed with a secret.",
	"deleteWebhook":      "Deletes one of your webhooks.",
	"testWebhook":        "Sends a test event to one of your webhooks and reports the response.",
	"listExports":        "Lists the daily ID exports your plan can download.",
	"getExport":          "Returns a short-lived download link for one daily ID export.",
}

// v1Summaries are the readable operation titles of the API reference sidebar ("List items", not "listItems";
// owner 2026-11-02). TestAPIReference_EveryOperationHasASummary keeps every operation covered.
var v1Summaries = map[string]string{
	"getUsage": "Get usage", "listEditions": "List book editions", "listEntityItems": "List entity items",
	"getEntity": "Get entity", "getPerson": "Get person", "getOrganization": "Get organization",
	"getFranchise": "Get franchise", "getPlatform": "Get platform",
	"listGameGenres": "List game genres", "listMovieGenres": "List movie genres",
	"listSeriesGenres": "List series genres", "listBookGenres": "List book genres",
	"listGameTrending": "List trending games", "listMovieTrending": "List trending movies",
	"listSeriesTrending": "List trending series", "listBookTrending": "List trending books",
	"getGameActivity": "Get game activity", "getMovieActivity": "Get movie activity",
	"getSeriesActivity": "Get series activity", "getBookActivity": "Get book activity",
	"searchEntities": "Search entities", "searchItems": "Search items", "lookupItem": "Look up item",
	"listChanges": "List changes", "listWebhooks": "List webhooks", "createWebhook": "Create webhook",
	"deleteWebhook": "Delete webhook", "testWebhook": "Test webhook", "listExports": "List exports",
	"getExport": "Get export", "getItem": "Get item", "listItems": "List items",
}
