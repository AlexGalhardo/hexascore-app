//go:build integration

package discover_test

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

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
			Source: "wikidata", ExternalID: t.ID, Payload: json.RawMessage(fmt.Sprintf(`{"type":%q}`, t.Type)),
		})
	}

	return out, nil
}

func gz(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()

	zw := gzip.NewWriter(w)
	_, _ = zw.Write([]byte(body))
	_ = zw.Close()
}

// The step walks the movie export, then the series one, maps ids to Wikidata items, stores their records once,
// remembers its position, and does nothing more the same day.
func TestDiscoverWalksTheExportsAndSavesTheirRecords(t *testing.T) {
	ctx := context.Background()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/movie_ids_11_03_2026"):
			gz(t, w, `{"adult":false,"id":603,"popularity":80}
{"adult":false,"id":12,"popularity":40}
{"adult":false,"id":99999,"popularity":40}
`)
		case strings.Contains(r.URL.Path, "/tv_series_ids_11_03_2026"):
			gz(t, w, `{"adult":false,"id":1399,"popularity":90}`)
		case r.URL.Path == "/sparql":
			q := r.URL.Query().Get("query")
			switch {
			case strings.Contains(q, "P4947"):
				_, _ = w.Write([]byte(`{"results":{"bindings":[
{"item":{"value":"http://www.wikidata.org/entity/Q83495"}},
{"item":{"value":"http://www.wikidata.org/entity/Q4"}}]}}`)) // 99999 has no Wikidata item
			default:
				_, _ = w.Write(
					[]byte(`{"results":{"bindings":[{"item":{"value":"http://www.wikidata.org/entity/Q23572"}}]}}`),
				)
			}
		default:
			http.NotFound(w, r)
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
VALUES (1, 'discover.tmdb', 'tmdb', 'full', 'running', 't')`,
	} {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatalf("%v\n%s", err, q)
		}
	}

	// Q4 is already stored: it must not be fetched again.
	if err := db.Tx(ctx, func(tx *database.Tx) error {
		_, err := store.Upsert(
			ctx,
			tx,
			1,
			ingest.Record{Source: "wikidata", ExternalID: "Q4", Payload: json.RawMessage(`{}`)},
		)

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
		Client: client, Fetcher: &fakeFetcher{types: types}, Batch: 2,
		Now:        func() time.Time { return time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC) },
		ExportBase: srv.URL + "/p/exports", SparqlURL: srv.URL + "/sparql",
	}

	step := func() run.Result {
		var res run.Result

		err := db.Tx(ctx, func(tx *database.Tx) (err error) {
			res, err = cfg.Step(ctx, tx, run.Run{ID: 1, Job: discover.Job})

			return err
		})
		if err != nil {
			t.Fatal(err)
		}

		return res
	}

	done := false

	for range 10 {
		if step().Done {
			done = true

			break
		}
	}

	if !done {
		t.Fatal("the step never finished")
	}

	if types["Q83495"] != "movie" || types["Q23572"] != "series" {
		t.Errorf("fetched types = %v, want the movie and the series", types)
	}

	if _, again := types["Q4"]; again {
		t.Error("Q4 already had a record and must not be fetched again")
	}

	var n int
	if err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM source_records WHERE source = 'wikidata' AND external_id IN ('Q83495', 'Q23572')`).
		Scan(&n); err != nil || n != 2 {
		t.Errorf("stored records = %d (%v), want 2", n, err)
	}

	if res := step(); !res.Done || res.Seen != 0 {
		t.Errorf("a second pass the same day = %+v, want done with nothing seen", res)
	}
}
