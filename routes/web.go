package routes

import (
	"cmp"
	"context"
	"net/url"
	"strings"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/controllers"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	appmiddleware "github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/notifications"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/primitives"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/related"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
	"github.com/AlexGalhardo/hexascore-app/app/seo"
	"github.com/AlexGalhardo/hexascore-app/app/wellknown"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
	"github.com/AlexGalhardo/hexascore-app/resources"
)

// localePrefixes maps the URL prefixes to locales (SDD 09 §7.1).
var localePrefixes = map[string]string{"pt-br": "pt-BR", "es": "es"}

// Web registers the browser routes. vukem route:list shows them.
func Web(r *router.Router, app *vukem.App) error {
	// Covers come from the bucket origin when COVERS_BASE_URL is set; that origin joins img-src.
	coversBase := catalog.CoversBase(app.Source.Get("COVERS_BASE_URL"))

	var imgHosts []string
	if u, err := url.Parse(coversBase); err == nil && u.Scheme == "https" && u.Host != "" {
		imgHosts = append(imgHosts, u.Scheme+"://"+u.Host)
	}

	// Server level too: MethodOverride parses urlencoded forms before routing. A limit set
	// only on the router is too late, and vukem would read up to 10 MiB. v0.34.0 then
	// answers 413 from that earlier MaxBytesError (VF-061).
	r.UseServer(limits.Body)

	// The URL prefix is the only locale source; the locale cookie only drives the redirect (SDD 09 §7.1).
	err := vukem.DefaultMiddlewareWith(app, r, vukem.MiddlewareOptions{
		SecureHeaders: appmiddleware.SecureHeaders(
			appmiddleware.UpgradeInsecure(app.Config.IsProduction(), app.Config.URL),
			imgHosts...,
		),
		// The browser posts violation reports with no session and no CSRF token.
		SessionBypass: []string{
			"POST /csp-reports",
			"POST /app/v1/client-errors",
			"POST /app/v1/client-metrics",
		},
		CSRFBypass: []string{
			"POST /csp-reports",
			"POST /stripe/webhook",
			"POST /app/v1/client-errors",
			"POST /app/v1/client-metrics",
		},
		I18n: &i18n.MiddlewareConfig{
			Order:        []i18n.Source{i18n.SourcePath},
			Locales:      []string{"en", "pt-BR", "es"},
			PathPrefixes: localePrefixes,
		},
	})
	if err != nil {
		return err
	}
	// Router level: replaces vukem's 10 MiB router default, which replaced the server-level limit.
	r.Use(limits.Body)

	r.UseServer(appmiddleware.CrossOriginResourcePolicy("same-site"))
	// Outside MovePageScript so the cache keeps the rewritten document. vukem's default
	// session cookie name; the app does not override it (production adds a __Host- prefix).
	// Outside RenderCache so item pages served from the cache still count as views (SDD 08 §7.1).
	r.UseServer(primitives.For(app).Middleware("vukem_session", localePrefixes))
	r.UseServer(appmiddleware.RenderCache("vukem_session", localePrefixes))
	// First UseServer is outermost. MovePageScript rewrites the finished HTML:
	// LCP text before the page JSON, and the module graph after first paint.
	// Gzip has to wrap StaticFS, or the file server answers /build and the
	// compressor never sees the body.
	r.UseServer(appmiddleware.MovePageScript)
	r.UseServer(appmiddleware.GzipStatic())
	r.UseServer(middleware.StaticFS(resources.Public(), middleware.StaticConfig{Prefix: "/"}))
	r.UseServer(appmiddleware.LocaleCookieRedirect(appmiddleware.LocalePrefixConfig{
		Prefixes: localePrefixes,
		Cookie:   "locale",
	}))
	// Innermost server middleware: overwrites the web header set on /api/v1.
	r.UseServer(appmiddleware.APIHeaders)

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
		app.Inertia.Middleware,
		appmiddleware.PageI18n([]string{
			"account.",
			"actions.",
			"auth.",
			"catalog.",
			"categories.",
			"itemTypes.",
			"score.",
			"meta.",
			"nav.",
			"home.",
			"footer.",
			"cookies.",
			"sandbox.",
			"entity.",
			"errors.",
			"item.",
			"updatedLabel",
			"lists.",
			"rating.",
			"ratings.",
			"search.",
			"social.",
			"notifications.",
			"library.",
			"forYou.",
			"userLists.",
			"community.",
		}, localePrefixes),
		appmiddleware.ShareProps(appmiddleware.ShareConfig{
			AppName:    cmp.Or(app.Source.Get("APP_NAME"), "HexaScore"),
			CoversBase: coversBase,
			SEO:        seoConfig(app),
			User:       func(c *web.Ctx) (any, error) { return acc.Web.User(c) },
		}),
		appmiddleware.NotBanned(acc.Web.User, acc.ProblemBase, app.Now),
	)
	// user_notifications exists only on Postgres (VF-033).
	if db.Dialect().Name() == "postgres" {
		r.Use(appmiddleware.UnreadNotifications(acc.Web.User, func(ctx context.Context, id int64) (int, error) {
			return notifications.Unread(ctx, db, id)
		}))
	}

	docsJSON, err := resources.DocsJSON()
	if err != nil {
		return err
	}

	docs, err := controllers.Docs(docsJSON)
	if err != nil {
		return err
	}

	r.Get("/help/score", controllers.HelpScore).Name("help.score")
	r.Get("/robots.txt", controllers.Robots(app)).Name("robots")
	r.Get("/.well-known/assetlinks.json", wellknown.AssetLinks(
		cmp.Or(app.Source.Get("ANDROID_PACKAGE"), "app.hexascore"),
		app.Source.Get("ANDROID_CERT_SHA256"),
	)).Name("wellknown.assetlinks")
	r.Get("/terms", controllers.Legal("terms")).Name("terms")
	r.Get("/policy", controllers.Legal("policy")).Name("policy")
	r.Get("/pricing", controllers.Pricing).Name("pricing")
	r.Get("/docs", docs).Name("docs")
	r.Get("/docs/{slug}", docs).Name("docs.page")
	r.Get("/", controllers.Home(app)).Name("home")
	r.Get("/covers/{key...}", controllers.Covers(app)).Name("covers")
	// Least specific pattern: unmatched GET pages render the Error page with
	// the shell, since vukem's own hook only covers Inertia visits.
	r.Get("/{path...}", controllers.NotFound).Name("not-found")

	cspLimiters, err := limits.Factory(app)
	if err != nil {
		return err
	}

	r.Post("/csp-reports", appmiddleware.CSPReport).
		Use(ratelimit.Middleware(limits.CSPReports.From(cspLimiters), ratelimit.ByIP)).
		Name("csp-reports")

	// catalog_* tables exist only on Postgres (VF-033).
	if db.Dialect().Name() == "postgres" {
		r.Get("/sitemap.xml", controllers.SitemapIndex(app)).Name("sitemap.index")
		r.Get("/sitemaps/{file}", controllers.Sitemap(app)).Name("sitemap.child")

		sc := controllers.NewSearch(app, search.New(db, nil), acc.Web.User)
		limiters, err := limits.Factory(app)
		if err != nil {
			return err
		}

		suggest := ratelimit.Middleware(limits.Suggest.From(limiters), ratelimit.ByIP)
		r.Get("/search", sc.Index).
			Name("search").
			Use(ratelimit.Middleware(limits.Search.From(limiters), ratelimit.ByIP))
		r.Get("/search/suggest", sc.Suggest).Name("search.suggest").Use(suggest)
		// Each click needs a search that returned the ref; this caps how
		// fast one client can mint searches and clicks to skew analytics.
		clicks := ratelimit.Middleware(limits.SearchEvents.From(limiters), ratelimit.ByIP)
		r.Post("/search/events", sc.Event).Use(clicks).Name("search.events")

		listings, err := controllers.NewCatalog(app)
		if err != nil {
			return err
		}
		reco := related.NewMetrics()
		items := controllers.NewItem(
			app,
			&scoring.Writer{DB: acc.DB, Events: app.Events, Now: app.Now},
			acc.Web.User,
			reco,
		)
		changes := controllers.NewRecentChanges(app)
		for section, typ := range map[string]string{"games": "game", "movies": "movie", "series": "series", "books": "book"} {
			r.Get("/"+section, listings.Index(typ)).Name("catalog." + typ)
			r.Get("/"+section+"/{slug}", items.Show(typ)).Name("item." + typ)
			r.Get("/"+section+"/{slug}/changes", changes.Item(typ)).Name("item." + typ + ".changes")
		}
		r.Get("/books/{slug}/editions/{id}", items.Edition).Name("item.edition")
		r.Get("/books/{slug}/editions/{id}/changes", changes.Edition()).Name("item.edition.changes")
		r.Get("/api/items/{id}/ratings", items.Ratings).Name("item.ratings")
		entities := controllers.NewEntity(app)
		for _, e := range []struct{ path, kind, param string }{
			{"people", "person", "slug"},
			{"companies", "organization", "slug"},
			{"franchises", "franchise", "slug"},
			{"platforms", "platform", "slug"},
			{"genres", "genre", "slug"},
			{"game-modes", "game_mode", "slug"},
			{"countries", "country", "code"},
			{"languages", "language", "code"},
		} {
			r.Get("/"+e.path+"/{"+e.param+"}", entities.Show(e.kind, e.param)).Name("entity." + e.kind)
			r.Get("/"+e.path+"/{"+e.param+"}/changes", changes.Entity(e.kind, e.param)).
				Name("entity." + e.kind + ".changes")
		}
		r.Get("/api/search/suggest", controllers.LegacySuggest).Name("search.suggest.legacy").Use(suggest)
	}

	return nil
}

// seoConfig derives canonical and hreflang URLs from APP_URL and the locale
// prefixes localePrefixes serves (SDD 09 §7.1, §7.7).
func seoConfig(app *vukem.App) seo.Config {
	return seo.Config{
		BaseURL: strings.TrimRight(cmp.Or(app.Source.Get("APP_URL"), "https://hexascore.app"), "/"),
		Locales: []seo.Locale{{Code: "en"}, {Code: "pt-BR", Prefix: "/pt-br"}, {Code: "es", Prefix: "/es"}},
	}
}
