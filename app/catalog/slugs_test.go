package catalog_test

import (
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

func TestSlugify(t *testing.T) {
	t.Parallel()

	for in, want := range map[string]string{
		"Half-Life 2":             "half-life-2",
		"Sénha: Ação & Fúria!":    "senha-acao-furia",
		"ﬁx  it":                  "fix-it",
		"  --Dom Casmurro--  ":    "dom-casmurro",
		"The Legend of Zelda™":    "the-legend-of-zeldatm",
		"日本語":                     "",
		strings.Repeat("ab ", 80): strings.TrimRight(strings.Repeat("ab-", 40), "-"),
	} {
		if got := catalog.Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}

	if got := catalog.Slugify(strings.Repeat("a", 200)); len(got) != catalog.MaxSlugLength {
		t.Errorf("long slug length %d, want %d", len(got), catalog.MaxSlugLength)
	}
}

func TestFirstFreeSlug(t *testing.T) {
	t.Parallel()

	taken := map[string]bool{"dune": true, "dune-2": true}
	isTaken := func(s string) bool { return taken[s] }

	for base, want := range map[string]string{"dune": "dune-3", "arrakis": "arrakis", "search": "search-2"} {
		if got := catalog.FirstFreeSlug(base, isTaken); got != want {
			t.Errorf("FirstFreeSlug(%q) = %q, want %q", base, got, want)
		}
	}
}
