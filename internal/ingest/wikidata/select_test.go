package wikidata_test

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikidata"
)

type stub struct {
	body    []byte
	queries []url.Values
}

func (s *stub) RoundTrip(req *http.Request) (*http.Response, error) {
	s.queries = append(s.queries, req.URL.Query())

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"application/sparql-results+json"}},
		Body:       io.NopCloser(strings.NewReader(string(s.body))),
		Request:    req,
	}, nil
}

func newConnector(t *testing.T) (*wikidata.Connector, *ingest.Client, *stub) {
	t.Helper()

	body, err := os.ReadFile("../testdata/sparql_top_movies.json")
	if err != nil {
		t.Fatal(err)
	}

	s := &stub{body: body}
	c := httpx.New(httpx.Config{
		ContactURL: "https://hexascore.app/bot",
		Hosts:      map[string]httpx.HostLimit{"query.wikidata.org": {PerSecond: 1000, InFlight: 1}},
		Base:       s,
	})

	return wikidata.New(c), c, s
}

func TestSelectQidsParsesSparqlBindings(t *testing.T) {
	w, c, _ := newConnector(t)

	qids, err := w.SelectQids(t.Context(), c, "movie", time.Time{}, 3)
	if err != nil {
		t.Fatal(err)
	}

	if len(qids) != 3 {
		t.Fatalf("want 3 qids, got %v", qids)
	}

	for _, q := range qids {
		if !strings.HasPrefix(q, "Q") || strings.Contains(q, "/") {
			t.Errorf("bad qid %q", q)
		}
	}
}

func TestQueriesTargetTypeClasses(t *testing.T) {
	q := wikidata.TopQuery("game", 10)
	if !strings.Contains(q, "wd:Q7889") || !strings.Contains(q, "LIMIT 10") {
		t.Errorf("top query: %s", q)
	}

	since := time.Date(2026, 3, 4, 5, 6, 7, 0, time.UTC)
	m := wikidata.ModifiedQuery("movie", since, 15)

	for _, want := range []string{"wd:Q11424", "wd:Q24869", "wd:Q202866", "?links >= 15", `"2026-03-04T05:06:07Z"^^xsd:dateTime`} {
		if !strings.Contains(m, want) {
			t.Errorf("modified query lacks %q: %s", want, m)
		}
	}
}

func TestTargetsBootstrapAndIncrementalAndResume(t *testing.T) {
	w, _, s := newConnector(t)

	collect := func(p ingest.RunParams, after string) []ingest.Target {
		var got []ingest.Target

		for tg, err := range w.Targets(t.Context(), nil, p, after) {
			if err != nil {
				t.Fatal(err)
			}

			got = append(got, tg)
		}

		return got
	}

	all := collect(ingest.RunParams{Mode: run.ModeBootstrap, Types: []string{"movie", "game"}, Limit: 2}, "")
	if len(all) != 6 || all[0].Type != "movie" || all[3].Type != "game" {
		t.Fatalf("targets: %+v", all)
	}

	if !strings.Contains(s.queries[0].Get("query"), "LIMIT 2") {
		t.Errorf("bootstrap query: %s", s.queries[0].Get("query"))
	}

	rest := collect(ingest.RunParams{Mode: run.ModeBootstrap, Types: []string{"movie", "game"}, Limit: 2}, all[1].ID)
	if len(rest) != 4 || rest[0] != all[2] {
		t.Fatalf("resume after %s: %+v", all[1].ID, rest)
	}

	s.queries = nil
	collect(
		ingest.RunParams{
			Mode:  run.ModeIncremental,
			Types: []string{"series"},
			Since: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		"",
	)

	if q := s.queries[0].Get("query"); !strings.Contains(q, "dateModified") || !strings.Contains(q, "?links >= 15") {
		t.Errorf("incremental query: %s", q)
	}
}
