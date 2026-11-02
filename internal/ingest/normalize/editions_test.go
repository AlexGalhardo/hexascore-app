package normalize_test

import (
	"encoding/json"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

func claims(values ...string) []map[string]any {
	out := make([]map[string]any, len(values))
	for i, v := range values {
		out[i] = map[string]any{"value": v, "rank": "normal"}
	}

	return out
}

func workRecord(t *testing.T, typ string, c map[string][]map[string]any) store.Record {
	t.Helper()

	raw, err := json.Marshal(map[string]any{
		"type": typ, "id": "Q9", "labels": map[string]string{"en": "A Book"}, "claims": c,
	})
	if err != nil {
		t.Fatal(err)
	}

	return store.Record{Source: "wikidata", ExternalID: "Q9", Payload: raw}
}

func TestExtractEditionsFromIsbnAndEditionQids(t *testing.T) {
	rec := workRecord(t, "book", map[string][]map[string]any{
		"P212": claims("978-0-451-52493-5", "9780451524930"),
		"P957": claims("0-451-52493-4", "0306406152"),
		"P747": claims("Q100", "Q101"),
	})

	got, dropped, err := normalize.ExtractEditions("Q9", []store.Record{rec})
	if err != nil {
		t.Fatal(err)
	}

	if dropped != 1 {
		t.Errorf("dropped = %d, want 1 (the bad 13-digit ISBN)", dropped)
	}

	isbns := map[string]bool{}
	qids := 0

	for _, c := range got {
		isbns[c.ISBN] = true

		if c.QID != "" {
			qids++
		}

		if c.Source != "wikidata" || c.Title != "A Book" {
			t.Errorf("candidate = %+v, want source wikidata and the work title", c)
		}
	}

	// P212 and the ISBN-10 of the same book (P957) collapse into one ISBN-13.
	if !isbns["9780451524935"] || !isbns["9780306406157"] || len(isbns) != 3 || qids != 2 {
		t.Errorf("isbns = %v, qid candidates = %d, want 2 ISBNs plus 2 QID-only editions", isbns, qids)
	}
}

func TestExtractEditionsOnlyForBooks(t *testing.T) {
	rec := workRecord(t, "movie", map[string][]map[string]any{"P212": claims("9780451524935")})

	got, _, err := normalize.ExtractEditions("Q9", []store.Record{rec})
	if err != nil || len(got) != 0 {
		t.Errorf("movie editions = %v, %v, want none", got, err)
	}
}

func editionRecord(t *testing.T, qid string, c map[string]any) store.Record {
	t.Helper()

	raw, err := json.Marshal(map[string]any{"id": qid, "labels": map[string]string{"en": "Ed " + qid}, "claims": c})
	if err != nil {
		t.Fatal(err)
	}

	return store.Record{Source: "wikidata:entity", ExternalID: qid, Payload: raw}
}

// SDD 06 §7.4, ADR 0012: a P747 edition takes its language (P407), year (P577) and ISBN from its own entity
// record; the same ISBN on the work does not make a second candidate.
func TestExtractEditionsReadsEditionEntities(t *testing.T) {
	year := func(y string) []map[string]any {
		return []map[string]any{
			{"value": map[string]any{"time": "+" + y + "-00-00T00:00:00Z", "precision": 9}, "rank": "normal"},
		}
	}
	work := workRecord(t, "book", map[string][]map[string]any{
		"P212": claims("9780451524935"),
		"P747": claims("Q100", "Q101", "Q102"),
	})
	recs := []store.Record{
		work,
		editionRecord(
			t,
			"Q100",
			map[string]any{"P407": claims("Q1860"), "P577": year("1950"), "P212": claims("9780451524935")},
		),
		editionRecord(t, "Q101", map[string]any{"P407": claims("Q750553"), "P577": year("2009")}),
		editionRecord(t, "Q102", map[string]any{"P407": claims("Q188"), "P957": claims("0306406152")}),
	}

	got, _, err := normalize.ExtractEditions("Q9", recs)
	if err != nil {
		t.Fatal(err)
	}

	byQID := map[string]int{}
	for i, c := range got {
		byQID[c.QID] = i
	}

	if len(got) != 3 {
		t.Fatalf("candidates = %+v, want the three editions and no ISBN-only duplicate", got)
	}

	if c := got[byQID["Q100"]]; c.Language != "en" || c.Year != 1950 || c.ISBN != "9780451524935" {
		t.Errorf("Q100 = %+v, want en, 1950 and the ISBN", c)
	}

	if c := got[0]; c.QID != "Q101" || c.Language != "pt" || c.Year != 2009 {
		t.Errorf("first = %+v, want Q101 in pt (Brazilian Portuguese), 2009", c)
	}

	if c := got[byQID["Q102"]]; c.Language != "und" || c.ISBN != "9780306406157" {
		t.Errorf("Q102 = %+v, want und and the ISBN-10 converted", c)
	}
}

// ADR 0012: an edition's publisher is its first P123; the name comes from the work's referenced labels.
func TestExtractEditionsCarriesThePublisher(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"type": "book", "id": "Q9", "labels": map[string]string{"en": "A Book"},
		"claims": map[string]any{"P747": claims("Q100", "Q101")},
		"refs":   map[string]string{"Q55": "Penguin Books"},
	})
	if err != nil {
		t.Fatal(err)
	}

	recs := []store.Record{
		{Source: "wikidata", ExternalID: "Q9", Payload: raw},
		editionRecord(t, "Q100", map[string]any{"P123": claims("Q55")}),
		editionRecord(t, "Q101", map[string]any{"P123": claims("Q56")}),
	}

	got, _, err := normalize.ExtractEditions("Q9", recs)
	if err != nil {
		t.Fatal(err)
	}

	pubs := map[string][2]string{}
	for _, c := range got {
		pubs[c.QID] = [2]string{c.PublisherQID, c.PublisherName}
	}

	if pubs["Q100"] != [2]string{"Q55", "Penguin Books"} || pubs["Q101"] != [2]string{"Q56", ""} {
		t.Errorf("publishers = %v, want Q100 → Q55 Penguin Books, Q101 → Q56 without a name", pubs)
	}
}

