package gog_test

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/gog"
)

var _ ingest.Connector = (*gog.Connector)(nil)

func fixture(t *testing.T) []byte {
	t.Helper()

	raw, err := os.ReadFile("../testdata/gog_catalog.json")
	if err != nil {
		t.Fatal(err)
	}

	return raw
}

func TestGogPicksProductBySlugPrefix(t *testing.T) {
	p, err := gog.ParseCatalog(fixture(t), "game/the_witcher_3_wild_hunt", "The Witcher 3: Wild Hunt", "Q4267401")
	if err != nil || p == nil {
		t.Fatalf("ParseCatalog = %v, %v", p, err)
	}

	if p.Slug != "the_witcher_3_wild_hunt_game_of_the_year_edition" {
		t.Errorf("slug = %q", p.Slug)
	}

	if !strings.HasPrefix(p.StoreURL, "https://www.gog.com/") {
		t.Errorf("store_url = %q", p.StoreURL)
	}

	if !slices.Equal(p.Platforms, []string{"Windows"}) {
		t.Errorf("platforms = %v", p.Platforms)
	}
}

func TestGogNoMatch(t *testing.T) {
	p, err := gog.ParseCatalog(fixture(t), "game/portal", "Portal", "Q1")
	if err != nil || p != nil {
		t.Errorf("ParseCatalog = %v, %v; want no match", p, err)
	}
}
