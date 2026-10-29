package routes

import (
	"net/http"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/controllers"
	appmiddleware "github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
	"github.com/AlexGalhardo/hexascore-app/resources"
)

// localeQuery is the reserved parameter LocalePrefix hands to i18n.Middleware.
const localeQuery = "_locale"

// Web registers the browser routes. vukem route:list shows them.
func Web(r *router.Router, app *vukem.App) error {
	// workaround(vukem): VF-024. The URL prefix is the only locale source
	// i18n.Middleware reads; the locale cookie only drives the redirect (SDD 09 §7.1).
	err := vukem.DefaultMiddlewareWith(app, r, vukem.MiddlewareOptions{I18n: &i18n.MiddlewareConfig{
		Query:   localeQuery,
		Order:   []i18n.Source{i18n.SourceQuery},
		Locales: []string{"en", "pt-BR", "es"},
	}})
	if err != nil {
		return err
	}

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
			"ratings.",
			"search.",
			"social.",
		),
		shareProps(app),
		appmiddleware.NotBanned(acc.Web.User, acc.ProblemBase, app.Now),
	)

	r.Get("/", controllers.Home).Name("home")

	// catalog_* tables exist only on Postgres (VF-033).
	if db.Dialect().Name() == "postgres" {
		r.Get("/sitemap.xml", controllers.SitemapIndex(app)).Name("sitemap.index")
		r.Get("/sitemaps/{file}", controllers.Sitemap(app)).Name("sitemap.child")

		sc := controllers.NewSearch(app, search.New(db, nil), acc.Web.User)
		r.Get("/search", sc.Index).Name("search")
		r.Get("/search/suggest", sc.Suggest).Name("search.suggest")
		r.Post("/search/events", sc.Event).Name("search.events")
		r.Get("/api/search/suggest", controllers.LegacySuggest).Name("search.suggest.legacy")
	}

	return nil
}

// shareProps adds the props every page receives (usePage().props).
func shareProps(app *vukem.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			inertia.Share(r.Context(), "appName", app.Config.Name)

			// The root view reads it as .Page.Props.theme (VF-028: custom view
			// funcs are not request-bound), so only known values pass.
			if c, err := r.Cookie("theme"); err == nil && (c.Value == "dark" || c.Value == "light") {
				inertia.Share(r.Context(), "theme", c.Value)
			}
			next.ServeHTTP(w, r)
		})
	}
}
