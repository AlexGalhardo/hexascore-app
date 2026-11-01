package normalize_test

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

// source returns a stored record of the given source for qid from the
// payloads written by testdata/enrich_reference.py.
func source(t *testing.T, key, src, qid string) store.Record {
	t.Helper()

	var all map[string]json.RawMessage
	if err := json.Unmarshal(read(t, "enrich_payloads.json"), &all); err != nil {
		t.Fatal(err)
	}

	p, ok := all[key]
	if !ok {
		t.Fatalf("no enrich payload %s", key)
	}

	return store.Record{Source: src, ExternalID: qid, Payload: p}
}

// without drops claim properties from a wikidata record.
func without(t *testing.T, r store.Record, props ...string) store.Record {
	t.Helper()

	var p map[string]any
	if err := json.Unmarshal(r.Payload, &p); err != nil {
		t.Fatal(err)
	}

	claims, _ := p["claims"].(map[string]any)
	for _, prop := range props {
		delete(claims, prop)
	}

	if len(props) == 0 {
		p["claims"] = map[string]any{}
	}

	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}

	r.Payload = raw

	return r
}

func jsonRecord(src, qid, payload string) store.Record {
	return store.Record{Source: src, ExternalID: qid, Payload: json.RawMessage(payload)}
}

func TestTmdbFillsOnlyWhatWikidataLacks(t *testing.T) {
	wd := wikidataRecord(t, "Q83495")
	tmdb := source(t, "tmdb:Q83495", "tmdb", "Q83495")

	full := build(t, "Q83495", wd, tmdb)
	if !slices.Contains(full.Meta.Genres, "science fiction film") ||
		slices.Contains(full.Meta.Genres, "Science Fiction") {
		t.Errorf("wikidata genres lost: %v", full.Meta.Genres)
	}

	if full.Provenance["meta.genres"] != "wikidata" {
		t.Errorf("genres provenance = %q", full.Provenance["meta.genres"])
	}

	bare := build(t, "Q83495", without(t, wd, "P136", "P2047", "P577", "P4947"), tmdb)

	if !slices.Equal(bare.Meta.Genres, []string{"Action", "Science Fiction"}) ||
		bare.Provenance["meta.genres"] != "tmdb" {
		t.Errorf("genres = %v from %q", bare.Meta.Genres, bare.Provenance["meta.genres"])
	}

	if bare.Meta.RuntimeMinutes != 136 || bare.Provenance["runtime_minutes"] != "tmdb" {
		t.Errorf("runtime = %d from %q", bare.Meta.RuntimeMinutes, bare.Provenance["runtime_minutes"])
	}

	if bare.ReleaseDate == nil || bare.ReleaseDate.Format("2006-01-02") != "1999-03-31" ||
		bare.ReleasePrecision != normalize.PrecisionDay || bare.Provenance["release_date"] != "tmdb" {
		t.Errorf("release = %v %q from %q", bare.ReleaseDate, bare.ReleasePrecision, bare.Provenance["release_date"])
	}

	if bare.ExternalIDs["tmdb"] != "603" || bare.Provenance["external_ids.tmdb"] != "tmdb" {
		t.Errorf("tmdb id = %q from %q", bare.ExternalIDs["tmdb"], bare.Provenance["external_ids.tmdb"])
	}

	if c := bare.Cover; c == nil || c.Source != "tmdb" || c.Attribution["source"] != "TMDB" ||
		c.Attribution["url"] != "https://www.themoviedb.org/movie/603" {
		t.Errorf("cover = %+v", c)
	}

	// TMDB only covers movies and series.
	game := build(
		t,
		"Q4267401",
		wikidataRecord(t, "Q4267401"),
		jsonRecord("tmdb", "Q4267401", `{"cover":"https://x/y.jpg"}`),
	)
	if game.Cover != nil {
		t.Errorf("tmdb cover on a game: %+v", game.Cover)
	}
}

func TestIgdbCoverWinsOverSteam(t *testing.T) {
	wd := wikidataRecord(t, "Q4267401")
	igdb := source(t, "igdb:Q4267401", "igdb", "Q4267401")
	steam := source(t, "steam:Q4267401", "steam", "Q4267401")

	item := build(t, "Q4267401", wd, steam, igdb)
	if c := item.Cover; c == nil || c.URL != "https://images.igdb.com/igdb/image/upload/t_cover_big_2x/coaarl.jpg" ||
		c.Source != "igdb" || c.Attribution["source"] != "IGDB" ||
		c.Attribution["url"] != "https://www.igdb.com/games/the-witcher-3-wild-hunt" {
		t.Errorf("cover = %+v", c)
	}

	if only := build(t, "Q4267401", wd, steam); only.Cover == nil || only.Cover.Source != "steam" ||
		only.Cover.Attribution != nil {
		t.Errorf("steam cover = %+v", only.Cover)
	}

	bare := build(t, "Q4267401", without(t, wd), steam, igdb)
	if !slices.Contains(bare.Meta.Platforms, "PlayStation 4") {
		t.Errorf("platforms = %v (igdb comes before steam)", bare.Meta.Platforms)
	}

	if !slices.Equal(bare.Meta.Creators, []string{"CD Projekt RED"}) {
		t.Errorf("creators = %v", bare.Meta.Creators)
	}

	if bare.ReleaseDate == nil || bare.ReleaseDate.Format("2006-01-02") != "2015-05-19" {
		t.Errorf("release = %v", bare.ReleaseDate)
	}
}

