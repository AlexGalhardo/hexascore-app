package wikidata_test

import (
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikidata"
)

var _ ingest.Connector = (*wikidata.Connector)(nil)

// api serves the recorded wbgetentities responses filtered to the requested ids.
type api struct {
	mu       sync.Mutex
	entities map[string]json.RawMessage
	requests []url.Values
}

func loadAPI(t *testing.T) *api {
	t.Helper()

	a := &api{entities: map[string]json.RawMessage{}}

	// entities first so a QID in both files keeps its claims
	for _, name := range []string{"wikidata_labels.json", "wikidata_entities.json"} {
		raw, err := os.ReadFile("../testdata/" + name)
		if err != nil {
			t.Fatal(err)
		}

		var doc struct {
			Entities map[string]json.RawMessage `json:"entities"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}

		maps.Copy(a.entities, doc.Entities)
	}

	return a
}

func (a *api) RoundTrip(req *http.Request) (*http.Response, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	q := req.URL.Query()
	a.requests = append(a.requests, q)

	out := map[string]json.RawMessage{}

	for id := range strings.SplitSeq(q.Get("ids"), "|") {
		e, ok := a.entities[id]
		if !ok {
			e = json.RawMessage(`{"id":"` + id + `","missing":""}`)
		}

		out[id] = e
	}

	body, _ := json.Marshal(map[string]any{"entities": out})

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(strings.NewReader(string(body))),
		Request:    req,
	}, nil
}

func fetcher(t *testing.T, a *api) (*wikidata.Connector, *ingest.Client) {
	t.Helper()

	c := httpx.New(httpx.Config{
		ContactURL: "https://hexascore.app/bot",
		Hosts:      map[string]httpx.HostLimit{"www.wikidata.org": {PerSecond: 1000, InFlight: 2}},
		Base:       a,
	})

	return wikidata.New(c), c
}

var recordedTypes = map[string]string{"Q83495": "movie", "Q4267401": "game", "Q1079": "series", "Q208460": "book"}

func fetchRecorded(t *testing.T) (map[string]map[string]any, []ingest.Record, *api) {
	t.Helper()

	a := loadAPI(t)
	w, c := fetcher(t, a)

	var batch []ingest.Target
	for _, id := range []string{"Q83495", "Q4267401", "Q1079", "Q208460"} {
		batch = append(batch, ingest.Target{ID: id, Type: recordedTypes[id]})
	}

	recs, errs := w.Fetch(t.Context(), c, batch)
	if len(errs) != 0 {
		t.Fatalf("errors: %+v", errs)
	}

	main := map[string]map[string]any{}

	for _, r := range recs {
		if r.Source != "wikidata" {
			continue
		}

		var p map[string]any
		if err := json.Unmarshal(r.Payload, &p); err != nil {
			t.Fatal(err)
		}

		main[r.ExternalID] = p
	}

	return main, recs, a
}

func TestTrimEntityKeepsOnlyUsedData(t *testing.T) {
	main, recs, _ := fetchRecorded(t)

	matrix := main["Q83495"]
	if matrix["type"] != "movie" {
		t.Errorf("type = %v", matrix["type"])
	}

	if matrix["labels"].(map[string]any)["en"] != "The Matrix" {
		t.Errorf("labels = %v", matrix["labels"])
	}

	if matrix["sitelinks"].(map[string]any)["enwiki"] != "The Matrix" {
		t.Errorf("sitelinks = %v", matrix["sitelinks"])
	}

	if n, _ := matrix["sitelinks_count"].(float64); n <= 50 {
		t.Errorf("sitelinks_count = %v", matrix["sitelinks_count"])
	}

	claims := matrix["claims"].(map[string]any)
	for prop := range claims {
		if !wikidata.KeptProps[prop] {
			t.Errorf("claim %s is not a kept property", prop)
		}
	}

	imdb := claims["P345"].([]any)[0].(map[string]any)["value"]
	if imdb != "tt0133093" {
		t.Errorf("P345 = %v", imdb)
	}

	runtime := claims["P2047"].([]any)[0].(map[string]any)["value"].(map[string]any)
	if runtime["amount"] != "+136" || runtime["unit"] != "Q7727" {
		t.Errorf("P2047 = %v", runtime)
	}

	refs := matrix["refs"].(map[string]any)
	for _, q := range []string{"Q9544977", "Q9545711"} { // one of the Wachowskis only has a "mul" label
		if _, ok := refs[q]; !ok {
			t.Errorf("refs lacks %s", q)
		}
	}

	var entities int

	for _, r := range recs {
		if r.Source == "wikidata:entity" {
			entities++
		}
	}

	if entities == 0 {
		t.Error("no wikidata:entity records for the referenced QIDs")
	}
}

// Deliberate divergence from legacy test_score_statements_keep_their_qualifiers:
// D-09 has no critic scores, so P444 is not kept at all.
func TestScoreStatementsAreDroppedDeliberateDivergenceFromLegacy(t *testing.T) {
	main, _, _ := fetchRecorded(t)

	if _, ok := main["Q83495"]["claims"].(map[string]any)["P444"]; ok {
		t.Error("P444 score statements must be removed")
	}
}

func TestTrimKeepsAliasesAndSeriesProperty(t *testing.T) {
	raw := `{"id":"Q5","labels":{"en":{"language":"en","value":"A"}},
	"aliases":{"en":[{"language":"en","value":"Alpha"},{"language":"en","value":"Aye"}],"fr":[{"language":"fr","value":"x"}]},
	"claims":{"P144":[{"mainsnak":{"snaktype":"value","datavalue":{"type":"wikibase-entityid","value":{"id":"Q9"}}},"rank":"normal"}]},
	"sitelinks":{}}`

	e, err := wikidata.ParseEntity([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}

	p := wikidata.Trim(e, "book")

	if !slices.Equal(p.Aliases["en"], []string{"Alpha", "Aye"}) || len(p.Aliases) != 1 {
		t.Errorf("aliases = %v", p.Aliases)
	}

	if got := p.Claims["P144"]; len(got) != 1 || got[0].Value != "Q9" {
		t.Errorf("P144 = %v", got)
	}

	if !slices.Equal(wikidata.ReferencedQids(p), []string{"Q9"}) {
		t.Errorf("refs = %v", wikidata.ReferencedQids(p))
	}
}

func TestTrimKeepsTitleClaimsAndLatnAliases(t *testing.T) {
	raw := `{"id":"Q5","labels":{},
	"aliases":{"ja-latn":[{"language":"ja-latn","value":"Kimi"}],"fr-latn":[{"language":"fr-latn","value":"y"}],"fr":[{"language":"fr","value":"x"}]},
	"claims":{
	"P1476":[{"mainsnak":{"snaktype":"value","datavalue":{"type":"monolingualtext","value":{"text":"Orig","language":"ja"}}},"rank":"normal"}],
	"P1448":[{"mainsnak":{"snaktype":"value","datavalue":{"type":"monolingualtext","value":{"text":"Off","language":"ja"}}},"rank":"normal"}],
	"P2561":[{"mainsnak":{"snaktype":"value","datavalue":{"type":"monolingualtext","value":{"text":"Old","language":"en"}}},"rank":"normal"}]},
	"sitelinks":{}}`

	e, err := wikidata.ParseEntity([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}

	p := wikidata.Trim(e, "series")

	for prop, want := range map[string]string{"P1476": "Orig", "P1448": "Off", "P2561": "Old"} {
		if got := p.Claims[prop]; len(got) != 1 || got[0].Value != want {
			t.Errorf("%s = %v", prop, got)
		}
	}

	if !slices.Equal(p.Aliases["ja-latn"], []string{"Kimi"}) || len(p.Aliases) != 2 {
		t.Errorf("aliases = %v", p.Aliases)
	}
}

func TestMissingEntitiesAreSkipped(t *testing.T) {
	a := loadAPI(t)
	w, c := fetcher(t, a)

	recs, errs := w.Fetch(t.Context(), c, []ingest.Target{{ID: "Q999999999", Type: "movie"}})
	if len(recs) != 0 || len(errs) != 0 {
		t.Fatalf("records %v errors %v", recs, errs)
	}
}

func TestInvalidEntityPayloadIsRejected(t *testing.T) {
	_, err := wikidata.ParseEntity(
		[]byte(`{"id":"Q1","claims":{"P31":[{"mainsnak":{"snaktype":"value"},"rank":"weird"}]}}`),
	)
	if err == nil {
		t.Fatal("rank weird must be rejected")
	}

	_, err = wikidata.ParseEntity([]byte(`{"id":"Q1","claims":{"P31":[{"rank":"normal"}]}}`))
	if err == nil {
		t.Fatal("a statement without mainsnak must be rejected")
	}
}

func TestFetchBatchesFiftyIds(t *testing.T) {
	a := loadAPI(t)
	w, c := fetcher(t, a)

	var batch []ingest.Target
	for i := range 120 {
		batch = append(
			batch,
			ingest.Target{
				ID:   "Q9" + strings.Repeat("0", 3) + string(rune('A'+i%26)) + string(rune('a'+i/26)),
				Type: "movie",
			},
		)
	}

	_, _ = w.Fetch(t.Context(), c, batch)

	var sizes []int
	for _, q := range a.requests {
		sizes = append(sizes, len(strings.Split(q.Get("ids"), "|")))
	}

	if !slices.Equal(sizes, []int{50, 50, 20}) {
		t.Errorf("request sizes = %v", sizes)
	}

	if a.requests[0].Get("props") != "labels|aliases|claims|sitelinks" {
		t.Errorf("props = %q", a.requests[0].Get("props"))
	}
}

// The reference hashes come from the legacy Python trim_entity with the same
// deliberate changes (no P444, aliases incl. *-latn, role and title properties); see
// testdata/trim_reference.py.
func TestTrimHashMatchesPython(t *testing.T) {
	main, _, _ := fetchRecorded(t)

	raw, err := os.ReadFile("../testdata/trim_hashes.json")
	if err != nil {
		t.Fatal(err)
	}

	var want map[string]string
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}

	if len(want) != len(recordedTypes) {
		t.Fatalf("reference has %d hashes", len(want))
	}

	for qid, payload := range main {
		b, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}

		got, err := store.CanonicalHash(b)
		if err != nil {
			t.Fatal(err)
		}

		if got != want[qid] {
			t.Errorf("%s: hash %s, python %s", qid, got, want[qid])
		}
	}
}
