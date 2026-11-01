package editions

import (
	"context"
	"slices"
	"testing"
)

func TestIsbn10To13(t *testing.T) {
	cases := map[string]string{
		"0-306-40615-2": "9780306406157",
		"0306406152":    "9780306406157",
		"080442957X":    "9780804429573",
		"080442957x":    "9780804429573",
	}

	for in, want := range cases {
		got, ok := ToISBN13(in)
		if !ok || got != want {
			t.Errorf("ToISBN13(%q) = %q, %v, want %q", in, got, ok, want)
		}
	}

	for _, bad := range []string{"0306406153", "080442957Y", "12345", "", "030640615"} {
		if got, ok := ToISBN13(bad); ok {
			t.Errorf("ToISBN13(%q) = %q, want rejected", bad, got)
		}
	}
}

func TestIsbn13Checksum(t *testing.T) {
	if got, ok := ToISBN13("978-0-306-40615-7"); !ok || got != "9780306406157" {
		t.Errorf("valid 13 = %q, %v", got, ok)
	}

	for _, bad := range []string{"9780306406158", "9770306406157", "978030640615X"} {
		if got, ok := ToISBN13(bad); ok {
			t.Errorf("ToISBN13(%q) = %q, want rejected", bad, got)
		}
	}
}

func TestSelectCapsOrdersAndCountsDropped(t *testing.T) {
	var in []Candidate

	in = append(in, Candidate{ISBN: "9780306406158", Language: "pt"}, Candidate{ISBN: "9780306406157", Language: "fr"})
	in = append(in, Candidate{ISBN: "0306406152", Language: "pt"}) // same book as the second, as ISBN-10
	in = append(in, Candidate{QID: "Q1", Language: "es"}, Candidate{QID: "Q2", Language: "en"})

	for i := range 60 {
		in = append(in, Candidate{QID: "Q" + string(rune('A'+i%26)) + string(rune('a'+i/26)), Language: "de"})
	}

	got, dropped := Select(in, 50)

	if dropped != 1 {
		t.Errorf("dropped = %d, want 1 (the bad checksum)", dropped)
	}

	if len(got) != 50 {
		t.Fatalf("kept %d, want 50", len(got))
	}

	var langs []string
	for _, c := range got[:3] {
		langs = append(langs, c.Language)
	}

	if !slices.Equal(langs, []string{"pt", "en", "es"}) {
		t.Errorf("first languages = %v, want pt, en, es", langs)
	}

	isbns := 0

	for _, c := range got {
		if c.ISBN == "9780306406157" {
			isbns++
		}
	}

	if isbns != 1 {
		t.Errorf("duplicate ISBN kept %d times", isbns)
	}
}

func TestMarcLanguage(t *testing.T) {
	for in, want := range map[string]string{"por": "pt", "eng": "en", "spa": "es", "fre": "und", "": "und"} {
		if got := Language(in); got != want {
			t.Errorf("Language(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPrimaryEdition(t *testing.T) {
	in := []Candidate{
		{QID: "Q1", Language: "pt", Year: 1990},
		{QID: "Q2", Language: "en", Year: 1980, HasCover: true},
		{QID: "Q3", Language: "en", Year: 1985, HasCover: true},
		{QID: "Q4", Language: "en", Year: 1970},
	}

	if got := PrimaryEdition(in, "en"); got != 1 {
		t.Errorf("primary = %d, want 1 (earliest with a cover in the original language)", got)
	}

	if got := PrimaryEdition(in, "fr"); got != -1 {
		t.Errorf("primary without a match = %d, want -1", got)
	}
}

type fakeStore struct {
	owner map[string]string
	saved map[string][]string
}

func (f *fakeStore) WorkOfISBN(_ context.Context, isbn string) (string, bool, error) {
	w, ok := f.owner[isbn]

	return w, ok, nil
}

func (f *fakeStore) Upsert(_ context.Context, work string, c Candidate) error {
	if f.saved == nil {
		f.saved = map[string][]string{}
	}

	f.saved[work] = append(f.saved[work], c.ISBN)

	return nil
}

func TestIsbnOnAnotherWorkIsNotMoved(t *testing.T) {
	f := &fakeStore{owner: map[string]string{"9780306406157": "work-A"}}

	out, err := Save(t.Context(), f, "work-B", Candidate{ISBN: "9780306406157"})
	if err != nil || out != Conflict {
		t.Fatalf("outcome %v, err %v, want Conflict", out, err)
	}

	if len(f.saved) != 0 {
		t.Fatalf("saved %v, want nothing", f.saved)
	}

	if out, _ := Save(t.Context(), f, "work-A", Candidate{ISBN: "9780306406157"}); out != Saved {
		t.Fatalf("same work = %v, want Saved", out)
	}

	if out, _ := Save(t.Context(), f, "work-B", Candidate{QID: "Q9"}); out != Saved {
		t.Fatalf("no ISBN = %v, want Saved", out)
	}
}
