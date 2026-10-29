// Package bootstrap wires the application together. There is no hidden
// registration: everything the app uses is listed here.
package bootstrap

import (
	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/app/listeners"
	"github.com/AlexGalhardo/hexascore-app/app/ssr"
	"github.com/AlexGalhardo/hexascore-app/database/migrations"
	"github.com/AlexGalhardo/hexascore-app/database/seeders"
	"github.com/AlexGalhardo/hexascore-app/resources"
	"github.com/AlexGalhardo/hexascore-app/routes"
)

// New builds the application; cmd/app and the tests call it.
func New(opts ...vukem.Option) (*vukem.App, error) {
	return NewWithSSRFailures(&ssr.Failures{}, opts...)
}

// NewWithSSRFailures is New with the fallback counter supplied, so tests can
// read it (the OTLP exporter for inertia_ssr_fallback_total comes in P2.C).
func NewWithSSRFailures(ssrFailures *ssr.Failures, opts ...vukem.Option) (*vukem.App, error) {
	base := []vukem.Option{
		vukem.WithProviders(Providers()...),
		vukem.WithDatabaseDrivers(postgres.Driver()),
		vukem.WithMigrations(migrations.All()),
		vukem.WithSeeders(seeders.All()...),
		vukem.WithListeners(listeners.All()...),
		vukem.WithJobsFunc(jobs.All),
		vukem.WithSchedule(routes.Schedule),
		vukem.WithViews(resources.Views()),
		vukem.WithVite(resources.Public()),
		vukem.WithInertia(inertia.Config{
			SSR: &inertia.SSRConfig{
				Bundle:    "bootstrap/ssr/ssr.js",
				OnFailure: ssrFailures.Observe,
				// Private pages skip SSR: no SEO value, user data stays out of the Node process (SDD 09 §6.1).
				Except: []string{"/admin", "/profile", "/me", "/billing", "/suggest", "/suggestions"},
			},
		}),
	}

	base = append(base, vukem.WithRoutes(func(_ *router.Router, app *vukem.App) error {
		app.AddHealthCheck("ssr", app.Inertia.SSRHealthy)
		return nil
	}))

	for _, fn := range routes.All() {
		base = append(base, vukem.WithRoutes(fn))
	}

	app, err := vukem.New(append(base, opts...)...)
	if err != nil {
		return nil, err
	}

	app.AddCommand(Commands()...)

	return app, nil
}
