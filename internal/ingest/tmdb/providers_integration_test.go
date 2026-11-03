//go:build integration

package tmdb_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/tmdb"
)

type providerRig struct {
	t      *testing.T
	db     *database.DB
	ctx    context.Context
	step   run.Step
	server *httptest.Server
	mu     sync.Mutex
	asked  []string
	clock  time.Time
}

func newProviderRig(t *testing.T, status map[string]int) *providerRig {
	t.Helper()

	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile("testdata/providers.json")
	if err != nil {
		t.Fatal(err)
	}

	r := &providerRig{t: t, db: db, ctx: context.Background(), clock: time.Date(2026, 11, 3, 1, 0, 0, 0, time.UTC)}

	r.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.mu.Lock()
		r.asked = append(r.asked, req.URL.Path)
		r.mu.Unlock()

		if s, ok := status[req.URL.Path]; ok && s != http.StatusOK {
			w.WriteHeader(s)

			return
		}

		if req.URL.Path == "/3/movie/603/watch/providers" {
			_, _ = w.Write(body)

			return
		}

		if req.URL.Path == "/3/tv/10/watch/providers" {
			_, _ = w.Write([]byte(`{"id":10,"results":{}}`))

			return
		}

		http.NotFound(w, req)
	}))
	t.Cleanup(r.server.Close)

	host := strings.TrimPrefix(r.server.URL, "http://")
	client := httpx.New(httpx.Config{
		ContactURL: "https://hexascore.app/bot",
		Hosts:      map[string]httpx.HostLimit{host: {PerSecond: 1000, InFlight: 4}},
		Allow:      []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")},
	})

	c := tmdb.New("token")
	c.APIURL = r.server.URL + "/3"
	r.step = tmdb.ProvidersStep(c, client, func() time.Time { return r.clock })

	return r
}

func (r *providerRig) exec(q string, args ...any) {
	r.t.Helper()

	if _, err := r.db.ExecContext(r.ctx, q, args...); err != nil {
		r.t.Fatalf("%v\n%s", err, q)
	}
}

var qid int

func (r *providerRig) item(typ, visibility, tmdbID string) string {
	r.t.Helper()

	qid++
	id := uuid.NewV7().String()
	r.exec(
		`INSERT INTO items (id, type, slug, wikidata_qid, visibility, external_ids) VALUES ($1, $2, $3, $4, $5, $6::jsonb)`,
		id,
		typ,
		fmt.Sprintf("p-%d", qid),
		fmt.Sprintf("Q%d", 9000+qid),
		visibility,
		fmt.Sprintf(`{"tmdb":%q}`, tmdbID),
	)

	return id
}

// runToEnd runs the step in pages until it says done and returns the summed result.
func (r *providerRig) runToEnd() (failed, changed int) {
	r.t.Helper()

	r.exec(
		`INSERT INTO job_runs (id, job, source, mode, status, params, triggered_by) OVERRIDING SYSTEM VALUE VALUES (1, 'offers.tmdb', 'tmdb', 'full', 'running', '{}', 'test')
ON CONFLICT DO NOTHING`,
	)

	var cur []byte

	for range 50 {
		var res run.Result

		err := r.db.Tx(r.ctx, func(tx *database.Tx) (err error) {
			res, err = r.step(r.ctx, tx, run.Run{ID: 1, Job: "offers.tmdb", Source: "tmdb", Cursor: cur})

			return err
		})
		if err != nil {
			r.t.Fatal(err)
		}

		failed += res.Failed
		changed += res.Changed
		cur = res.Cursor

		if res.Done {
			return failed, changed
		}
	}

	r.t.Fatal("step did not finish")

	return 0, 0
}

func (r *providerRig) count(q string, args ...any) (n int) {
	r.t.Helper()

	if err := r.db.QueryRowContext(r.ctx, q, args...).Scan(&n); err != nil {
		r.t.Fatal(err)
	}

	return n
}

func TestProvidersStepFetchesParsesAndRemembers(t *testing.T) {
	r := newProviderRig(t, map[string]int{"/3/movie/700/watch/providers": http.StatusBadRequest})

	film := r.item("movie", "public", "603")
	series := r.item("series", "public", "10")
	broken := r.item("movie", "public", "700")
	r.item("movie", "hidden", "603") // not public: never asked
	r.item("book", "public", "603")  // not a film or series
	r.item("movie", "public", "6x3") // not a TMDB id

	failed, changed := r.runToEnd()
	if failed != 1 || changed != 7 {
		t.Fatalf("failed=%d changed=%d, want 1 failure (the 400) and 7 offers", failed, changed)
	}

	if n := r.count(
		`SELECT count(*) FROM offers WHERE item_id = $1 AND source = 'tmdb_providers' AND active`,
		film,
	); n != 7 {
		t.Fatalf("offers of the film = %d, want 7", n)
	}

	if n := r.count(`SELECT count(*) FROM offer_source_checks WHERE item_id IN ($1, $2)`, film, series); n != 2 {
		t.Fatalf("checked items = %d, want the film and the series (a title nobody carries is still checked)", n)
	}

	if n := r.count(`SELECT count(*) FROM offer_source_checks WHERE item_id = $1`, broken); n != 0 {
		t.Fatal("an item whose request failed must be asked about again next run")
	}

	r.mu.Lock()
	got := strings.Join(r.asked, " ")
	r.asked = nil
	r.mu.Unlock()

	for _, bad := range []string{"6x3", "/book/"} {
		if strings.Contains(got, bad) {
			t.Fatalf("asked about something it must not: %s", got)
		}
	}

	// a week later only the item that failed is asked about again
	r.clock = r.clock.AddDate(0, 0, 7)
	r.runToEnd()

	r.mu.Lock()
	again := append([]string(nil), r.asked...)
	r.mu.Unlock()

	if len(again) != 1 || again[0] != "/3/movie/700/watch/providers" {
		t.Fatalf("second run asked %v, want only the failed item", again)
	}

	// 31 days later everything is due again
	r.mu.Lock()
	r.asked = nil
	r.mu.Unlock()

	r.clock = r.clock.AddDate(0, 0, 31)
	r.runToEnd()

	r.mu.Lock()
	n := len(r.asked)
	r.mu.Unlock()

	if n != 3 {
		t.Fatalf("a month on, %d requests, want the three public films and series", n)
	}
}

func TestProvidersStepStopsOnARejectedToken(t *testing.T) {
	r := newProviderRig(t, map[string]int{"/3/movie/603/watch/providers": http.StatusUnauthorized})
	r.item("movie", "public", "603")
	r.exec(
		`INSERT INTO job_runs (id, job, source, mode, status, params, triggered_by) OVERRIDING SYSTEM VALUE VALUES (1, 'offers.tmdb', 'tmdb', 'full', 'running', '{}', 'test')`,
	)

	err := r.db.Tx(r.ctx, func(tx *database.Tx) error {
		_, err := r.step(r.ctx, tx, run.Run{ID: 1, Job: "offers.tmdb", Source: "tmdb"})

		return err
	})
	if err == nil {
		t.Fatal("a rejected token must stop the run, not log a failure per item")
	}
}
