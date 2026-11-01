// Package bootstrap wires the application together. There is no hidden
// registration: everything the app uses is listed here.
package bootstrap

import (
	"cmp"
	"context"
	"html/template"
	"log/slog"
	"os"
	"strconv"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/flags"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/openapi"
	"github.com/AlexGalhardo/vukem/queue"
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
	appmail "github.com/AlexGalhardo/hexascore-app/app/mail"
	"github.com/AlexGalhardo/hexascore-app/app/scheduler"
	"github.com/AlexGalhardo/hexascore-app/app/social"
	"github.com/AlexGalhardo/hexascore-app/app/ssr"
	"github.com/AlexGalhardo/hexascore-app/config"
	"github.com/AlexGalhardo/hexascore-app/database/migrations"
	"github.com/AlexGalhardo/hexascore-app/database/seeders"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
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
// read it.
func NewWithSSRFailures(ssrFailures *ssr.Failures, opts ...vukem.Option) (*vukem.App, error) {
	return build(ssrFailures, jobs.IngestSources{}, opts...)
}

// NewWithIngest is New with the ingestion client and source endpoints replaced; tests only.
func NewWithIngest(src jobs.IngestSources, opts ...vukem.Option) (*vukem.App, error) {
	return build(&ssr.Failures{}, src, opts...)
}

func build(ssrFailures *ssr.Failures, ingestSrc jobs.IngestSources, opts ...vukem.Option) (*vukem.App, error) {
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

	// The Inflight middleware feeds this gate and the ingestion runner reads it (SDD 06 §7.5).
	pressure := ingest.NewPressure(time.Now)
	ingestSrc.Pressure = pressure

	// Set by the routes callback below, read at render time by the turnstileScript view function.
	var turnstile routes.Turnstile

	base := []vukem.Option{
		vukem.WithEnv(nil, config.WithQueueLimits(os.LookupEnv)),
		vukem.WithProviders(Providers()...),
		vukem.WithDatabaseDrivers(postgres.Driver()),
		vukem.WithMigrations(migrations.All()),
		vukem.WithSeeders(seeders.All()...),
		vukem.WithAppSeeders(devSeeders()...),
		vukem.WithListeners(listeners.All()...),
		vukem.WithJobsFunc(func(app *vukem.App) []queue.Registration { return jobs.AllWith(app, ingestSrc) }),
		vukem.WithSchedule(func(s *schedule.Schedule, app *vukem.App) {
			routes.Schedule(s, app)
			s.Func("heartbeat", heartbeat.Beat).EveryMinute()
		}),
		vukem.WithFlags(
			append(
				[]flags.AnyFlag{community.SubmissionsFlag, community.TrustedModeratorsFlag, social.Flag},
				run.All()...)...),
		vukem.WithViews(resources.Views()),
		vukem.WithViewFuncs(template.FuncMap{
			"turnstileScript": func(nonce string) template.HTML { return turnstile.Script(nonce) },
		}),
		vukem.WithVite(resources.Public()),
		vukem.WithTranslations(resources.Lang()),
		vukem.WithOpenAPI(apiInfo(config.APIHost(os.Getenv))),
		vukem.WithInertia(inertia.Config{
			Version: version,
			SSR: &inertia.SSRConfig{
				Bundle: "bootstrap/ssr/ssr.js",
				OnFailure: func(ctx context.Context, err inertia.SSRError) {
					ssrFailures.Observe(ctx, err)
				},
				// Private pages skip SSR: no SEO value, user data stays out of the Node process (SDD 09 §6.1).
				Except: ssr.PrivatePrefixes,
			},
		}),
	}

	base = append(base, vukem.WithRoutes(func(r *router.Router, app *vukem.App) error {
		ts, err := routes.LoadTurnstile(app)
		if err != nil {
			return err
		}
		turnstile = ts
		appmail.Configure(app.Source.Get("MAIL_FROM"), app.Source.Get("MAIL_SUPPORT_ADDRESS"))
		// MAIL_REPLY_TO wins; otherwise automatic mail answers to MAIL_SUPPORT_ADDRESS (Q-A17). A mailer error
		// is not swallowed: Send returns the same error on first use.
		if m, err := app.Mailer(); err == nil && app.Source.Get("MAIL_REPLY_TO") == "" {
			m.SetReplyTo(appmail.SupportAddress())
		}
		if err := addReadyChecks(app, heartbeat); err != nil {
			return err
		}
		// vukem's Logger takes app.Log when routes.Web installs the default stack, which runs after this callback;
		// wrapping app.Log here keeps LOG_FORMAT, dedup and trace correlation from vukem's own handler chain. The
		// router adds the route pattern to the request line and NotBanned the user (web.AddLogAttrs).
		logKey, err := app.DeriveKey("log", 32)
		if err != nil {
			return err
		}
		app.Log = slog.New(logging.NewHandler(app.Log.Handler(), logKey))
		r.UseServer(middleware.Inflight(pressure))
		edge, err := middleware.APIEdge(
			config.APIHost(app.Source.Get),
			cmp.Or(app.Source.Get("APP_URL"), "https://hexascore.app"),
			cmp.Or(app.Source.Get("APP_PROBLEM_BASE_URL"), "https://hexascore.app/problems/"),
		)
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
		r.Services().ProblemDecorator = middleware.ProblemDecorator(config.APIHost(app.Source.Get),
			cmp.Or(app.Source.Get("APP_PROBLEM_BASE_URL"), "https://hexascore.app/problems/"))
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

// apiInfo documents the public API the way the api host serves it: server https://<apiHost>/v1 and paths without
// /api/v1 (APIEdge maps /v1 to the /api/v1 routes; SDD 05 section 6.1). The app document (/app/v1) is first-party,
// has no such host form and keeps full paths through its Docs entry.
func apiInfo(apiHost string) openapi.Info {
	return openapi.Info{
		Title:       "HexaScore API",
		Servers:     []string{"https://" + apiHost + "/v1"},
		StripPrefix: "/api/v1",
		Docs: map[string]openapi.DocInfo{
			"/app/v1": {Servers: []string{"https://" + apiHost}},
		},
	}
}
