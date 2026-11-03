package tmdb_test

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/offers"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/tmdb"
)

func TestParseProviders(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile("testdata/providers.json")
	if err != nil {
		t.Fatal(err)
	}

	got, err := tmdb.ParseProviders(raw, offers.Countries)
	if err != nil {
		t.Fatal(err)
	}

	var lines []string
	for _, p := range got {
		lines = append(lines, strings.Join([]string{p.Country, p.Kind, p.Name}, "/"))
	}

	slices.Sort(lines)

	want := []string{
		"BR/buy/Apple TV",
		"BR/rent/Apple TV",
		"BR/stream/Amazon Prime Video", // the name is trimmed
		"BR/stream/Netflix",
		"MX/stream/Disney Plus", // entries without an id or a name are dropped
		"US/free/Pluto TV",      // "ads" and "free" are both free
		"US/free/Tubi TV",
		// DE is not a configured country; ES has an http link; PT links outside themoviedb.org
	}

	if !slices.Equal(lines, want) {
		t.Fatalf("providers =\n%s\nwant\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}

	for _, p := range got {
		if !strings.HasPrefix(p.Link, "https://www.themoviedb.org/") {
			t.Fatalf("link %q is not a TMDB watch page", p.Link)
		}
	}
}

func TestParseProvidersRejectsBadJSON(t *testing.T) {
	t.Parallel()

	if _, err := tmdb.ParseProviders([]byte(`{"results": [`), offers.Countries); err == nil {
		t.Fatal("malformed JSON was accepted")
	}

	got, err := tmdb.ParseProviders([]byte(`{"id": 1}`), offers.Countries)
	if err != nil || len(got) != 0 {
		t.Fatalf("a title nobody streams: %v, %v", got, err)
	}
}
