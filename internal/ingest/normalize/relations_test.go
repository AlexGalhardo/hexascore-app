package normalize_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

func syntheticRecord(t *testing.T, qid string) store.Record {
	t.Helper()

	var f struct {
		Items map[string]json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(read(t, "parity_based_on.json"), &f); err != nil {
		t.Fatal(err)
	}

	return store.Record{Source: "wikidata", ExternalID: qid, Payload: f.Items[qid]}
}

func relations(t *testing.T, rec store.Record, known ...string) *normalize.Relations {
	t.Helper()

	set := map[string]bool{}
	for _, k := range known {
		set[k] = true
	}

	r, err := normalize.ExtractRelations(rec.ExternalID, []store.Record{rec}, set)
	if err != nil {
		t.Fatal(err)
	}

	return r
}

func byRole(r *normalize.Relations, role string) []normalize.Credit {
	return slices.DeleteFunc(slices.Clone(r.Credits), func(c normalize.Credit) bool { return c.Role != role })
}

func TestRelationsByRole(t *testing.T) {
	matrix := relations(t, wikidataRecord(t, "Q83495"))
	if d := byRole(
		matrix,
		"director",
	); len(d) != 2 || d[0].Kind != "person" || d[0].Position != 0 ||
		d[1].Position != 1 {
		t.Errorf("directors = %+v", d)
	}

	if len(byRole(matrix, "genre")) != 9 || len(byRole(matrix, "country")) != 2 {
		t.Errorf("genres/countries = %d/%d", len(byRole(matrix, "genre")), len(byRole(matrix, "country")))
	}

	game := relations(t, wikidataRecord(t, "Q4267401"))
	if d := byRole(game, "developer"); len(d) != 1 || d[0].Kind != "organization" || d[0].Name != "CD Projekt RED" {
		t.Errorf("developer = %+v", d)
	}

	if p := byRole(game, "platform"); len(p) != 6 || p[0].Kind != "platform" {
		t.Errorf("platforms = %+v", p)
	}

	if a := byRole(relations(t, wikidataRecord(t, "Q208460")), "author"); len(a) != 1 || a[0].Name != "George Orwell" {
		t.Errorf("author = %+v", a)
	}

	if c := byRole(relations(t, wikidataRecord(t, "Q1079")), "creator"); len(c) != 1 {
		t.Errorf("series creator = %+v", c)
	}
}

func TestRelationsCastIsCappedAtFifteenInOrder(t *testing.T) {
	cast := byRole(relations(t, syntheticRecord(t, "Q900001")), "cast")
	if len(cast) != 15 || cast[0].QID != "Q910000" || cast[14].QID != "Q910014" || cast[14].Position != 14 {
		t.Fatalf("cast = %d first=%+v last=%+v", len(cast), cast[0], cast[len(cast)-1])
	}
}

func TestRelationsDropRolesIncompatibleWithTheItemType(t *testing.T) {
	// the anime (series) carries P50 author, which is a book role
	r := relations(t, syntheticRecord(t, "Q900001"))
	if a := byRole(r, "author"); len(a) != 0 {
		t.Errorf("author on a series = %+v", a)
	}

	if n := byRole(r, "network"); len(n) != 1 || n[0].Name != "Example Network" {
		t.Errorf("network = %+v", n)
	}

	// network is series-only
	film := syntheticRecord(t, "Q900004")
	if n := byRole(relations(t, film), "network"); len(n) != 0 {
		t.Errorf("network on a movie = %+v", n)
	}
}

func TestAltTitles(t *testing.T) {
	got := relations(t, syntheticRecord(t, "Q900001")).AltTitles
	want := []normalize.AltTitle{
		{Title: "Shingeki Example", Lang: "ja-Latn", Kind: "transliteration", Source: "wikidata"},
		{Title: "Anime Example", Lang: "en", Kind: "alias", Source: "wikidata"},
		{Title: "EX", Lang: "", Kind: "alias", Source: "wikidata"},
	}

	if !slices.Equal(got, want) {
		t.Errorf("alt titles = %+v", got)
	}

	many := relations(t, syntheticRecord(t, "Q900007")).AltTitles
	if len(many) != normalize.MaxAltTitles || many[0].Title != "Alias 0" {
		t.Errorf("cap: %d first=%+v", len(many), many[0])
	}
}

func TestBasedOnLink(t *testing.T) {
	anime := relations(t, syntheticRecord(t, "Q900001"), "Q900002", "Q900001")
	if want := []normalize.Link{
		{FromQID: "Q900001", ToQID: "Q900002", Relation: "based_on", Source: "wikidata"},
	}; !slices.Equal(
		anime.Links,
		want,
	) {
		t.Errorf("anime links = %+v (duplicate, unknown and deprecated targets must not link)", anime.Links)
	}

	film := relations(t, syntheticRecord(t, "Q900004"), "Q900004", "Q900005")
	if len(film.Links) != 1 || film.Links[0].ToQID != "Q900005" {
		t.Errorf("film links = %+v", film.Links)
	}

	if none := relations(t, syntheticRecord(t, "Q900004")); len(none.Links) != 0 {
		t.Errorf("no known catalogue items, links = %+v", none.Links)
	}
}
