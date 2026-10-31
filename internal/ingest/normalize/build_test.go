package normalize_test

import (
	"encoding/json"
	"errors"
	"iter"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

var now = time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)

func read(t *testing.T, name string) []byte {
	t.Helper()

	raw, err := os.ReadFile("../testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}

	return raw
}

// wikidataRecord returns the frozen trimmed payload of one recorded entity.
func wikidataRecord(t *testing.T, qid string) store.Record {
	t.Helper()

	var all map[string]json.RawMessage
	if err := json.Unmarshal(read(t, "normalize_wikidata_payloads.json"), &all); err != nil {
		t.Fatal(err)
	}

	p, ok := all[qid]
	if !ok {
		t.Fatalf("no recorded payload for %s", qid)
	}

	return store.Record{Source: "wikidata", ExternalID: qid, Payload: p}
}

func build(t *testing.T, qid string, recs ...store.Record) *normalize.Item {
	t.Helper()

	item, err := normalize.BuildItem(qid, recs, now)
	if err != nil {
		t.Fatal(err)
	}

	return item
}

func TestSlugify(t *testing.T) {
	for in, want := range map[string]string{
		"The Witcher 3: Wild Hunt":    "the-witcher-3-wild-hunt",
		"Amélie — Le Fabuleux Destin": "amelie-le-fabuleux-destin",
		"千と千尋の神隠し":                    "",
	} {
		if got := normalize.Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}

	long := normalize.Slugify(
		"a b c d e f g h i j k l m n o p q r s t u v w x y z a b c d e f g h i j k l m n o p q r s t u v w x y z",
	)
	if len(long) > 80 || long[len(long)-1] == '-' {
		t.Errorf("long slug = %q (%d)", long, len(long))
	}
}

func TestSlugCandidates(t *testing.T) {
	if got := normalize.SlugCandidates(
		"Dune",
		2021,
		"Q1",
	); !slices.Equal(
		got,
		[]string{"dune", "dune-2021", "dune-q1"},
	) {
		t.Errorf("got %v", got)
	}

	if got := normalize.SlugCandidates("Dune", 0, "Q1"); !slices.Equal(got, []string{"dune", "dune-q1"}) {
		t.Errorf("no year: %v", got)
	}

	if got := normalize.SlugCandidates("千と千尋", 0, "Q155653"); !slices.Equal(got, []string{"q155653"}) {
		t.Errorf("empty base: %v", got)
	}
}

func TestParseTimePrecisions(t *testing.T) {
	for _, c := range []struct {
		time      string
		precision float64
		want      string
		wantPrec  string
		ok        bool
	}{
		{"+1999-03-31T00:00:00Z", 11, "1999-03-31", "day", true},
		{"+1999-03-00T00:00:00Z", 10, "1999-03-01", "month", true},
		{"+1949-00-00T00:00:00Z", 9, "1949-01-01", "year", true},
		{"+1900-00-00T00:00:00Z", 7, "", "", false},
		{"-0500-00-00T00:00:00Z", 9, "", "", false},
	} {
		d, ok := normalize.ParseTime(map[string]any{"time": c.time, "precision": c.precision})
		if ok != c.ok || (ok && (d.Time.Format(time.DateOnly) != c.want || d.Precision != c.wantPrec)) {
			t.Errorf("%s/%v = %v %v", c.time, c.precision, d, ok)
		}
	}

	if _, ok := normalize.ParseTime("nope"); ok {
		t.Error("non-object must not parse")
	}
}

func TestMovieMapping(t *testing.T) {
	item := build(t, "Q83495", wikidataRecord(t, "Q83495"))

	if item.Type != "movie" || item.Popularity != 114 {
		t.Errorf("type/popularity = %s %v", item.Type, item.Popularity)
	}

	if got := item.ReleaseDate.Format(
		time.DateOnly,
	); got != "1999-03-31" || item.ReleasePrecision != "day" ||
		item.ReleaseStatus != "released" {
		t.Errorf("release = %s %s %s", got, item.ReleasePrecision, item.ReleaseStatus)
	}

	if item.ExternalIDs["imdb"] != "tt0133093" || item.ExternalIDs["tmdb"] != "603" {
		t.Errorf("external ids = %v", item.ExternalIDs)
	}

	if item.Meta.RuntimeMinutes != 136 {
		t.Errorf("runtime = %d", item.Meta.RuntimeMinutes)
	}

	for _, c := range []string{"Lana Wachowski", "Lilly Wachowski"} {
		if !slices.Contains(item.Meta.Creators, c) {
			t.Errorf("creators = %v", item.Meta.Creators)
		}
	}

	if !slices.Contains(item.Meta.Genres, "science fiction film") {
		t.Errorf("genres = %v", item.Meta.Genres)
	}

	if n := len(item.Meta.Trailers); n < 1 || n > 2 {
		t.Errorf("trailers = %v", item.Meta.Trailers)
	}

	if len(item.Meta.Platforms) != 0 {
		t.Errorf("platforms = %v", item.Meta.Platforms)
	}

	want := map[string]string{"en": "The Matrix", "pt-BR": "Matrix", "es": "The Matrix"}
	if diff := cmp.Diff(want, item.Titles); diff != "" {
		t.Errorf("titles (-want +got):\n%s", diff)
	}
}

func TestMovieTitleFallsBackToWikipediaSummary(t *testing.T) {
	rec := wikidataRecord(t, "Q83495")

	var p map[string]any
	if err := json.Unmarshal(rec.Payload, &p); err != nil {
		t.Fatal(err)
	}

	p["labels"] = map[string]string{"en": "The Matrix"}
	rec.Payload, _ = json.Marshal(p)

	pt := store.Record{
		Source: "wikipedia:pt", ExternalID: "Q83495",
		Payload: json.RawMessage(`{"qid":"Q83495","lang":"pt","title":"Matrix (filme)","extract":"x","url":"u"}`),
	}

	item := build(t, "Q83495", rec, pt)
	if item.Titles["pt-BR"] != "Matrix (filme)" || item.Titles["es"] != "" {
		t.Errorf("titles = %v", item.Titles)
	}
}

func TestSeriesMapping(t *testing.T) {
	item := build(t, "Q1079", wikidataRecord(t, "Q1079"))

	if got := item.ReleaseDate.Format(time.DateOnly); got != "2008-01-20" {
		t.Errorf("release = %s", got)
	}

	if item.Meta.Seasons != 5 || item.Meta.Episodes != 62 || item.ExternalIDs["tmdb"] != "1396" {
		t.Errorf("meta = %+v ids = %v", item.Meta, item.ExternalIDs)
	}
}

func TestBookAndGameMapping(t *testing.T) {
	book := build(t, "Q208460", wikidataRecord(t, "Q208460"))
	if book.ReleaseDate.Year() != 1949 || book.ReleasePrecision != "year" {
		t.Errorf("book release = %v %s (P577 wins over P571)", book.ReleaseDate, book.ReleasePrecision)
	}

	if !slices.Equal(book.Meta.Creators, []string{"George Orwell"}) {
		t.Errorf("creators = %v", book.Meta.Creators)
	}

	game := build(t, "Q4267401", wikidataRecord(t, "Q4267401"))
	if got := game.ReleaseDate.Format(time.DateOnly); got != "2015-05-19" {
		t.Errorf("game release = %s, want the earliest P577", got)
	}

	if game.ExternalIDs["steam"] != "292030" || game.ExternalIDs["gog"] != "game/the_witcher_3_wild_hunt" {
		t.Errorf("game ids = %v", game.ExternalIDs)
	}

	if game.Meta.RuntimeMinutes != 0 {
		t.Errorf("games have no runtime: %d", game.Meta.RuntimeMinutes)
	}
}

func TestFutureReleaseIsAnnounced(t *testing.T) {
	rec := wikidataRecord(t, "Q83495")

	var p map[string]any
	if err := json.Unmarshal(rec.Payload, &p); err != nil {
		t.Fatal(err)
	}

	p["claims"] = map[string]any{"P577": []any{map[string]any{
		"rank": "normal", "value": map[string]any{"time": "+2027-05-00T00:00:00Z", "precision": float64(10)},
	}}}
	rec.Payload, _ = json.Marshal(p)

	item := build(t, "Q83495", rec)
	if item.ReleaseStatus != "announced" || item.ReleasePrecision != "month" {
		t.Errorf("status/precision = %s %s", item.ReleaseStatus, item.ReleasePrecision)
	}

	p["claims"] = map[string]any{}
	rec.Payload, _ = json.Marshal(p)

	item = build(t, "Q83495", rec)
	if item.ReleaseDate != nil || item.ReleasePrecision != "unknown" || item.ReleaseStatus != "released" {
		t.Errorf("undated = %v %s %s", item.ReleaseDate, item.ReleasePrecision, item.ReleaseStatus)
	}
}

func TestItemsWithoutWikidataOrEnglishTitleAreSkipped(t *testing.T) {
	steam := store.Record{Source: "steam", ExternalID: "Q1", Payload: json.RawMessage(`{"platforms":[]}`)}
	if _, err := normalize.BuildItem("Q1", []store.Record{steam}, now); !errors.Is(err, normalize.ErrNoWikidata) {
		t.Errorf("err = %v", err)
	}

	rec := wikidataRecord(t, "Q83495")

	var p map[string]any
	if err := json.Unmarshal(rec.Payload, &p); err != nil {
		t.Fatal(err)
	}

	p["labels"], p["sitelinks"] = map[string]string{"pt": "Matrix"}, map[string]string{}
	rec.Payload, _ = json.Marshal(p)

	if _, err := normalize.BuildItem("Q83495", []store.Record{rec}, now); !errors.Is(err, normalize.ErrNoEnglishTitle) {
		t.Errorf("err = %v", err)
	}
}

func TestBuildItemOrderIndependent(t *testing.T) {
	recs := []store.Record{
		wikidataRecord(t, "Q83495"),
		{
			Source:     "wikipedia:en",
			ExternalID: "Q83495",
			Payload:    json.RawMessage(`{"qid":"Q83495","lang":"en","title":"The Matrix","extract":"x","url":"u"}`),
		},
		{
			Source:     "wikipedia:pt",
			ExternalID: "Q83495",
			Payload:    json.RawMessage(`{"qid":"Q83495","lang":"pt","title":"Matrix","extract":"x","url":"u"}`),
		},
		{Source: "steam", ExternalID: "Q83495", Payload: json.RawMessage(`{}`)},
	}

	want := build(t, "Q83495", recs...)

	// every permutation, not a random sample: 4 records give 24 orders
	for shuffled := range permutations(recs) {
		if diff := cmp.Diff(want, build(t, "Q83495", shuffled...)); diff != "" {
			t.Fatalf("order changed the output (-want +got):\n%s", diff)
		}
	}
}

func permutations(recs []store.Record) iter.Seq[[]store.Record] {
	return func(yield func([]store.Record) bool) {
		var rec func(k int, cur []store.Record) bool

		rec = func(k int, cur []store.Record) bool {
			if k == len(cur) {
				return yield(slices.Clone(cur))
			}

			for i := k; i < len(cur); i++ {
				cur[k], cur[i] = cur[i], cur[k]
				ok := rec(k+1, cur)
				cur[k], cur[i] = cur[i], cur[k]

				if !ok {
					return false
				}
			}

			return true
		}

		rec(0, slices.Clone(recs))
	}
}