func TestGameMappingWithSteamEnrichment(t *testing.T) {
	item := build(t, "Q4267401", wikidataRecord(t, "Q4267401"), source(t, "steam:Q4267401", "steam", "Q4267401"))

	if item.ExternalIDs["steam"] != "292030" || item.ExternalIDs["gog"] != "game/the_witcher_3_wild_hunt" {
		t.Errorf("external ids = %v", item.ExternalIDs)
	}

	if !slices.Contains(item.Meta.Platforms, "Microsoft Windows") || item.Provenance["meta.platforms"] != "wikidata" {
		t.Errorf("wikidata must win over steam: %v from %q", item.Meta.Platforms, item.Provenance["meta.platforms"])
	}

	if !slices.Equal(item.Meta.Creators, []string{"CD Projekt RED"}) || item.Meta.RuntimeMinutes != 0 {
		t.Errorf("creators = %v runtime = %d", item.Meta.Creators, item.Meta.RuntimeMinutes)
	}

	bare := build(t, "Q4267401", without(t, wikidataRecord(t, "Q4267401"), "P400"),
		source(t, "steam:Q4267401", "steam", "Q4267401"))
	if !slices.Equal(bare.Meta.Platforms, []string{"Windows"}) || bare.Provenance["meta.platforms"] != "steam" {
		t.Errorf("steam fallback = %v from %q", bare.Meta.Platforms, bare.Provenance["meta.platforms"])
	}
}

func TestBookMappingUsesOpenLibraryPagesAsFallback(t *testing.T) {
	ol := source(t, "openlibrary:Q208460", "openlibrary", "Q208460")
	item := build(t, "Q208460", wikidataRecord(t, "Q208460"), ol)

	if item.ReleaseDate == nil || item.ReleaseDate.Year() != 1949 ||
		!slices.Equal(item.Meta.Creators, []string{"George Orwell"}) {
		t.Errorf("release = %v creators = %v", item.ReleaseDate, item.Meta.Creators)
	}

	if item.Meta.Pages != 328 || item.Provenance["page_count"] != "openlibrary" {
		t.Errorf("pages = %d from %q", item.Meta.Pages, item.Provenance["page_count"])
	}

	if c := item.Cover; c == nil || c.URL != "https://covers.openlibrary.org/b/id/15257377-L.jpg" ||
		c.Source != "openlibrary" {
		t.Errorf("cover = %+v", c)
	}
}

func TestOwnSynopsisOnlyFillsLocalesWithoutWikipedia(t *testing.T) {
	var extracts struct {
		Query struct {
			Pages []struct{ Extract, Title string }
		}
	}
	if err := json.Unmarshal(read(t, "wikipedia_extracts_en.json"), &extracts); err != nil {
		t.Fatal(err)
	}

	enPayload, _ := json.Marshal(map[string]string{
		"qid":     "Q83495",
		"lang":    "en",
		"title":   "The Matrix",
		"extract": "Wikipedia text.",
		"url":     "https://en.wikipedia.org/wiki/The_Matrix",
	})
	own := jsonRecord("synopsis", "Q83495", `{"en":"Own English text.","pt-BR":"Texto próprio.","es":""}`)
	wiki := store.Record{Source: "wikipedia:en", ExternalID: "Q83495", Payload: enPayload}

	item := build(t, "Q83495", wikidataRecord(t, "Q83495"), wiki, own)

	en := item.Synopses["en"]
	if en.Text != "Wikipedia text." || en.Source != "wikipedia" || en.Licence != "CC-BY-SA-4.0" ||
		en.URL != "https://en.wikipedia.org/wiki/The_Matrix" {
		t.Errorf("en = %+v", en)
	}

	pt := item.Synopses["pt-BR"]
	if pt.Text != "Texto próprio." || pt.Source != "hexascore" || pt.Licence != "proprietary" || pt.URL != "" {
		t.Errorf("pt-BR = %+v", pt)
	}

	if _, ok := item.Synopses["es"]; ok {
		t.Errorf("es has a synopsis: %+v", item.Synopses["es"])
	}

	if item.Provenance["synopsis.en"] != "wikipedia" || item.Provenance["synopsis.pt-BR"] != "hexascore" {
		t.Errorf("provenance = %v", item.Provenance)
	}
}

