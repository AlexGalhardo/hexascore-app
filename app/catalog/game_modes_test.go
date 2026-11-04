package catalog

import (
	"strings"
	"testing"
	"unicode"
)

func TestGameModeForSimplifiesTheWikidataLabels(t *testing.T) {
	for label, want := range map[string]string{
		"single-player video game":          "single-player",
		"multiplayer video game":            "multiplayer",
		"co-op mode":                        "co-op",
		"player versus player":              "pvp",
		"multiplayer online game":           "online",
		"shared/splitscreen multiplayer":    "split-screen",
		"massively multiplayer online game": "mmo",
		"local multiplayer":                 "split-screen",
		"Cooperative gameplay":              "co-op",
	} {
		if got, ok := GameModeFor("game", label); !ok || got != want {
			t.Errorf("%q = %q (%v), want %q", label, got, ok, want)
		}
	}

	if got, ok := GameModeFor("game", "asymmetric whatever"); ok {
		t.Errorf("an unknown label resolved to %q", got)
	}

	if got, ok := GameModeFor("movie", "single-player video game"); ok {
		t.Errorf("a mode is for games only, got %q", got)
	}
}

// Every canonical mode has its three names, each starting with a capital letter (owner 2026-11-04).
func TestEveryGameModeHasCapitalisedNames(t *testing.T) {
	for code, n := range GameModes {
		for locale, name := range map[string]string{"en": n.En, "pt-BR": n.PtBR, "es": n.Es} {
			if strings.TrimSpace(name) == "" || !unicode.IsUpper([]rune(name)[0]) {
				t.Errorf("%s %s = %q: empty or not capitalised", code, locale, name)
			}
		}
	}

	for _, r := range gameModeRules {
		if _, ok := GameModes[r.code]; !ok {
			t.Errorf("rule for %q has no canonical mode", r.code)
		}
	}
}