func openLibraryRecord(t *testing.T, p map[string]any) store.Record {
	t.Helper()

	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}

	return store.Record{Source: "openlibrary", ExternalID: "Q9", Payload: raw}
}

// SDD 06 §7.4 (Open Library editions after A-05, cleared by Q-A16): the Open Library edition fills the language
// and year Wikidata lacks for the same ISBN and adds its id; Wikidata values win.
func TestExtractEditionsFillsFromOpenLibrary(t *testing.T) {
	work := workRecord(t, "book", map[string][]map[string]any{"P212": claims("9780451524935")})
	ol := openLibraryRecord(t, map[string]any{
		"qid": "Q9", "key": "/books/OL62592264M", "isbn": "0451524934", "language": "por", "year": 1984,
	})

	got, _, err := normalize.ExtractEditions("Q9", []store.Record{work, ol})
	if err != nil {
		t.Fatal(err)
	}

	if len(got) != 1 {
		t.Fatalf("candidates = %+v, want one merged edition", got)
	}

	if c := got[0]; c.OpenLibraryID != "OL62592264M" || c.Language != "pt" || c.Year != 1984 || c.Source != "wikidata" {
		t.Errorf("edition = %+v, want the Open Library id, pt, 1984, source wikidata", c)
	}
}

func TestExtractEditionsAddsAnOpenLibraryOnlyEdition(t *testing.T) {
	work := workRecord(t, "book", map[string][]map[string]any{"P212": claims("9780451524935")})
	ol := openLibraryRecord(t, map[string]any{
		"qid": "Q9", "key": "/books/OL1M", "isbn": "9780306406157", "language": "eng", "year": 2001,
	})

	got, _, err := normalize.ExtractEditions("Q9", []store.Record{work, ol})
	if err != nil {
		t.Fatal(err)
	}

	var found bool

	for _, c := range got {
		if c.OpenLibraryID == "OL1M" {
			found = c.ISBN == "9780306406157" && c.Language == "en" && c.Year == 2001 && c.Source == "openlibrary"
		}
	}

	if len(got) != 2 || !found {
		t.Errorf("candidates = %+v, want the Wikidata ISBN plus the Open Library edition (en, 2001)", got)
	}

	work2 := workRecord(t, "book", nil)
	olWork := openLibraryRecord(t, map[string]any{"qid": "Q9", "key": "/works/OL1W"})

	if got, _, _ := normalize.ExtractEditions("Q9", []store.Record{work2, olWork}); len(got) != 0 {
		t.Errorf("work document candidates = %+v, want none (only edition documents are editions)", got)
	}
}

func TestOriginalLanguage(t *testing.T) {
	for lang, want := range map[string]string{"Q1860": "en", "Q5146": "pt", "Q150": ""} {
		rec := workRecord(t, "book", map[string][]map[string]any{"P407": claims(lang)})
		if got := normalize.OriginalLanguage("Q9", []store.Record{rec}); got != want {
			t.Errorf("OriginalLanguage(%s) = %q, want %q", lang, got, want)
		}
	}
}
