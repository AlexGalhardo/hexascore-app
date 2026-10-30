package community

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// now is the clock the table freezes so "now + 10 y" is 2036-10-30.
var validateNow = time.Date(2026, 10, 30, 15, 0, 0, 0, time.UTC)

func TestValidate(t *testing.T) {
	t.Parallel()

	longSynopsis := strings.Repeat("a", 50)
	maxSynopsis := strings.Repeat("b", 2000)
	wikidata := "https://www.wikidata.org/wiki/Q127367"

	tests := []struct {
		name    string
		in      SuggestionInput
		wantErr bool
	}{
		{
			name: "title accepted",
			in:   edit(change("title", nil, "O Senhor dos Anéis"), wikidata),
		},
		{
			name:    "title empty",
			in:      edit(change("title", nil, ""), wikidata),
			wantErr: true,
		},
		{
			name:    "title over 300",
			in:      edit(change("title", nil, strings.Repeat("á", 301)), wikidata),
			wantErr: true,
		},
		{
			name:    "title control char",
			in:      edit(change("title", nil, "Hello\n"), wikidata),
			wantErr: true,
		},
		{
			name: "translation title accepted",
			in:   edit(change("translations.pt-BR.title", nil, "Duna"), wikidata),
		},
		{
			name:    "translation unknown locale",
			in:      edit(change("translations.fr.title", nil, "Dune"), wikidata),
			wantErr: true,
		},
		{
			name: "synopsis bounds",
			in:   edit(change("synopsis.en", nil, longSynopsis), wikidata),
		},
		{
			name: "synopsis max",
			in:   edit(change("synopsis.pt-BR", nil, maxSynopsis), wikidata),
		},
		{
			name:    "synopsis too short",
			in:      edit(change("synopsis.en", nil, strings.Repeat("a", 49)), wikidata),
			wantErr: true,
		},
		{
			name:    "synopsis too long",
			in:      edit(change("synopsis.es", nil, strings.Repeat("a", 2001)), wikidata),
			wantErr: true,
		},
		{
			name: "release date floor",
			in:   edit(dateChange("1870-01-01", "day"), wikidata),
		},
		{
			name: "release date ceiling",
			in:   edit(dateChange("2036-10-30", "day"), wikidata),
		},
		{
			name:    "release date before 1870",
			in:      edit(dateChange("1869-12-31", "day"), wikidata),
			wantErr: true,
		},
		{
			name:    "release date after now plus 10 years",
			in:      edit(dateChange("2036-10-31", "year"), wikidata),
			wantErr: true,
		},
		{
			name:    "release date bad precision",
			in:      edit(dateChange("2001-12-19", "hour"), wikidata),
			wantErr: true,
		},
		{
			name: "wikidata qid",
			in:   edit(change("external_ids.wikidata", nil, "Q127367"), wikidata),
		},
		{
			name:    "wikidata not a qid",
			in:      edit(change("external_ids.wikidata", nil, "127367"), wikidata),
			wantErr: true,
		},
		{
			name: "tmdb digits",
			in:   edit(change("external_ids.tmdb", nil, "550"), wikidata),
		},
		{
			name:    "tmdb not digits",
			in:      edit(change("external_ids.tmdb", nil, "tt550"), wikidata),
			wantErr: true,
		},
		{
			name: "igdb digits",
			in:   edit(change("external_ids.igdb", nil, "1020"), wikidata),
		},
		{
			name: "open library work",
			in:   edit(change("external_ids.openlibrary", nil, "OL45883W"), wikidata),
		},
		{
			name:    "open library missing suffix",
			in:      edit(change("external_ids.openlibrary", nil, "OL45883"), wikidata),
			wantErr: true,
		},
		{
			name: "isbn13 checksum",
			in:   edit(change("external_ids.isbn13", nil, "9780306406157"), wikidata),
		},
		{
			name:    "isbn13 bad checksum",
			in:      edit(change("external_ids.isbn13", nil, "9780306406158"), wikidata),
			wantErr: true,
		},
		{
			name: "twenty relation ops",
			in:   edit(relations("genres", 10, 10), wikidata),
		},
		{
			name:    "twenty one relation ops",
			in:      edit(relations("creators", 11, 10), wikidata),
			wantErr: true,
		},
		{
			name:    "unknown creator role",
			in:      edit(creator("not-a-role"), wikidata),
			wantErr: true,
		},
		{
			name: "three source urls",
			in: edit(change("title", nil, "Dune"),
				wikidata,
				"https://en.wikipedia.org/wiki/Dune",
				"https://openlibrary.org/works/OL45883W",
			),
		},
		{
			name: "four source urls",
			in: edit(change("title", nil, "Dune"),
				wikidata,
				"https://en.wikipedia.org/wiki/Dune",
				"https://openlibrary.org/works/OL45883W",
				"https://example.com/dune",
			),
			wantErr: true,
		},
		{
			name:    "hexascore.app refused",
			in:      edit(change("title", nil, "Dune"), "https://hexascore.app/item/dune"),
			wantErr: true,
		},
		{
			name:    "hexascore subdomain refused",
			in:      edit(change("title", nil, "Dune"), "https://www.hexascore.app/item/dune"),
			wantErr: true,
		},
		{
			name:    "shortener refused",
			in:      edit(change("title", nil, "Dune"), "https://bit.ly/dune"),
			wantErr: true,
		},
		{
			name:    "http source refused",
			in:      edit(change("title", nil, "Dune"), "http://example.com/dune"),
			wantErr: true,
		},
		{
			name:    "factual field without source",
			in:      edit(change("title", nil, "Dune")),
			wantErr: true,
		},
		{
			name:    "empty effective change",
			in:      edit(change("title", "Dune", "Dune"), wikidata),
			wantErr: true,
		},
		{
			name:    "nfc equivalent is empty",
			in:      edit(change("title", "Am\u00e9lie", "Ame\u0301lie"), wikidata),
			wantErr: true,
		},
		{
			name: "nfc stored form differs from raw bytes",
			in:   edit(change("title", nil, "Ame\u0301lie"), wikidata),
		},
		{
			name:    "unknown field",
			in:      edit(change("runtime_minutes", nil, "120"), wikidata),
			wantErr: true,
		},
		{
			name: "cover https optional rules",
			in:   edit(change("cover", nil, "https://upload.wikimedia.org/cover.jpg"), wikidata),
		},
		{
			name:    "cover http refused",
			in:      edit(change("cover", nil, "http://upload.wikimedia.org/cover.jpg"), wikidata),
			wantErr: true,
		},
		{
			name:    "note over 1000",
			in:      withNote(edit(change("title", nil, "Dune"), wikidata), strings.Repeat("n", 1001)),
			wantErr: true,
		},
		{
			name:    "create without title",
			in:      SuggestionInput{Kind: "create", TargetType: "movie", Locale: "en", Changes: map[string]FieldChange{}},
			wantErr: true,
		},
		{
			name: "create with title",
			in: SuggestionInput{
				Kind: "create", TargetType: "movie", Locale: "en",
				SourceURLs: []string{wikidata},
				Changes:    map[string]FieldChange{"title": change("title", nil, "Dune")["title"]},
			},
		},
		{
			name:    "merge different type",
			in:      SuggestionInput{Kind: "merge", TargetType: "movie", OtherType: "book", Locale: "en"},
			wantErr: true,
		},
		{
			name: "merge same type",
			in: SuggestionInput{
				Kind: "merge", TargetType: "movie", OtherType: "movie", Locale: "en",
				Changes: map[string]FieldChange{"*merge*": change("*merge*", "src", "dst")["*merge*"]},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Validate(tt.in, validateNow)
			if tt.wantErr {
				var ve *ValidationError
				if !errors.As(err, &ve) || ve.Status != 422 {
					t.Fatalf("Validate() err = %v, want 422 ValidationError", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Validate() err = %v", err)
			}
		})
	}

	if !utf8.ValidString(longSynopsis) {
		t.Fatal("fixture")
	}
}

func edit(changes map[string]FieldChange, urls ...string) SuggestionInput {
	return SuggestionInput{
		Kind:       "edit",
		TargetType: "movie",
		Locale:     "pt-BR",
		SourceURLs: urls,
		Changes:    changes,
	}
}

func withNote(in SuggestionInput, note string) SuggestionInput {
	in.Note = note
	return in
}

func change(field string, before, after any) map[string]FieldChange {
	return map[string]FieldChange{
		field: {Before: mustJSON(before), After: mustJSON(after)},
	}
}

func dateChange(value, precision string) map[string]FieldChange {
	dated := map[string]string{"value": value, "precision": precision}
	return map[string]FieldChange{
		"release_date": {Before: mustJSON(nil), After: mustJSON(dated)},
	}
}

func relations(field string, addN, removeN int) map[string]FieldChange {
	return map[string]FieldChange{
		field: {Add: ops(field, addN), Remove: ops(field, removeN)},
	}
}

func ops(field string, n int) []RelationOp {
	out := make([]RelationOp, n)
	for i := range out {
		out[i] = RelationOp{ID: field + "-" + string(rune('a'+i%26)) + string(rune('0'+i/26))}
		if field == "creators" || field == "organizations" {
			out[i].Role = "director"
			if field == "organizations" {
				out[i].Role = "publisher"
			}
		}
	}
	return out
}

func creator(role string) map[string]FieldChange {
	return map[string]FieldChange{
		"creators": {Add: []RelationOp{{ID: "ent-1", Role: role}}},
	}
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return b
}
