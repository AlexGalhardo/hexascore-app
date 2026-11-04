//go:build integration

package discover_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/discover"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

type fakeFetcher struct{ types map[string]string }

func (f *fakeFetcher) Fetch(
	_ context.Context, _ *ingest.Client, batch []ingest.Target,
) ([]ingest.Record, []ingest.TargetError) {
	var out []ingest.Record

	for _, t := range batch {
		f.types[t.ID] = t.Type
		out = append(out, ingest.Record{
			Source: "wikidata", ExternalID: t.ID,
			Payload: json.RawMessage(fmt.Sprintf(`{"type":%q,"sitelinks_count":10}`, t.Type)),
		})
	}

	return out, nil
}

// The step starts below the least popular stored item of a type, selects the next page, stores only the records it
// has not fetched, remembers the position, and ends a type when a page comes back short.
func TestDiscoverWalksDownFromTheStoredItems(t *testing.T) {
	ctx := context.Background()

	var queries []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("query")
		queries = append(queries, q)

		switch {
		case strings.Contains(q, "wd:Q11424"): // movies: one full page of 2, then a short one
			if strings.Contains(q, `STR(?item) > ""`) { // first page, no cursor yet
				_, _ = w.Write([]byte(`{"results":{"bindings":[
{"item":{"value":"http://www.wikidata.org/entity/Q1"},"links":{"value":"9"}},
{"item":{"value":"http://www.wikidata.org/entity/Q2"},"links":{"value":"8"}}]}}`))
			} else {
				_, _ = w.Write([]byte(`{"results":{"bindings":[
{"item":{"value":"http://www.wikidata.org/entity/Q3"},"links":{"value":"7"}}]}}`))
			}
		default: // every other type has nothing below the stored items
			_, _ = w.Write([]byte(`{"results":{"bindings":[]}}`))
		}
	}))
	t.Cleanup(srv.Close)

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
	})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	for _, q := range []string{
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata','Wikidata','CC0-1.0','https://www.wikidata.org')`,
		`INSERT INTO job_runs (id, job, source, mode, status, triggered_by) OVERRIDING SYSTEM VALUE
VALUES (1, 'discover.wikidata', 'wikidata', 'full', 'running', 't')`,
	} {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatalf("%v\n%s", err, q)
		}
	}

	// Q2 is already stored (a movie with 12 sitelinks): the walk starts at 12 and Q2 is not fetched again.
	if err := db.Tx(ctx, func(tx *database.Tx) error {
		_, err := store.Upsert(ctx, tx, 1, ingest.Record{
			Source: "wikidata", ExternalID: "Q2", Payload: json.RawMessage(`{"type":"movie","sitelinks_count":12}`),
		})

		return err
	}); err != nil {
		t.Fatal(err)
	}

	host := strings.TrimPrefix(srv.URL, "http://")
	client := httpx.New(httpx.Config{
		ContactURL: "https://hexascore.app/bot",
		Hosts:      map[string]httpx.HostLimit{host: {PerSecond: 1000, InFlight: 4}},
		Allow:      []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")},
	})

	types := map[string]string{}
	cfg := &discover.Config{
		Client: client, Fetcher: &fakeFetcher{types: types}, Batch: 2, SparqlURL: srv.URL + "/sparql",
	}

	done := false

	for range 20 {
		var res run.Result

		if err := db.Tx(ctx, func(tx *database.Tx) (err error) {
			res, err = cfg.Step(ctx, tx, run.Run{ID: 1, Job: discover.Job})

			return err
		}); err != nil {
			t.Fatal(err)
		}

		if res.Done {
			done = true

			break
		}
	}

	if !done {
		t.Fatal("the walk never finished")
	}

	if types["Q1"] != "movie" || types["Q3"] != "movie" {
		t.Errorf("fetched = %v, want Q1 and Q3 as movies", types)
	}

	if _, again := types["Q2"]; again {
		t.Error("Q2 already had a record and must not be fetched again")
	}

	if !strings.Contains(queries[0], "?links < 12 || (?links = 12") {
		t.Errorf("the walk must start at the least popular stored movie (12 sitelinks):\n%s", queries[0])
	}
}

// A lane walks only its category and keeps its own position row.
func TestDiscoverLaneWalksOneCategory(t *testing.T) {
	ctx := context.Background()

	var queries []string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query().Get("query"))
		_, _ = w.Write(
			[]byte(
				`{"results":{"bindings":[{"item":{"value":"http://www.wikidata.org/entity/Q7"},"links":{"value":"9"}}]}}`,
			),
		)
	}))
	t.Cleanup(srv.Close)

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
	})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	for _, q := range []string{
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata','Wikidata','CC0-1.0','https://www.wikidata.org')`,
		`INSERT INTO job_runs (id, job, source, mode, status, triggered_by) OVERRIDING SYSTEM VALUE
VALUES (1, 'discover.wikidata', 'wikidata:book', 'full', 'running', 't')`,
	} {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatalf("%v\n%s", err, q)
		}
	}

	host := strings.TrimPrefix(srv.URL, "http://")
	client := httpx.New(httpx.Config{
		ContactURL: "https://hexascore.app/bot",
		Hosts:      map[string]httpx.HostLimit{host: {PerSecond: 1000, InFlight: 4}},
		Allow:      []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")},
	})

	types := map[string]string{}
	cfg := &discover.Config{
		Client:    client,
		Fetcher:   &fakeFetcher{types: types},
		Batch:     2,
		SparqlURL: srv.URL + "/sparql",
	}

	for range 5 {
		var res run.Result

		if err := db.Tx(ctx, func(tx *database.Tx) (err error) {
			res, err = cfg.Step(
				ctx,
				tx,
				run.Run{ID: 1, Job: discover.Job, Params: json.RawMessage(`{"types":["book"]}`)},
			)

			return err
		}); err != nil {
			t.Fatal(err)
		}

		if res.Done {
			break
		}
	}

	if types["Q7"] != "book" {
		t.Errorf("fetched = %v, want Q7 as a book", types)
	}

	for _, q := range queries {
		if !strings.Contains(q, "wd:Q7725634") || strings.Contains(q, "wd:Q11424") {
			t.Errorf("a book lane queried another category:\n%s", q)
		}
	}

	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM ingest_host_state WHERE key = 'discover:wikidata:book'`).
		Scan(&n); err != nil || n != 1 {
		t.Errorf("lane state rows = %d (%v), want 1", n, err)
	}
}
