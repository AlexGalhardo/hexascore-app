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
