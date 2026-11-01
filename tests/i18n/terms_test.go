package i18n_test

import (
	"encoding/json"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"
)

// flatKeys lists the leaf keys under v as dotted paths.
func flatKeys(prefix string, v any, out map[string]string) {
	switch t := v.(type) {
	case map[string]any:
		for k, child := range t {
			flatKeys(prefix+"."+k, child, out)
		}
	case string:
		out[prefix] = t
	}
}

// TestTermsParityAndScrapingClause: the Terms have the same keys in every locale, and the acceptable-use
// section carries the anti-scraping clause that points bulk data users to the paid API (Q-A18).
func TestTermsParityAndScrapingClause(t *testing.T) {
	t.Parallel()

	var want []string

	for _, locale := range []string{"en", "pt-BR", "es"} {
		raw, err := os.ReadFile("../../resources/lang/" + locale + ".json")
		if err != nil {
			t.Fatal(err)
		}

		var cat map[string]any
		if err := json.Unmarshal(raw, &cat); err != nil {
			t.Fatal(err)
		}

		keys := map[string]string{}
		flatKeys("terms", cat["terms"], keys)

		got := slices.Sorted(maps.Keys(keys))
		if want == nil {
			want = got
		} else if !slices.Equal(got, want) {
			t.Errorf("%s: terms keys differ from en", locale)
		}

		clause := keys["terms.sections.3.body.1"]
		if !strings.Contains(clause, "/docs") || !strings.Contains(clause, "/pricing") {
			t.Errorf("%s: terms.sections.3.body.1 = %q, want the anti-scraping clause linking /docs and /pricing",
				locale, clause)
		}
	}
}
