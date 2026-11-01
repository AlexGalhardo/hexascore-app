package openlibrary_test

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/openlibrary"
)

var _ ingest.Connector = (*openlibrary.Connector)(nil)

func fixture(t *testing.T, name string) []byte {
	t.Helper()

	raw, err := os.ReadFile("../testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}

	return raw
}

func TestOpenlibraryPathsPreferEditions(t *testing.T) {
	cases := []struct {
		ids  map[string]string
		want string
	}{
		{map[string]string{"isbn13": "978-0-451-52493-5"}, "/isbn/9780451524935.json"},
		{map[string]string{"openlibrary": "OL1M"}, "/books/OL1M.json"},
		{map[string]string{"openlibrary": "OL1168083W"}, "/works/OL1168083W.json"},
		{map[string]string{"openlibrary": "OL23A"}, ""},
		{map[string]string{"isbn13": "978-0-451-52493-5", "openlibrary": "OL1168083W"}, "/isbn/9780451524935.json"},
	}

	for _, c := range cases {
		if got := openlibrary.DocumentPath(c.ids); got != c.want {
			t.Errorf("DocumentPath(%v) = %q, want %q", c.ids, got, c.want)
		}
	}
}

func TestOpenlibraryEdition(t *testing.T) {
	p, err := openlibrary.ParseDoc(fixture(t, "openlibrary_isbn.json"), "Q208460", nil)
	if err != nil {
		t.Fatal(err)
	}

	want := openlibrary.Payload{Qid: "Q208460", Key: "/books/OL62592264M", Covers: []int{15257377}}
	if p.Qid != want.Qid || p.Key != want.Key || p.Pages != nil || p.Overview != nil ||
		!slices.Equal(p.Covers, want.Covers) {
		t.Errorf("payload = %+v, want %+v", p, want)
	}
}

func TestOpenlibraryWorkFillsWhatTheEditionLacks(t *testing.T) {
	edition := fixture(t, "openlibrary_isbn.json")
	work := fixture(t, "openlibrary_work.json")

	if got := openlibrary.WorkPath(edition); got != "/works/OL1168083W.json" {
		t.Errorf("WorkPath(edition) = %q", got)
	}

	if got := openlibrary.WorkPath(work); got != "" {
		t.Errorf("WorkPath(work) = %q, want empty", got)
	}

	var doc map[string]any
	if err := json.Unmarshal(edition, &doc); err != nil {
		t.Fatal(err)
	}

	doc["covers"] = []int{-1}

	noCover, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}

	p, err := openlibrary.ParseDoc(noCover, "Q208460", work)
	if err != nil {
		t.Fatal(err)
	}

	var w struct {
		Covers []int `json:"covers"`
	}
	if err := json.Unmarshal(work, &w); err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(p.Covers, w.Covers[:min(3, len(w.Covers))]) {
		t.Errorf("covers = %v, want the work's first three %v", p.Covers, w.Covers)
	}

	if p.Overview == nil || *p.Overview == "" {
		t.Error("the work's description must fill the overview")
	}

	if p.Key != "/books/OL62592264M" {
		t.Errorf("key = %q, the edition key must be kept", p.Key)
	}

	own, err := openlibrary.ParseDoc(edition, "Q208460", work)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(own.Covers, []int{15257377}) {
		t.Errorf("the edition's own cover must win: %v", own.Covers)
	}
}

func TestOpenlibraryReadsTypedDescriptions(t *testing.T) {
	raw := []byte(`{"key": "/works/OL1W", "description": {"type": "/type/text", "value": " A story. "}}`)

	p, err := openlibrary.ParseDoc(raw, "Q1", nil)
	if err != nil {
		t.Fatal(err)
	}

	if p.Overview == nil || *p.Overview != "A story." {
		t.Errorf("overview = %v", p.Overview)
	}

	plain, err := openlibrary.ParseDoc([]byte(`{"key": "/works/OL1W", "description": "  "}`), "Q1", nil)
	if err != nil {
		t.Fatal(err)
	}

	if plain.Overview != nil {
		t.Errorf("a blank description must be null, got %q", *plain.Overview)
	}
}

func TestOpenlibraryDropsPlaceholderCovers(t *testing.T) {
	p, err := openlibrary.ParseDoc(
		[]byte(`{"key": "/works/OL1W", "covers": [-1, 5], "number_of_pages": 320}`),
		"Q1",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(p.Covers, []int{5}) {
		t.Errorf("covers = %v, want [5]", p.Covers)
	}

	if p.Pages == nil || *p.Pages != 320 {
		t.Errorf("pages = %v, want 320", p.Pages)
	}
}
