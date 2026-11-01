package trailers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/trailers"
)

var _ ingest.Connector = (*trailers.Connector)(nil)

const (
	idA = "aaaaaaaaaaa"
	idB = "bbbbbbbbbbb"
	idC = "ccccccccccc"
	idD = "ddddddddddd"
)

func TestCandidatesOrderWikidataFirstAndDropMalformedIds(t *testing.T) {
	got := trailers.Candidates(trailers.Sources{
		Wikidata: []trailers.Claim{
			{Value: idB, Rank: "normal"},
			{Value: idA, Rank: "preferred"},
			{Value: "short", Rank: "normal"},
		},
		TMDB: []string{idC, idB, "not an id!"},
		IGDB: []string{idD},
	})

	if want := []string{idA, idB, idC, idD}; !slices.Equal(got, want) {
		t.Errorf("candidates = %v, want %v", got, want)
	}
}

// asker answers from a table and counts the questions.
type asker struct {
	verdict map[string]bool
	err     map[string]error
	asked   []string
}

func (a *asker) Available(_ context.Context, id string) (bool, error) {
	a.asked = append(a.asked, id)

	return a.verdict[id], a.err[id]
}

func TestCheckKeepsTwoAvailableAndStopsAsking(t *testing.T) {
	a := &asker{verdict: map[string]bool{idA: false, idB: true, idC: true, idD: true}}

	got, err := trailers.Check(t.Context(), a, []string{idA, idB, idC, idD}, trailers.Checked{})
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(got.Available, []string{idB, idC}) || !slices.Equal(got.Unavailable, []string{idA}) {
		t.Errorf("checked = %+v", got)
	}

	if slices.Contains(a.asked, idD) {
		t.Errorf("asked about %v after two were available", a.asked)
	}
}

func TestCheckReusesKnownVerdicts(t *testing.T) {
	a := &asker{verdict: map[string]bool{idC: true}}

	got, err := trailers.Check(t.Context(), a, []string{idA, idB, idC},
		trailers.Checked{Available: []string{idB}, Unavailable: []string{idA}})
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(a.asked, []string{idC}) || !slices.Equal(got.Available, []string{idB, idC}) {
		t.Errorf("asked %v, checked %+v", a.asked, got)
	}
}

func TestRateLimitIsNotAVerdict(t *testing.T) {
	status := map[string]int{idA: 404, idB: 429, idC: 503, idD: 200}
	o, _ := oembed(t, status)

	for id, wantOK := range map[string]bool{idA: false, idD: true} {
		ok, err := o.Available(t.Context(), id)
		if err != nil || ok != wantOK {
			t.Errorf("%s = %v, %v, want %v, nil", id, ok, err, wantOK)
		}
	}

	for _, id := range []string{idB, idC} {
		if _, err := o.Available(t.Context(), id); !errors.Is(err, trailers.ErrNoVerdict) {
			t.Errorf("%s error = %v, want ErrNoVerdict", id, err)
		}
	}

	a := &asker{verdict: map[string]bool{}, err: map[string]error{idA: trailers.ErrNoVerdict}}
	if _, err := trailers.Check(t.Context(), a, []string{idA}, trailers.Checked{}); err == nil {
		t.Error("Check swallowed a missing verdict")
	}
}

// stepClock never waits: Sleep moves time forward, so the client retries a 429 or 5xx at once.
type stepClock struct{ now time.Time }

func (c *stepClock) Now() time.Time { return c.now }

func (c *stepClock) Sleep(_ context.Context, d time.Duration) error {
	c.now = c.now.Add(d)

	return nil
}

type oembedFake struct {
	mu   sync.Mutex
	seen []string
}

// oembed serves the status of each video id and returns an OEmbed over it.
func oembed(t *testing.T, status map[string]int) (*trailers.OEmbed, *oembedFake) {
	t.Helper()

	f := &oembedFake{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Query().Get("url"), "https://www.youtube.com/watch?v=")

		f.mu.Lock()
		f.seen = append(f.seen, id)
		f.mu.Unlock()

		w.WriteHeader(status[id])
	}))
	t.Cleanup(srv.Close)

	client := httpx.New(httpx.Config{
		ContactURL: "https://hexascore.app/bot",
		Hosts:      map[string]httpx.HostLimit{strings.TrimPrefix(srv.URL, "http://"): {PerSecond: 1000, InFlight: 4}},
		Allow:      []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")},
		Clock:      &stepClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	})

	return &trailers.OEmbed{Client: client, URL: srv.URL}, f
}

func target(t *testing.T, in trailers.Input) ingest.Target {
	t.Helper()

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}

	return ingest.Target{ID: in.QID, Type: string(raw)}
}

func TestFetchWritesOneRecordPerItemAndSkipsFailures(t *testing.T) {
	o, _ := oembed(t, map[string]int{idA: 200, idB: 404, idC: 429})
	c := &trailers.Connector{OEmbedURL: o.URL}

	batch := []ingest.Target{
		target(
			t,
			trailers.Input{
				QID:     "Q1",
				Sources: trailers.Sources{Wikidata: []trailers.Claim{{Value: idA, Rank: "normal"}, {Value: idB}}},
			},
		),
		target(t, trailers.Input{QID: "Q2", Sources: trailers.Sources{TMDB: []string{idC}}}),
		target(t, trailers.Input{QID: "Q3"}),
	}

	records, errs := c.Fetch(t.Context(), o.Client, batch)

	if len(records) != 1 || records[0].Source != "youtube" || records[0].ExternalID != "Q1" {
		t.Fatalf("records = %+v", records)
	}

	var got trailers.Checked
	if err := json.Unmarshal(records[0].Payload, &got); err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(got.Available, []string{idA}) || !slices.Equal(got.Unavailable, []string{idB}) {
		t.Errorf("payload = %s", records[0].Payload)
	}

	if len(errs) != 1 || errs[0].Target.ID != "Q2" || !errors.Is(errs[0].Err, trailers.ErrNoVerdict) {
		t.Errorf(
			"errors = %+v, want Q2 with ErrNoVerdict (Q3 has no candidates and no stored verdict: skipped silently)",
			errs,
		)
	}
}

func TestNormalizeShowsOnlyCheckedTrailers(t *testing.T) {
	claims := make([]map[string]any, 0, 3)
	for _, id := range []string{idA, idB, idC} {
		claims = append(claims, map[string]any{"value": id, "rank": "normal"})
	}

	wd, err := json.Marshal(map[string]any{
		"type": "movie", "id": "Q9", "labels": map[string]string{"en": "A Movie"},
		"claims": map[string]any{"P1651": claims},
	})
	if err != nil {
		t.Fatal(err)
	}

	yt, err := json.Marshal(trailers.Checked{Available: []string{idB, idC}, Unavailable: []string{idA}})
	if err != nil {
		t.Fatal(err)
	}

	item, err := normalize.BuildItem("Q9", []store.Record{
		{Source: "wikidata", ExternalID: "Q9", Payload: wd},
		{Source: "youtube", ExternalID: "Q9", Payload: yt},
	}, time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	if want := []string{idB, idC}; !slices.Equal(item.Meta.Trailers, want) {
		t.Errorf("trailers = %v, want %v", item.Meta.Trailers, want)
	}
}
