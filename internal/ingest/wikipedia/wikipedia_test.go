package wikipedia_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikipedia"
)

var _ ingest.Connector = (*wikipedia.Connector)(nil)

func fixture(t *testing.T) []byte {
	t.Helper()

	raw, err := os.ReadFile("../testdata/wikipedia_extracts_en.json")
	if err != nil {
		t.Fatal(err)
	}

	return raw
}

func TestWikipediaExtracts(t *testing.T) {
	got, err := wikipedia.ParseExtracts(fixture(t), map[string]bool{"Q83495": true, "Q189600": true}, "en")
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 || got[0].Qid != "Q83495" || got[1].Qid != "Q189600" {
		t.Fatalf("payloads = %+v", got)
	}

	if got[0].URL != "https://en.wikipedia.org/wiki/The_Matrix" || got[0].Lang != "en" || got[0].Title != "The Matrix" {
		t.Errorf("payload = %+v", got[0])
	}

	if !strings.HasPrefix(got[0].Extract, "The Matrix") {
		t.Errorf("extract = %q", got[0].Extract)
	}
}

func TestWikipediaRejectsDisambiguationMissingAndOtherEntities(t *testing.T) {
	// Q190069 is the "Matrix" disambiguation page; Q1 matches nothing; the missing page has no pageprops.
	got, err := wikipedia.ParseExtracts(fixture(t), map[string]bool{"Q190069": true, "Q1": true}, "en")
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 0 {
		t.Fatalf("payloads = %+v", got)
	}
}

func TestWikipediaRejectsInvalidPayload(t *testing.T) {
	if _, err := wikipedia.ParseExtracts([]byte(`{"query":{"pages":"x"}}`), nil, "en"); err == nil {
		t.Fatal("want decode error")
	}
}

func TestWikipediaSupportsOnlyEnPtEs(t *testing.T) {
	for _, l := range []string{"en", "pt", "es"} {
		if _, err := wikipedia.New(l); err != nil {
			t.Errorf("%s: %v", l, err)
		}
	}

	if _, err := wikipedia.New("fr"); err == nil {
		t.Fatal("fr must be rejected")
	}
}

func TestWikipediaSource(t *testing.T) {
	w, _ := wikipedia.New("pt")
	if w.Source() != "wikipedia:pt" {
		t.Errorf("source = %q", w.Source())
	}
}

// server answers titles "T<n>" with a page for Q<n>, tracks concurrency, and optionally fails.
type server struct {
	now, max atomic.Int32
	mu       sync.Mutex
	sizes    []int
	fail     func(titles []string) bool
}

func (s *server) RoundTrip(req *http.Request) (*http.Response, error) {
	titles := strings.Split(req.URL.Query().Get("titles"), "|")

	s.mu.Lock()
	s.sizes = append(s.sizes, len(titles))
	s.mu.Unlock()

	n := s.now.Add(1)
	for {
		m := s.max.Load()
		if n <= m || s.max.CompareAndSwap(m, n) {
			break
		}
	}

	time.Sleep(20 * time.Millisecond)
	s.now.Add(-1)

	status, body := http.StatusOK, ""
	if s.fail != nil && s.fail(titles) {
		status = http.StatusBadRequest
	} else {
		var pages []map[string]any
		for _, title := range titles {
			pages = append(pages, map[string]any{
				"title":     title,
				"extract":   title + " intro.",
				"fullurl":   "https://en.wikipedia.org/wiki/" + title,
				"pageprops": map[string]string{"wikibase_item": "Q" + title[1:]},
			})
		}

		b, _ := json.Marshal(map[string]any{"query": map[string]any{"pages": pages}})
		body = string(b)
	}

	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func client(s *server) *ingest.Client {
	return httpx.New(httpx.Config{
		ContactURL: "https://hexascore.app/bot",
		Hosts:      map[string]httpx.HostLimit{"en.wikipedia.org": {PerSecond: 1000, InFlight: 4}},
		Base:       s,
	})
}

func targets(from, to int) []ingest.Target {
	var out []ingest.Target
	for n := from; n < to; n++ {
		out = append(out, ingest.Target{ID: fmt.Sprintf("Q%d", n), Type: fmt.Sprintf("T%d", n)})
	}

	return out
}

func TestWikipediaFetchRunsRequestsConcurrently(t *testing.T) {
	s := &server{}
	w, _ := wikipedia.New("en")

	// 85 targets -> batches of 20, 20, 20, 20, 5.
	recs, errs := w.Fetch(t.Context(), client(s), targets(0, 85))
	if len(errs) != 0 {
		t.Fatalf("errors: %+v", errs)
	}

	if len(recs) != 85 {
		t.Fatalf("records = %d", len(recs))
	}

	for i, r := range recs {
		if r.Source != "wikipedia:en" || r.ExternalID != fmt.Sprintf("Q%d", i) {
			t.Fatalf("record %d = %s %s: order must follow the targets", i, r.Source, r.ExternalID)
		}
	}

	slices.Sort(s.sizes)

	if !slices.Equal(s.sizes, []int{5, 20, 20, 20, 20}) {
		t.Errorf("batch sizes = %v", s.sizes)
	}

	if m := s.max.Load(); m < 2 || m > 4 {
		t.Errorf("max concurrent = %d, want 2..4", m)
	}
}

func TestWikipediaFailedBatchBecomesTargetErrors(t *testing.T) {
	s := &server{fail: func(titles []string) bool { return slices.Contains(titles, "T25") }}
	w, _ := wikipedia.New("en")

	recs, errs := w.Fetch(t.Context(), client(s), targets(0, 40))

	if len(recs) != 20 {
		t.Errorf("records = %d, want the 20 of the healthy batch", len(recs))
	}

	if len(errs) != 20 || errs[0].Target.ID != "Q20" || errs[0].Err == nil {
		t.Errorf("errors = %d first=%+v", len(errs), errs[0])
	}
}
