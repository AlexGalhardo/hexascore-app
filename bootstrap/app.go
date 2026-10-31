// Package bootstrap wires the application together. There is no hidden
// registration: everything the app uses is listed here.
package bootstrap

import (
	"cmp"
	"context"
	"log/slog"
	"os"
	"strconv"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/openapi"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/schedule"
	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/i18nhash"
	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/listeners"
	"github.com/AlexGalhardo/hexascore-app/app/logging"
	"github.com/AlexGalhardo/hexascore-app/app/metrics"
	"github.com/AlexGalhardo/hexascore-app/app/scheduler"
	"github.com/AlexGalhardo/hexascore-app/app/ssr"
	"github.com/AlexGalhardo/hexascore-app/config"
	"github.com/AlexGalhardo/hexascore-app/database/migrations"
	"github.com/AlexGalhardo/hexascore-app/database/seeders"
	"github.com/AlexGalhardo/hexascore-app/resources"
	"github.com/AlexGalhardo/hexascore-app/routes"
)

// ReadyCheckTimeout bounds each /up?checks=1 check (SDD 12 §7.1), so the
// probe answers within Railway's healthcheck window even with a hung dependency.
const ReadyCheckTimeout = 2 * time.Second

// New builds the application; cmd/app and the tests call it.
func New(opts ...vukem.Option) (*vukem.App, error) {
	return NewWithSSRFailures(&ssr.Failures{}, opts...)
}

// NewWithSSRFailures is New with the fallback counter supplied, so tests can
// read it (the OTLP exporter for inertia_ssr_fallback_total comes in P2.C).
func NewWithSSRFailures(ssrFailures *ssr.Failures, opts ...vukem.Option) (*vukem.App, error) {
	langHash, err := i18nhash.Catalogs(resources.Lang())
	if err != nil {
		return nil, err
	}

	// The asset version also changes when only a catalogue changes, so the
	// once-prop messages are refetched after a deploy (SDD 09 §6.1).
	var built *vukem.App
	version := func() string {
		if built == nil || built.Vite == nil {
			return langHash
		}
		return built.Vite.Version() + "." + langHash
	}

	heartbeat := scheduler.NewHeartbeat(time.Now)

	base := []vukem.Option{
		vukem.WithProviders(Providers()...),
		vukem.WithDatabaseDrivers(postgres.Driver()),
		vukem.WithMigrations(migrations.All()),
		vukem.WithSeeders(seeders.All()...),
		vukem.WithListeners(listeners.All()...),
		vukem.WithJobsFunc(jobs.All),
		vukem.WithSchedule(func(s *schedule.Schedule, app *vukem.App) {
			routes.Schedule(s, app)
			s.Func("heartbeat", scheduler.Counted("heartbeat", heartbeat.Beat)).EveryMinute()
		}),
		vukem.WithFlags(community.SubmissionsFlag, community.TrustedModeratorsFlag),
		vukem.WithViews(resources.Views()),
		vukem.WithVite(resources.Public()),
		vukem.WithTranslations(resources.Lang()),
		// Routes keep the /api/v1 prefix; the api host also serves them as /v1 (APIEdge), so the host alone is the server.
		vukem.WithOpenAPI(
			openapi.Info{Title: "HexaScore API", Servers: []string{"https://" + config.APIHost(os.Getenv)}},
		),
		vukem.WithInertia(inertia.Config{
			Version: version,
			SSR: &inertia.SSRConfig{
				Bundle: "bootstrap/ssr/ssr.js",
				OnFailure: func(ctx context.Context, err inertia.SSRError) {
					ssrFailures.Observe(ctx, err)
					metrics.RecordSSRFallback(ctx, err.URL, metrics.PublicPage(err.URL))
				},
				// Private pages skip SSR: no SEO value, user data stays out of the Node process (SDD 09 §6.1).
				Except: []string{"/admin", "/profile", "/me", "/billing", "/suggest", "/suggestions"},
			},
		}),
	}

	base = append(base, vukem.WithRoutes(func(r *router.Router, app *vukem.App) error {
		if err := addReadyChecks(app, heartbeat); err != nil {
			return err
		}
		// workaround(vukem): VF-022. vukem's Logger takes app.Log when routes.Web installs the
		// default stack, which runs after this callback; wrapping app.Log here keeps
		// LOG_FORMAT, dedup and trace correlation from vukem's own handler chain.
		logKey, err := app.DeriveKey("log", 32)
		if err != nil {
			return err
		}
		app.Log = slog.New(logging.NewHandler(app.Log.Handler(), logKey))
		r.UseServer(middleware.RequestLog)
		r.Use(middleware.RequestRoute)
		edge, err := middleware.APIEdge(config.APIHost(app.Source.Get),
			cmp.Or(app.Source.Get("APP_PROBLEM_BASE_URL"), "https://hexascore.app/problems/"))
		if err != nil {
			return err
		}
		// Production only: dev and tests are reached by localhost, 127.0.0.1 and test hosts.
		if app.Config.IsProduction() {
			canon, err := middleware.CanonicalHost(cmp.Or(app.Source.Get("APP_URL"), "https://hexascore.app"),
				config.APIHost(app.Source.Get))
			if err != nil {
				return err
			}
			r.UseServer(canon)
		}
		r.UseServer(edge)
		r.UseServer(middleware.Version(middleware.ResolveVersion(app.Source.Get)))
		routes.InstallAppMinVersion(r, app)
		return nil
	}))

	for _, fn := range routes.All() {
		base = append(base, vukem.WithRoutes(fn))
	}

	app, err := vukem.New(append(base, opts...)...)
	if err != nil {
		return nil, err
	}

	if err := metrics.Install(app, heartbeat.Age); err != nil {
		return nil, err
	}

	built = app
	app.AddCommand(Commands()...)

	return app, nil
}

// addReadyChecks registers the four readiness checks of SDD 12 §7.1, each
// bounded by ReadyCheckTimeout. db replaces the framework's 3 s default.
func addReadyChecks(app *vukem.App, heartbeat *scheduler.Heartbeat) error {
	db, err := app.DB()
	if err != nil {
		return err
	}

	app.AddHealthCheckTimeout("db", ReadyCheckTimeout, db.Ping)
	app.AddHealthCheckTimeout("ssr", ReadyCheckTimeout, app.Inertia.SSRHealthy)

	// shortcut: probes the local exports disk until the bucket lands (P2.A.2.1), then check that disk instead.
	disk, err := storage.NewLocal(library.ExportsDir(app.Source.Get), "")
	if err != nil {
		return err
	}

	app.OnShutdown(func(context.Context) error { return disk.Close() })
	app.AddHealthCheckTimeout("storage", ReadyCheckTimeout, func(ctx context.Context) error {
		return storage.Check(ctx, disk)
	})

	// The heartbeat is in memory: a serve that does not run the schedule would always fail it.
	inServe, err := strconv.ParseBool(cmp.Or(app.Source.Get("SCHEDULE_IN_SERVE"), "true"))
	if err != nil {
		return err
	}

	if inServe {
		app.AddHealthCheckTimeout("scheduler", ReadyCheckTimeout, heartbeat.Check)
	}

	return nil
}