// TMDB, IGDB and Open Library carry a licensed overview; it is input for the
// rewriting step only and never becomes a translation (FR-06-3).
func TestThirdPartyTextNeverReachesTranslations(t *testing.T) {
	cases := []struct {
		qid  string
		recs []store.Record
	}{
		{"Q83495", []store.Record{source(t, "tmdb:Q83495", "tmdb", "Q83495")}},
		{"Q4267401", []store.Record{source(t, "igdb:Q4267401", "igdb", "Q4267401")}},
		{"Q208460", []store.Record{jsonRecord("openlibrary", "Q208460", `{"pages":9,"overview":"Licensed text."}`)}},
	}

	for _, c := range cases {
		item := build(t, c.qid, append(c.recs, wikidataRecord(t, c.qid))...)
		if len(item.Synopses) != 0 {
			t.Errorf("%s: synopses = %+v", c.qid, item.Synopses)
		}
	}
}

func TestCommonsCoverAttribution(t *testing.T) {
	var all map[string]json.RawMessage
	if err := json.Unmarshal(read(t, "enrich_payloads.json"), &all); err != nil {
		t.Fatal(err)
	}

	cover, err := normalize.ParseCommonsCover(all["commons"])
	if err != nil || cover == nil {
		t.Fatalf("cover = %v, err = %v", cover, err)
	}

	a := cover.Attribution
	if cover.Source != "commons" || cover.URL[:8] != "https://" || a["source"] != "Wikimedia Commons" ||
		a["license"] == "" || a["author"] == "" || a["url"] == "" {
		t.Errorf("cover = %+v", cover)
	}

	if a["author"] != "Gage Skidmore" && len(a["author"]) > 0 && a["author"][0] == '<' {
		t.Errorf("author keeps markup: %q", a["author"])
	}

	if none, err := normalize.ParseCommonsCover(
		[]byte(`{"query":{"pages":[{"missing":true}]}}`),
	); err != nil ||
		none != nil {
		t.Errorf("missing page = %v, %v", none, err)
	}
}

func TestProvenanceAndLicenceComeFromDataSources(t *testing.T) {
	item := build(t, "Q83495", wikidataRecord(t, "Q83495"))

	for field, want := range map[string]string{
		"translations.en.title": "wikidata", "release_date": "wikidata", "external_ids.imdb": "wikidata",
		"runtime_minutes": "wikidata",
	} {
		if item.Provenance[field] != want {
			t.Errorf("provenance[%s] = %q, want %s", field, item.Provenance[field], want)
		}
	}

	for _, code := range []string{"wikidata", "wikipedia", "tmdb", "igdb", "steam", "gog", "openlibrary", "hexascore", "commons"} {
		if normalize.Licence(code) == "" {
			t.Errorf("no licence for %s", code)
		}
	}

	for _, v := range item.Provenance {
		if normalize.Licence(v) == "" {
			t.Errorf("provenance source %q is not a data source", v)
		}
	}
}

type fakeCommunity struct {
	fields    map[string]normalize.CommunityField
	suggested []normalize.SystemSuggestion
}

func (f *fakeCommunity) CommunityFields(context.Context, string) (map[string]normalize.CommunityField, error) {
	return f.fields, nil
}

func (f *fakeCommunity) OpenSystemSuggestion(_ context.Context, s normalize.SystemSuggestion) error {
	f.suggested = append(f.suggested, s)

	return nil
}

func TestApplyCommunitySkipsAndSuggests(t *testing.T) {
	edited := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	title := json.RawMessage(`"Matrix, The"`)

	for _, c := range []struct {
		name      string
		current   json.RawMessage
		days      int
		wantSkip  bool
		wantSuggs int
	}{
		{"day 179 diverges", title, 179, true, 1},
		{"day 181 diverges", title, 181, true, 1},
		{"day 179 same value", json.RawMessage(`"The Matrix"`), 179, true, 0},
		{"day 181 same value", json.RawMessage(`"The Matrix"`), 181, false, 0},
	} {
		item := build(t, "Q83495", wikidataRecord(t, "Q83495"))
		fake := &fakeCommunity{fields: map[string]normalize.CommunityField{
			"translations.en.title": {Since: edited, Current: c.current},
		}}

		res, err := normalize.ApplyCommunity(t.Context(), fake, "item-1", item, edited.AddDate(0, 0, c.days))
		if err != nil {
			t.Fatal(err)
		}

		if item.Skip["translations.en.title"] != c.wantSkip {
			t.Errorf("%s: skip = %v", c.name, item.Skip)
		}

		if c.wantSkip && item.Provenance["translations.en.title"] != "" {
			t.Errorf("%s: provenance kept for a skipped field", c.name)
		}

		if len(fake.suggested) != c.wantSuggs || len(res.Suggested) != c.wantSuggs {
			t.Fatalf("%s: suggestions = %d", c.name, len(fake.suggested))
		}

		if c.wantSuggs == 1 {
			s := fake.suggested[0]
			if s.ItemID != "item-1" || s.Domain != "movie" || s.Field != "translations.en.title" ||
				string(s.Before) != `"Matrix, The"` || string(s.After) != `"The Matrix"` || s.Source != "wikidata" {
				t.Errorf("%s: suggestion = %+v", c.name, s)
			}
		}
	}
}
