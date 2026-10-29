package routes

import (
	"net/http"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/controllers"
	"github.com/AlexGalhardo/hexascore-app/resources"
)

// Web registers the browser routes. vukem route:list shows them.
func Web(r *router.Router, app *vukem.App) error {
	if err := vukem.DefaultMiddleware(app, r); err != nil {
		return err
	}

	r.UseServer(middleware.StaticFS(resources.Public(), middleware.StaticConfig{Prefix: "/"}))
	// Only the app catalogue namespaces reach the browser, never framework
	// messages or future server-only keys (SDD 09 §6.1).
	r.Use(app.Inertia.Middleware, inertia.I18n("account.", "actions.", "api.", "auth.", "billing.", "catalog.", "categories.", "itemTypes.", "score.", "meta.", "nav.", "home.", "footer.", "cookies.", "sandbox.", "contact.", "docs.", "errors.", "item.", "updatedLabel.", "policy.", "terms.", "lists.", "ratings.", "search.", "social."), shareProps(app))

	r.Get("/", controllers.Home).Name("home")

	return nil
}

// shareProps adds the props every page receives (usePage().props).
func shareProps(app *vukem.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			inertia.Share(r.Context(), "appName", app.Config.Name)
			next.ServeHTTP(w, r)
		})
	}
}
