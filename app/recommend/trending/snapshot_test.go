package trending

import (
	"fmt"
	"slices"
	"testing"
)

func TestSnapshotDeterminism(t *testing.T) {
	var cands []Candidate

	for i := range 40 {
		cands = append(cands, Candidate{
			ID:    fmt.Sprintf("id-%02d", i),
			Trend: float64(40 + i%6*10), Actors: float64(5 + i%3), // many ties on T and on prior
			Prior: float64(i % 4), DaysSinceRelease: float64(i * 9),
		})
	}

	// not eligible: A below 5, so they can only take prior slots
	cands = append(cands, Candidate{ID: "low", Trend: 500, Actors: 4.99, Prior: 80, DaysSinceRelease: 30})

	want, _ := Pick(slices.Clone(cands))

	perms := map[string]func([]Candidate){
		"reverse": slices.Reverse[[]Candidate],
		"rotate": func(c []Candidate) {
			slices.Reverse(c[:7])
			slices.Reverse(c[7:])
			slices.Reverse(c)
		},
		"stride": func(c []Candidate) {
			out := make([]Candidate, len(c))
			for i := range c {
				out[(i*7)%len(c)] = c[i]
			}

			copy(c, out)
		},
	}

	for name, shuffle := range perms {
		shuffled := slices.Clone(cands)
		shuffle(shuffled)

		if got, _ := Pick(shuffled); !slices.Equal(got, want) {
			t.Fatalf("%s: order of input changed the result: %v vs %v", name, got, want)
		}
	}

	if len(want) != ListSize {
		t.Fatalf("got %d entries, want %d", len(want), ListSize)
	}

	for i, e := range want {
		if e.Rank != i+1 || e.Basis != "trend" {
			t.Fatalf("entry %d: %+v", i, e)
		}
	}
}

func TestSnapshotPriorFill(t *testing.T) {
	entries, eligible := Pick([]Candidate{
		{ID: "b", Trend: 60, Actors: 6, Prior: 1},
		{ID: "a", Trend: 60, Actors: 6, Prior: 1},
		{ID: "classic", Prior: 200, DaysSinceRelease: 9000},
		{ID: "new", Prior: 80, DaysSinceRelease: 30},
	})

	if eligible != 2 {
		t.Fatalf("eligible = %d, want 2", eligible)
	}

	want := []Entry{{1, "a", 60, "trend"}, {2, "b", 60, "trend"}}
	if !slices.Equal(entries[:2], want) {
		t.Fatalf("trend entries = %v", entries[:2])
	}

	// SDD 08 §7.3: 80 sitelinks 30 days old gives 6.89, a classic with 200 gives 5.30.
	newer, classic := entries[2], entries[3]
	if newer.ID != "new" || classic.ID != "classic" || newer.Basis != "prior" || !near(newer.Score, 6.89, 0.01) ||
		!near(classic.Score, 5.30, 0.01) {
		t.Fatalf("prior entries = %v", entries[2:])
	}

	if entries[3].Rank != 4 {
		t.Fatalf("ranks are not gapless: %v", entries)
	}
}
