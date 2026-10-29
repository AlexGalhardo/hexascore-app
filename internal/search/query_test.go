package search

import (
	"errors"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name, in        string
		text, match     string
		isbn            string
		empty, tooShort bool
		exclusionOnly   bool
	}{
		{name: "control chars become spaces", in: "ze\x00l\tda\n", text: "ze l da", match: "ze l da"},
		{name: "collapse and trim", in: "  the   last  ", text: "the last", match: "the last"},
		{name: "empty", in: " \x01 ", empty: true},
		{name: "too short", in: "a", text: "a", tooShort: true, exclusionOnly: true},
		{
			name:  "operators removed from match",
			in:    `"the last of us" -part`,
			text:  `"the last of us" -part`,
			match: "the last of us",
		},
		{name: "or dropped", in: "mario or luigi", text: "mario or luigi", match: "mario luigi"},
		{name: "negated phrase", in: `zelda -"breath of"`, text: `zelda -"breath of"`, match: "zelda"},
		{name: "exclusion only", in: "-zelda -mario", text: "-zelda -mario", exclusionOnly: true},
		{name: "isbn plain", in: "9780306406157", text: "9780306406157", match: "9780306406157", isbn: "9780306406157"},
		{
			name:  "isbn hyphens",
			in:    "978-0-306-40615-7",
			text:  "978-0-306-40615-7",
			match: "978-0-306-40615-7",
			isbn:  "9780306406157",
		},
		{name: "not isbn prefix", in: "9770306406157", text: "9770306406157", match: "9770306406157"},
		{name: "nfc", in: "ão", text: "ão", match: "ão"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := Parse(tc.in)
			if err != nil {
				t.Fatal(err)
			}

			if p.Text != tc.text || p.Match != tc.match || p.ISBN13 != tc.isbn ||
				p.Empty != tc.empty || p.TooShort != tc.tooShort || p.ExclusionOnly() != tc.exclusionOnly {
				t.Fatalf("Parse(%q) = %+v", tc.in, p)
			}
		})
	}
}

func TestParseLimits(t *testing.T) {
	if _, err := Parse(strings.Repeat("a", 100)); err != nil {
		t.Fatalf("100 runes rejected: %v", err)
	}

	if _, err := Parse(strings.Repeat("é", 101)); !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("101 runes: %v", err)
	}

	if _, err := Parse("a b c d e f g h i j"); err != nil {
		t.Fatalf("10 terms rejected: %v", err)
	}

	if _, err := Parse("a b c d e f g h i j k"); !errors.Is(err, ErrInvalidQuery) {
		t.Fatalf("11 terms: %v", err)
	}
}

func TestParsePaging(t *testing.T) {
	s := &Postgres{}

	for _, tc := range []struct{ page, per int }{{11, 20}, {5, 50}, {1, 51}, {-1, 20}} {
		_, err := s.Search(t.Context(), Query{Text: "zelda", Audience: AudienceSite, Page: tc.page, PerPage: tc.per})
		if !errors.Is(err, ErrInvalidQuery) {
			t.Errorf("page %d per %d: %v", tc.page, tc.per, err)
		}
	}

	if _, err := s.Search(t.Context(), Query{Text: "zelda"}); !errors.Is(err, ErrInvalidQuery) {
		t.Errorf("missing audience: %v", err)
	}
}

func TestAudienceSQL(t *testing.T) {
	api := itemSQL("", true, false, false, false)
	site := itemSQL("", false, false, true, true)

	if !strings.Contains(api, "search_tsv_api") ||
		strings.Contains(strings.ReplaceAll(api, "search_tsv_api", ""), "search_tsv") {
		t.Error("api audience must rank on search_tsv_api only")
	}

	if strings.Contains(site, "search_tsv_api") || !strings.Contains(site, "search_tsv @@") {
		t.Error("site audience must rank on search_tsv")
	}

	if strings.Contains(entitySQL(true), "default_name FROM entities") {
		t.Error("api entity search must skip the unsourced default name")
	}
}

func TestAudienceSources(t *testing.T) {
	s := &Postgres{apiSources: []string{"wikidata", `we"ird`}}

	if s.sources(AudienceSite) != nil {
		t.Error("site must not filter by source")
	}

	if got := s.sources(AudienceAPI); got != `{"wikidata","we\"ird"}` {
		t.Errorf("api sources = %v", got)
	}

	// Fail closed: an empty allowlist is an empty array, not NULL.
	if got := (&Postgres{}).sources(AudienceAPI); got != "{}" {
		t.Errorf("empty allowlist = %v", got)
	}
}
