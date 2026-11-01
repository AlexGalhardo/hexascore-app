package jobs

import (
	"context"
	"fmt"
	"sync"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/fetch"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikidata"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikipedia"
)

// IngestSources replaces the outbound client and the source endpoints; tests only, set before the app is built.
type IngestSources struct {
	Client      *ingest.Client
	WikidataAPI string
	WikidataRDF string
	// WikipediaAPI is a URL prefix; the edition code is appended.
	WikipediaAPI string
}

// IngestOverride is the test seam of handleIngest; production leaves it zero.
var IngestOverride IngestSources

const defaultContactURL = "https://hexascore.app"

// handleIngest registers the ingestion run job (queue ingest). The Runner is built on first use because it
// needs the queue, which is built from this very list.
func handleIngest(app *vukem.App) queue.Registration {
	return run.LazyRegistration(
		sync.OnceValues(func() (*run.Runner, error) { return newIngestRunner(app, IngestOverride) }),
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

	steps := map[string]run.Step{
		"fetch.wikidata": fetch.Step(wd, client, app.Now),
		"fetch.entities": fetch.Step(wd.Entities(), client, app.Now),
		// The Open Library licence is unverified (A-05): the stage exists so the chain is complete, and the kill switch keeps it skipped.
		"fetch.openlibrary": fetch.Unavailable("openlibrary", "licence unverified, A-05"),
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

	return &run.Runner{
		DB:    db,
		Queue: q,
		Flags: ff,
		Steps: steps,
		Now:   app.Now,
		After: func(ctx context.Context, id int64) error { return ingest.Advance(ctx, db, q, id) },
	}, nil
}
