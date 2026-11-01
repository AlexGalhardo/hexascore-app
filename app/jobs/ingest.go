package jobs

import (
	"context"
	"fmt"
	"sync"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/s3"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/backfill"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/covers"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/dq"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/fetch"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/gog"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/igdb"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/openlibrary"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/rollback"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/steam"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/tmdb"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/trailers"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikidata"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikipedia"
)

// IngestSources replaces the outbound client and the source endpoints; tests only, passed to AllWith or
// bootstrap.NewWithIngest. The zero value is production.
type IngestSources struct {
	Client      *ingest.Client
	WikidataAPI string
	WikidataRDF string
	// WikipediaAPI is a URL prefix; the edition code is appended.
	WikipediaAPI string
	TMDBAPI      string
	// OpenLibraryAPI replaces https://openlibrary.org.
	OpenLibraryAPI string
	// TwitchTokenURL and IGDBAPI replace the IGDB endpoints.
	TwitchTokenURL string
	IGDBAPI        string
	// SteamAPI and GOGAPI replace the storefront endpoints.
	SteamAPI string
	GOGAPI   string
	// Covers and CoversEncoder replace the bucket and cwebp.
	Covers        storage.Disk
	CoversEncoder covers.Encoder
	// Pressure holds ingestion chunks while the site is busy; bootstrap fills it and feeds it from the Inflight middleware.
	Pressure *ingest.Pressure
}

const defaultContactURL = "https://hexascore.app"

// handleIngest registers the ingestion run job (queue ingest). The Runner is built on first use because it
// needs the queue, which is built from this very list.
func handleIngest(app *vukem.App, src IngestSources) queue.Registration {
	return run.LazyRegistration(
		sync.OnceValues(func() (*run.Runner, error) { return newIngestRunner(app, src) }),
	)
}

func newIngestRunner(app *vukem.App, src IngestSources) (*run.Runner, error) {
	db, err := app.DB()
	if err != nil {
		return nil, err
	}

	q, err := app.Queue()
	if err != nil {
		return nil, err
	}

	ff, err := app.Flags()
	if err != nil {
		return nil, err
	}

	client := src.Client
	if client == nil {
		contact := app.Source.Get("INGEST_CONTACT_URL")
		if contact == "" {
			contact = defaultContactURL
		}

		client = httpx.New(httpx.Config{ContactURL: contact, Hosts: httpx.DefaultHosts()})
	}

	wd := wikidata.New(client)
	wd.SparqlURL = src.WikidataRDF
	wd.APIURL = src.WikidataAPI

	tc := tmdb.New(app.Source.Get("THEMOVIEDATABASE_API_READ_ACCESS_TOKEN"))
	if src.TMDBAPI != "" {
		tc.APIURL = src.TMDBAPI
	}

	ig := igdb.New(app.Source.Get("IGDB_TWITCH_CLIENT_ID"), app.Source.Get("IGDB_TWITCH_CLIENT_SECRET"))
	if src.TwitchTokenURL != "" {
		ig.TokenURL = src.TwitchTokenURL
	}

	if src.IGDBAPI != "" {
		ig.APIURL = src.IGDBAPI
	}

	disk, err := CoversDisk(app)
	if err != nil {
		// A broken bucket config must not stop the other sources: the covers step then fails with ErrNoDisk.
		app.Log.Error("ingest runner: covers disk", "error", err)

		disk = nil
	}

	if src.Covers != nil {
		disk = src.Covers
	}

	var stat dq.Stat
	if disk != nil {
		stat = disk.Exists
	}

	coversJob := &covers.Job{
		Disk: disk, HTTP: client.Image, Encoder: &covers.Cwebp{}, Now: app.Now,
		ItemsChanged: func(ctx context.Context, ex database.Executor, ids []string) error {
			return q.DispatchTx(ctx, ex, catalog.ItemsChanged{IDs: ids})
		},
	}
	if src.CoversEncoder != nil {
		coversJob.Encoder = src.CoversEncoder
	}

	steps := map[string]run.Step{
		covers.JobName:       coversJob.Step,
		"fetch.igdb":         fetch.Step(ig, client, app.Now),
		"fetch.steam":        fetch.Step(&steam.Connector{APIURL: src.SteamAPI}, client, app.Now),
		"fetch.gog":          fetch.Step(&gog.Connector{APIURL: src.GOGAPI}, client, app.Now),
		igdb.JobName:         igdb.RefreshStep(ig, client, app.Now),
		"fetch.tmdb":         fetch.Step(tc, client, app.Now),
		tmdb.JobName:         tmdb.ExpireStep(tc, client, app.Now),
		"fetch.wikidata":     fetch.Step(wd, client, app.Now),
		"fetch.entities":     fetch.Step(wd.Entities(), client, app.Now),
		"fetch.youtube":      fetch.Step(&trailers.Connector{}, client, app.Now),
		backfill.EditionsJob: backfill.Editions{}.Step,
		// workaround: Serialise (DQ-I-12) stays nil until the API probe exists; Stat is nil without a bucket.
		dq.Job:       dq.Asserts{Stat: stat}.Step,
		rollback.Job: rollback.Step,
		dq.FrontendJob: dq.FrontendFields{
			Suggest: community.IngestGate{DB: db}.OpenMissingField,
			Now:     app.Now,
		}.Step,
		"fetch.openlibrary": fetch.Step(&openlibrary.Connector{BaseURL: src.OpenLibraryAPI}, client, app.Now),
		"normalize": normalize.Normalizer{
			Community: community.IngestGate{DB: db},
			Now:       app.Now,
			ItemsChanged: func(ctx context.Context, ex database.Executor, ids []string) error {
				return q.DispatchTx(ctx, ex, catalog.ItemsChanged{IDs: ids})
			},
			ScoreChanged: func(ctx context.Context, ex database.Executor, id string) error {
				return q.DispatchTx(ctx, ex, scoring.RecomputeItemScore{ItemID: id})
			},
		}.Step,
	}

	for _, lang := range wikipedia.Langs {
		c, err := wikipedia.New(lang)
		if err != nil {
			return nil, fmt.Errorf("ingest runner: %w", err)
		}

		if src.WikipediaAPI != "" {
			c.APIURL = src.WikipediaAPI + lang
		}

		steps["fetch."+c.Source()] = fetch.Step(c, client, app.Now)
	}

	runner := &run.Runner{
		DB:    db,
		Queue: q,
		Flags: ff,
		Steps: steps,
		Now:   app.Now,
		After: func(ctx context.Context, id int64) error { return ingest.Advance(ctx, db, q, id) },
	}
	if src.Pressure != nil {
		runner.Gate = src.Pressure
	}

	return runner, nil
}

// CoversDisk is the S3 bucket covers live in, or nil when S3_BUCKET is unset (development: the embedded demo
// covers are served instead and the covers job fails with covers.ErrNoDisk).
func CoversDisk(app *vukem.App) (storage.Disk, error) {
	if app.Source.Get("S3_BUCKET") == "" {
		return nil, nil //nolint:nilnil // no bucket is a supported setup
	}

	cfg, err := config.Load[s3.Config](app.Source)
	if err != nil {
		return nil, fmt.Errorf("covers disk: %w", err)
	}

	return s3.New(cfg)
}
