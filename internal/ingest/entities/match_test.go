package entities

import (
	"context"
	"slices"
	"testing"
)

func TestNormalizeName(t *testing.T) {
	cases := map[string]string{
		"  Hayao   Miyazaki ": "hayao miyazaki",
		"Amélie Poulain":      "amelie poulain",
		"Søren Kierkegaard":   "soren kierkegaard",
		"Studio Ghibli, Inc.": "studio ghibli inc",
		"J.R.R. Tolkien":      "j r r tolkien",
		"宮崎 駿":                "宮崎 駿",
		"!!!":                 "",
	}

	for in, want := range cases {
		if got := NormalizeName(in); got != want {
			t.Errorf("NormalizeName(%q) = %q, want %q", in, got, want)
		}
	}
}

// fake is an in-memory Store.
type fake struct {
	ents    []fakeEntity
	linked  []string
	created []Credit
	reviews map[string][]string
}

type fakeEntity struct {
	e                        Entity
	qid, extKey, extID, name string
}

func (f *fake) ByQID(_ context.Context, qid string) (Entity, bool, error) {
	for _, e := range f.ents {
		if e.qid == qid {
			return e.e, true, nil
		}
	}

	return Entity{}, false, nil
}

func (f *fake) ByExternalID(_ context.Context, key, id, kind string) ([]Entity, error) {
	var out []Entity

	for _, e := range f.ents {
		if e.extKey == key && e.extID == id && e.e.Kind == kind {
			out = append(out, e.e)
		}
	}

	return out, nil
}

func (f *fake) ByName(_ context.Context, name, kind string) ([]Entity, error) {
	var out []Entity

	for _, e := range f.ents {
		if NormalizeName(e.name) == name && e.e.Kind == kind {
			out = append(out, e.e)
		}
	}

	return out, nil
}

func (f *fake) Create(_ context.Context, c Credit) (string, error) {
	f.created = append(f.created, c)
	id := "new-" + c.Name
	f.ents = append(
		f.ents,
		fakeEntity{
			e:      Entity{ID: id, Kind: c.Kind},
			qid:    c.QID,
			extKey: c.ExternalIDKey,
			extID:  c.ExternalID,
			name:   c.Name,
		},
	)

	return id, nil
}

func (f *fake) Link(_ context.Context, entityID string, _ Credit) error {
	f.linked = append(f.linked, entityID)

	return nil
}

func (f *fake) OpenReview(_ context.Context, c Credit, candidates []string) error {
	if f.reviews == nil {
		f.reviews = map[string][]string{}
	}

	f.reviews[c.Name] = candidates

	return nil
}

func resolve(t *testing.T, f *fake, c Credit) Outcome {
	t.Helper()

	out, err := Resolve(t.Context(), f, c)
	if err != nil {
		t.Fatal(err)
	}

	return out
}

func TestMatchByQid(t *testing.T) {
	f := &fake{ents: []fakeEntity{{e: Entity{ID: "e1", Kind: "person"}, qid: "Q1", name: "Someone Else"}}}

	if got := resolve(t, f, Credit{Kind: "person", QID: "Q1", Name: "Hayao Miyazaki"}); got != Linked {
		t.Fatalf("outcome = %v, want Linked", got)
	}

	if !slices.Equal(f.linked, []string{"e1"}) || len(f.created) != 0 {
		t.Fatalf("linked %v created %v", f.linked, f.created)
	}
}

func TestQidUnderAnotherKindOpensReview(t *testing.T) {
	f := &fake{ents: []fakeEntity{{e: Entity{ID: "e1", Kind: "organization"}, qid: "Q1"}}}

	if got := resolve(t, f, Credit{Kind: "person", QID: "Q1", Name: "Ghibli"}); got != Reviewed {
		t.Fatalf("outcome = %v, want Reviewed", got)
	}

	if !slices.Equal(f.reviews["Ghibli"], []string{"e1"}) || len(f.linked) != 0 {
		t.Fatalf("reviews %v linked %v", f.reviews, f.linked)
	}
}

func TestMatchByExternalId(t *testing.T) {
	f := &fake{ents: []fakeEntity{
		{e: Entity{ID: "e1", Kind: "person"}, extKey: "tmdb_person", extID: "42", name: "Unrelated"},
		{e: Entity{ID: "e2", Kind: "organization"}, extKey: "tmdb_person", extID: "42"},
	}}

	c := Credit{Kind: "person", ExternalIDKey: "tmdb_person", ExternalID: "42", Name: "Hayao Miyazaki"}
	if got := resolve(t, f, c); got != Linked || !slices.Equal(f.linked, []string{"e1"}) {
		t.Fatalf("outcome %v linked %v", got, f.linked)
	}
}

func TestMatchByNameWithinOneYear(t *testing.T) {
	f := &fake{ents: []fakeEntity{
		{e: Entity{ID: "e1", Kind: "person", Year: 1941}, name: "Hayao Miyazaki"},
		{e: Entity{ID: "e2", Kind: "person", Year: 1970}, name: "Hayao Miyazaki"},
	}}

	if got := resolve(t, f, Credit{Kind: "person", Name: "hayao  MIYAZAKI", Year: 1942}); got != Linked {
		t.Fatalf("outcome = %v, want Linked", got)
	}

	if !slices.Equal(f.linked, []string{"e1"}) {
		t.Fatalf("linked %v", f.linked)
	}
}

func TestAmbiguousOpensReview(t *testing.T) {
	f := &fake{ents: []fakeEntity{
		{e: Entity{ID: "e1", Kind: "person"}, name: "John Smith"},
		{e: Entity{ID: "e2", Kind: "person", Year: 1950}, name: "John Smith"},
	}}

	if got := resolve(t, f, Credit{Kind: "person", Name: "John Smith"}); got != Reviewed {
		t.Fatalf("outcome = %v, want Reviewed", got)
	}

	if !slices.Equal(f.reviews["John Smith"], []string{"e1", "e2"}) || len(f.linked) != 0 || len(f.created) != 0 {
		t.Fatalf("reviews %v linked %v created %v", f.reviews, f.linked, f.created)
	}
}

func TestNoMatchCreatesNative(t *testing.T) {
	f := &fake{}

	c := Credit{Kind: "genre", QID: "Q319221", Name: "Adventure"}
	if got := resolve(t, f, c); got != Created {
		t.Fatalf("outcome = %v, want Created", got)
	}

	if len(f.created) != 1 || !slices.Equal(f.linked, []string{"new-Adventure"}) {
		t.Fatalf("created %v linked %v", f.created, f.linked)
	}

	if got := resolve(t, f, c); got != Linked || len(f.created) != 1 {
		t.Fatalf("second run: outcome %v created %d", got, len(f.created))
	}
}

func TestNamelessCreditWithoutIdsIsSkipped(t *testing.T) {
	f := &fake{}

	if got := resolve(t, f, Credit{Kind: "person", Name: "!!!"}); got != Skipped || len(f.created) != 0 {
		t.Fatalf("outcome %v created %v", got, f.created)
	}
}
