package catalog

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/AlexGalhardo/hexascore-app/config"
)

// fixtureItem mixes open and TMDB/IGDB-sourced fields, like an enriched game.
func fixtureItem() ItemSource {
	return ItemSource{
		ID: "0192f1c6-0000-7000-8000-000000000001", Type: "game", Slug: "half-life", WikidataQID: "Q279744",
		Title:            Sourced[string]{"Half-Life", "wikidata"},
		Synopsis:         &SynopsisIn{Text: "A scientist...", Source: "igdb", Licence: "proprietary-igdb"},
		ReleaseDate:      &Sourced[time.Time]{time.Date(1998, 11, 19, 0, 0, 0, 0, time.UTC), "wikidata"},
		ReleasePrecision: "day",
		RuntimeMinutes:   &Sourced[int]{720, "tmdb"},
		Cover:            &Sourced[string]{"https://cdn.example/c.webp", "igdb"},
		ExternalIDs:      map[string]string{"wikidata": "Q279744", "igdb": "231", "steam": "70"},
		Entities: []EntityRefIn{
			{
				ID:         "e1",
				Kind:       "organization",
				Role:       "developer",
				Name:       "Valve",
				NameSource: "wikidata",
				Source:     "wikidata",
			},
			{ID: "e2", Kind: "person", Role: "composer", Name: "Kelly Bailey", Source: "igdb"},
		},
		AffiliateLinks: []string{"https://store.example/hl"},
	}
}

func TestVisibility(t *testing.T) {
	sources := config.LoadCatalog(func(string) string { return "" }).APISources

	site := Assemble(fixtureItem(), NewPolicy(AudienceSite, sources))
	if site.Synopsis == nil || site.RuntimeMinutes == nil || site.Cover == "" || site.ExternalIDs["igdb"] == "" ||
		len(site.Entities) != 2 || len(site.AffiliateLinks) != 1 || site.Meta != nil {
		t.Fatalf("site must see every field, got %+v", site)
	}

	api := Assemble(fixtureItem(), NewPolicy(AudienceAPI, sources))
	if api.Title != "Half-Life" || api.ReleaseDate != "1998-11-19" || api.ExternalIDs["wikidata"] != "Q279744" {
		t.Fatalf("api lost open fields: %+v", api)
	}
	want := []string{"cover", "entities", "externalIds", "runtimeMinutes", "synopsis"}
	if api.Meta == nil {
		t.Fatal("api meta missing")
	}
	if diff := cmp.Diff(want, api.Meta.OmittedFields); diff != "" {
		t.Fatalf("omittedFields (-want +got):\n%s", diff)
	}
	if _, ok := api.ExternalIDs["igdb"]; ok {
		t.Fatal("igdb id must not reach the api")
	}
	if len(api.Entities) != 1 || len(api.AffiliateLinks) != 0 {
		t.Fatalf("api entities/affiliates: %+v %+v", api.Entities, api.AffiliateLinks)
	}

	raw, err := json.Marshal(api)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"synopsis":`, `"runtimeMinutes":`, `"cover":`, `"affiliateLinks":`, "null"} {
		if strings.Contains(string(raw), key) {
			t.Errorf("api JSON must omit %s, got %s", key, raw)
		}
	}
}

func TestVisibilityOpenLibraryLimited(t *testing.T) {
	src := fixtureItem()
	src.Title = Sourced[string]{"Dom Casmurro", "openlibrary"}
	src.Synopsis = &SynopsisIn{Text: "...", Source: "openlibrary", Licence: "unknown"}
	api := Assemble(src, NewPolicy(AudienceAPI, config.DefaultCatalogAPISources))
	if api.Title != "Dom Casmurro" || api.Synopsis != nil {
		t.Fatalf("openlibrary: titles yes, content no (01-2); got title %q synopsis %+v", api.Title, api.Synopsis)
	}
}

func TestVisibilityUnknownSourceFailsClosed(t *testing.T) {
	src := fixtureItem()
	src.Title = Sourced[string]{"X", ""}
	if got := Assemble(src, NewPolicy(AudienceAPI, config.DefaultCatalogAPISources)).Title; got != "" {
		t.Fatalf("unknown source must be hidden from the api, got %q", got)
	}
}

func TestLoadCatalogParses(t *testing.T) {
	got := config.LoadCatalog(func(string) string { return " Wikidata, ,TMDB " }).APISources
	if diff := cmp.Diff([]string{"wikidata", "tmdb"}, got); diff != "" {
		t.Fatal(diff)
	}
}
