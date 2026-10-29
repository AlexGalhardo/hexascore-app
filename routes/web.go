package routes

import (
	"cmp"
	"context"
	"strings"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/controllers"
	appmiddleware "github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/notifications"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
	"github.com/AlexGalhardo/hexascore-app/app/seo"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
	"github.com/AlexGalhardo/hexascore-app/resources"
)

// localeQuery is the reserved parameter LocalePrefix hands to i18n.Middleware.
const localeQuery = "_locale"

// Web registers the browser routes. vukem route:list shows them.
func Web(r *router.Router, app *vukem.App) error {
	// workaround(vukem): VF-024. The URL prefix is the only locale source
	// i18n.Middleware reads; the locale cookie only drives the redirect (SDD 09 §7.1).
	err := vukem.DefaultMiddlewareWith(app, r, vukem.MiddlewareOptions{
		SecureHeaders: appmiddleware.SecureHeaders(app.Config.IsProduction()),
		I18n: &i18n.MiddlewareConfig{
			Query:   localeQuery,
			Order:   []i18n.Source{i18n.SourceQuery},
			Locales: []string{"en", "pt-BR", "es"},
		},
	})
	if err != nil {
		return err
	}

	r.UseServer(appmiddleware.CrossOriginResourcePolicy("same-site"))
	r.UseServer(middleware.StaticFS(resources.Public(), middleware.StaticConfig{Prefix: "/"}))
	r.UseServer(appmiddleware.LocalePrefix(appmiddleware.LocalePrefixConfig{
		Prefixes: map[string]string{"pt-br": "pt-BR", "es": "es"},
		Query:    localeQuery,
		Cookie:   "locale",
	}))

	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	db, err := app.DB()
	if err != nil {
		return err
	}

	// catalog_redirects exists only on Postgres (VF-033); SQLite scaffold tests skip it.
	if db.Dialect().Name() == "postgres" {
		r.UseServer(appmiddleware.CatalogRedirect(db, app.Log))
	}
	// Only the app catalogue namespaces reach the browser, never framework
	// messages or future server-only keys (SDD 09 §6.1).
	r.Use(
		appmiddleware.RestoreURL,
		app.Inertia.Middleware,
		inertia.I18n(
			"account.",
			"actions.",
			"api.",
			"auth.",
			"billing.",
			"catalog.",
			"categories.",
			"itemTypes.",
			"score.",
			"meta.",
			"nav.",
			"home.",
			"help.",
			"footer.",
			"cookies.",
			"sandbox.",
			"contact.",
			"docs.",
			"errors.",
			"item.",
			"updatedLabel.",
			"policy.",
			"terms.",
			"lists.",
			"rating.",
			"ratings.",
			"search.",
			"social.",
			"notifications.",
		),
		appmiddleware.ShareProps(appmiddleware.ShareConfig{
			AppName: app.Config.Name,
			SEO:     seoConfig(app),
			User:    func(c *web.Ctx) (any, error) { return acc.Web.User(c) },
		}),
		appmiddleware.NotBanned(acc.Web.User, acc.ProblemBase, app.Now),
	)
	// user_notifications exists only on Postgres (VF-033).
	if db.Dialect().Name() == "postgres" {
		r.Use(appmiddleware.UnreadNotifications(acc.Web.User, func(ctx context.Context, id int64) (int, error) {
			return notifications.Unread(ctx, db, id)
		}))
	}

	r.Get("/help/score", controllers.HelpScore).Name("help.score")
	r.Get("/", controllers.Home(app)).Name("home")
	// Least specific pattern: unmatched GET pages render the Error page with
	// the shell, since vukem's own hook only covers Inertia visits.
	r.Get("/{path...}", controllers.NotFound).Name("not-found")

	// catalog_* tables exist only on Postgres (VF-033).
	if db.Dialect().Name() == "postgres" {
		r.Get("/sitemap.xml", controllers.SitemapIndex(app)).Name("sitemap.index")
		r.Get("/sitemaps/{file}", controllers.Sitemap(app)).Name("sitemap.child")

		sc := controllers.NewSearch(app, search.New(db, nil), acc.Web.User)
		r.Get("/search", sc.Index).Name("search")
		r.Get("/search/suggest", sc.Suggest).Name("search.suggest")
		limiters, err := app.RateLimiters()
		if err != nil {
			return err
		}
		// Each click needs a search that returned the ref; this caps how
		// fast one client can mint searches and clicks to skew analytics.
		clicks := ratelimit.Middleware(limiters("search.events", 30, time.Minute), ratelimit.ByIP)
		r.Post("/search/events", sc.Event).Use(clicks).Name("search.events")

		items := controllers.NewItem(app, &scoring.Writer{DB: acc.DB, Events: app.Events, Now: app.Now}, acc.Web.User)
		for section, typ := range map[string]string{"games": "game", "movies": "movie", "series": "series", "books": "book"} {
			r.Get("/"+section+"/{slug}", items.Show(typ)).Name("item." + typ)
		}
		r.Get("/books/{slug}/editions/{id}", items.Edition).Name("item.edition")
		byUser := ratelimit.Middleware(limiters("item.report", 10, time.Minute), ratelimit.ByUser(acc.Web.Name()))
		r.Post("/items/{id}/report", items.Report).
			Name("item.report").
			Use(auth.Required(acc.Web), auth.Verified(acc.Web), byUser)
		r.Get("/api/search/suggest", controllers.LegacySuggest).Name("search.suggest.legacy")
	}

	return nil
}

// seoConfig derives canonical and hreflang URLs from APP_URL and the locale
// prefixes LocalePrefix serves (SDD 09 §7.1, §7.7).
func seoConfig(app *vukem.App) seo.Config {
	return seo.Config{
		BaseURL: strings.TrimRight(cmp.Or(app.Source.Get("APP_URL"), "https://hexascore.app"), "/"),
		Locales: []seo.Locale{{Code: "en"}, {Code: "pt-BR", Prefix: "/pt-br"}, {Code: "es", Prefix: "/es"}},
	}
}
