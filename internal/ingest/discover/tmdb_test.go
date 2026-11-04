package discover_test

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/discover"
)

const export = `{"adult":false,"id":603,"original_title":"The Matrix","popularity":80.5,"video":false}
{"adult":true,"id":10,"original_title":"X","popularity":50,"video":false}
{"adult":false,"id":7,"original_title":"Stub","popularity":0.2,"video":false}
{"adult":false,"id":12,"original_title":"Finding Nemo","popularity":40,"video":false}
{"adult":false,"id":12,"original_title":"Finding Nemo again","popularity":40,"video":false}
`

func TestParseExportKeepsOrderedPopularNonAdultIds(t *testing.T) {
	ids, err := discover.ParseExport(strings.NewReader(export), discover.MinPopularity)
	if err != nil {
		t.Fatal(err)
	}

	if want := []int{12, 603}; !slices.Equal(ids, want) {
		t.Errorf("ids = %v, want %v (adult, stub and duplicate entries left out, ascending)", ids, want)
	}
}

func TestNextBatchFollowsTheSequence(t *testing.T) {
	ids := []int{1, 5, 9, 20, 30}

	for _, c := range []struct {
		after, n int
		want     []int
	}{
		{0, 2, []int{1, 5}},
		{5, 2, []int{9, 20}},
		{20, 5, []int{30}},
		{30, 5, nil},
	} {
		if got := discover.NextBatch(ids, c.after, c.n); !slices.Equal(got, c.want) {
			t.Errorf("after %d n %d = %v, want %v", c.after, c.n, got, c.want)
		}
	}
}

func TestExportURLIsTheDailyFile(t *testing.T) {
	day := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)

	if got, want := discover.ExportURL("https://files.tmdb.org/p/exports", "movie_ids", day),
		"https://files.tmdb.org/p/exports/movie_ids_11_03_2026.json.gz"; got != want {
		t.Errorf("url = %s, want %s", got, want)
	}
}
