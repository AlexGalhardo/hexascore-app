package i18n_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

type pluralCase struct {
	Locale string `json:"locale"`
	Key    string `json:"key"`
	Count  int    `json:"count"`
}

// TestPluralGolden renders every plural catalogue key in every locale with
// counts that hit each CLDR category, in Go (i18n.T) and in TypeScript
// (resources/js/lib/i18n.ts via bun), and requires identical strings.
func TestPluralGolden(t *testing.T) {
	bun, err := exec.LookPath("bun")
	if err != nil {
		t.Skip("bun not installed")
	}

	app := vukemtest.New(t, bootstrap.New)
	keys := []string{
		"api.profile.plan.rate", "billing.checkout.closeIn", "catalog.resultCount", "catalog.ratingCount",
		"item.timeToBeat.reports", "lists.itemCount", "social.profile.counts.game",
		"social.profile.counts.movie", "social.profile.counts.series", "social.profile.counts.book",
	}
	var cases []pluralCase
	for _, locale := range []string{"en", "pt-BR", "es"} {
		for _, key := range keys {
			for _, n := range []int{0, 1, 2, 5, 21, 1000000} {
				cases = append(cases, pluralCase{locale, key, n})
			}
		}
	}

	in, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), bun, "packages/i18n/src/golden.ts")
	cmd.Dir, cmd.Stdin, cmd.Stderr = root, bytes.NewReader(in), os.Stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("bun golden.ts: %v", err)
	}
	var ts []string
	if err := json.Unmarshal(out, &ts); err != nil {
		t.Fatalf("decode bun output: %v", err)
	}
	if len(ts) != len(cases) {
		t.Fatalf("bun returned %d results for %d cases", len(ts), len(cases))
	}

	for i, c := range cases {
		goOut := app.I18n.T(c.Locale, c.Key, map[string]any{"count": c.Count})
		if goOut != ts[i] {
			t.Errorf("%s %s count=%d: go %q, ts %q", c.Locale, c.Key, c.Count, goOut, ts[i])
		}
		if goOut == c.Key {
			t.Errorf("%s %s: key not found", c.Locale, c.Key)
		}
	}
}
