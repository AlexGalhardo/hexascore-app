//go:build integration

package ingest_test

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"strings"
	"sync"
	"testing"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
)

// fakeSources stands in for Wikidata (SPARQL and wbgetentities) and Wikipedia on one loopback server. The tests
// never reach a live API: the client allows only this server's address.
type fakeSources struct {
	srv *httptest.Server

	mu      sync.Mutex
	queries []string
	// topByClass maps a class QID found in the SPARQL query to the one item that type returns.
	topByClass map[string]string
	entities   map[string]json.RawMessage
	// titles maps "<lang>|<title>" to the QID the fixture entity links there.
	titles map[string]string
}

func loadFake(t *testing.T) *fakeSources {
	t.Helper()

	f := &fakeSources{
		topByClass: map[string]string{
			"wd:Q11424":   "Q83495",
			"wd:Q5398426": "Q1079",
			"wd:Q7889":    "Q4267401",
			"wd:Q7725634": "Q208460",
		},
		entities: map[string]json.RawMessage{},
		titles:   map[string]string{},
	}

	for _, name := range []string{"wikidata_labels.json", "wikidata_entities.json"} {
		raw, err := os.ReadFile("../../internal/ingest/testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}

		var doc struct {
			Entities map[string]json.RawMessage `json:"entities"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}

		for id, e := range doc.Entities {
			f.entities[id] = e

			var sl struct {
				Sitelinks map[string]struct {
					Title string `json:"title"`
				} `json:"sitelinks"`
			}

			if json.Unmarshal(e, &sl) != nil {
				continue
			}

			for key, s := range sl.Sitelinks {
				if len(key) == 6 && strings.HasSuffix(key, "wiki") {
					f.titles[key[:2]+"|"+s.Title] = id
				}
			}
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/sparql", f.sparql)
	mux.HandleFunc("/wikidata", f.wbgetentities)
	mux.HandleFunc("/wikipedia/", f.extracts)
	mux.HandleFunc("/3/", f.tmdb)

	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)

	return f
}

func (f *fakeSources) sparql(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("query")

	f.mu.Lock()
	f.queries = append(f.queries, q)
	f.mu.Unlock()

	bindings := []map[string]any{}

	for class, qid := range f.topByClass {
		if strings.Contains(q, class) {
			bindings = append(
				bindings,
				map[string]any{"item": map[string]string{"value": "http://www.wikidata.org/entity/" + qid}},
			)
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"results": map[string]any{"bindings": bindings}})
}

func (f *fakeSources) wbgetentities(w http.ResponseWriter, r *http.Request) {
	out := map[string]json.RawMessage{}

	for id := range strings.SplitSeq(r.URL.Query().Get("ids"), "|") {
		e, ok := f.entities[id]
		if !ok {
			e = json.RawMessage(`{"id":"` + id + `","missing":""}`)
		}

		out[id] = e
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"entities": out})
}

// extracts answers /wikipedia/<lang> with one page per requested title that some fixture entity links to.
func (f *fakeSources) extracts(w http.ResponseWriter, r *http.Request) {
	lang := strings.TrimPrefix(r.URL.Path, "/wikipedia/")
	pages := []map[string]any{}

	for title := range strings.SplitSeq(r.URL.Query().Get("titles"), "|") {
		if qid, ok := f.titles[lang+"|"+title]; ok {
			pages = append(pages, map[string]any{
				"title":     title,
				"extract":   "Synopsis of " + title + " (" + lang + ")",
				"fullurl":   "https://" + lang + ".wikipedia.org/wiki/" + strings.ReplaceAll(title, " ", "_"),
				"pageprops": map[string]string{"wikibase_item": qid},
			})
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"query": map[string]any{"pages": pages}})
}

// newIngestApp builds the real app with the fake sources and the database queue, so queue:work runs the
// handlers jobs.All registered.
func newIngestApp(t *testing.T, f *fakeSources, flagsOn ...string) *vukemtest.App {
	t.Helper()

	host := strings.TrimPrefix(f.srv.URL, "http://")
	client := httpx.New(httpx.Config{
		ContactURL: "https://hexascore.app/bot",
		Hosts:      map[string]httpx.HostLimit{host: {PerSecond: 1000, InFlight: 4}},
		Allow:      []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")},
	})

	src := jobs.IngestSources{
		Client:       client,
		WikidataRDF:  f.srv.URL + "/sparql",
		WikidataAPI:  f.srv.URL + "/wikidata",
		WikipediaAPI: f.srv.URL + "/wikipedia/",
		TMDBAPI:      f.srv.URL + "/3",
	}

	env := map[string]string{
		"DATABASE_URL":                           pgtest.Database(t, ""),
		"QUEUE_CONNECTION":                       "database",
		"QUEUE_NAMES":                            "ingest,default",
		"THEMOVIEDATABASE_API_READ_ACCESS_TOKEN": "test-token",
	}
	for _, name := range flagsOn {
		env["FLAG_INGEST_"+name] = "true"
	}

	return vukemtest.New(t, func(opts ...vukem.Option) (*vukem.App, error) {
		return bootstrap.NewWithIngest(src, opts...)
	}, vukemtest.Options{Env: env})
}

func runCommand(t *testing.T, app *vukemtest.App, name string, args ...string) {
	t.Helper()

	for _, c := range app.Commands() {
		if c.Name() != name {
			continue
		}

		fs := flag.NewFlagSet(name, flag.ContinueOnError)
		c.Flags(fs)

		if err := fs.Parse(args); err != nil {
			t.Fatal(err)
		}

		if err := c.Run(context.Background(), app.App, fs.Args()); err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		return
	}

	t.Fatalf("command %q is not registered", name)
}

func startChain(t *testing.T, app *vukemtest.App, c ingest.Chain) {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	q, err := app.Queue()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := c.Start(context.Background(), db, q, "test", nil); err != nil {
		t.Fatal(err)
	}
}

func jobStatus(t *testing.T, app *vukemtest.App) map[string]string {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	rows, err := db.QueryContext(context.Background(), `SELECT job, status FROM job_runs`)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = rows.Close() }()

	got := map[string]string{}

	for rows.Next() {
		var job, status string
		if err := rows.Scan(&job, &status); err != nil {
			t.Fatal(err)
		}

		got[job] = status
	}

	return got
}

func appCount(t *testing.T, app *vukemtest.App, q string, args ...any) int {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	return count(t, db, q, args...)
}

func TestBootstrapThenQueueWorkNormalises(t *testing.T) {
	f := loadFake(t)
	app := newIngestApp(t, f, "ENABLED", "WIKIDATA")

	runCommand(t, app, "ingest:bootstrap", "--limit", "20")
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	want := map[string]string{"fetch.wikidata": "succeeded", "fetch.entities": "succeeded", "normalize": "succeeded"}
	got := jobStatus(t, app)

	for job, status := range want {
		if got[job] != status {
			t.Errorf("job %s = %q, want %q (all: %v)", job, got[job], status, got)
		}
	}

	if n := appCount(t, app, `SELECT count(*) FROM items`); n != 4 {
		t.Errorf("items = %d, want one per type (4)", n)
	}

	if n := appCount(t, app, `SELECT count(DISTINCT type) FROM items`); n != 4 {
		t.Errorf("item types = %d, want 4", n)
	}

	if n := appCount(t, app, `SELECT count(*) FROM source_records WHERE source = 'wikidata:entity'`); n == 0 {
		t.Error("no referenced entity was stored")
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.queries) == 0 || !strings.Contains(f.queries[0], "LIMIT 20") {
		t.Errorf("--limit 20 did not reach the SPARQL query: %v", f.queries)
	}
}

func TestWikipediaChainRunsEveryEdition(t *testing.T) {
	f := loadFake(t)
	app := newIngestApp(t, f, "ENABLED", "WIKIDATA", "WIKIPEDIA")

	runCommand(t, app, "ingest:bootstrap")
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	startChain(t, app, ingest.WikipediaWeekly)
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	got := jobStatus(t, app)
	for _, job := range []string{"fetch.wikipedia:en", "fetch.wikipedia:pt", "fetch.wikipedia:es"} {
		if got[job] != "succeeded" {
			t.Errorf("job %s = %q (all: %v)", job, got[job], got)
		}
	}

	if n := appCount(t, app, `SELECT count(*) FROM source_records WHERE source = 'wikipedia:en'`); n != 4 {
		t.Errorf("wikipedia:en records = %d, want 4", n)
	}
}

func TestOpenLibraryStageAnswersSkippedWhileUnlicensed(t *testing.T) {
	f := loadFake(t)
	app := newIngestApp(t, f, "ENABLED", "WIKIDATA")

	startChain(t, app, ingest.OpenLibraryWeekly)
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	got := jobStatus(t, app)
	if got["fetch.openlibrary"] != "skipped" {
		t.Errorf("fetch.openlibrary = %q, want skipped", got["fetch.openlibrary"])
	}

	if _, ok := got["normalize"]; ok {
		t.Error("a skipped stage queued the normaliser")
	}
}

func TestWeeklyWikidataRefreshesStoredEntities(t *testing.T) {
	f := loadFake(t)
	app := newIngestApp(t, f, "ENABLED", "WIKIDATA")

	runCommand(t, app, "ingest:bootstrap")
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	entities := appCount(t, app, `SELECT count(*) FROM source_records WHERE source = 'wikidata:entity'`)
	if entities == 0 {
		t.Fatal("bootstrap stored no entity")
	}

	f.mu.Lock()
	f.entities["Q1"] = json.RawMessage(`{"id":"Q1","missing":""}`)
	f.mu.Unlock()

	startChain(t, app, ingest.WikidataWeekly)
	runCommand(t, app, "queue:work", "-queue", "ingest,default", "-stop-when-empty")

	if n := appCount(t, app, `SELECT count(*) FROM job_runs WHERE status = 'succeeded'`); n != 6 {
		t.Errorf("succeeded runs = %d, want 3 bootstrap + 3 weekly", n)
	}

	if n := appCount(t, app, `SELECT count(*) FROM source_records WHERE source = 'wikidata:entity'`); n != entities {
		t.Errorf("entities = %d after refresh, want %d", n, entities)
	}
}
