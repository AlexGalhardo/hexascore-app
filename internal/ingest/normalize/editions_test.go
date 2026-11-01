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
