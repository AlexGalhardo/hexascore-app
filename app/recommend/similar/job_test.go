package similar

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestSimilarFilterKeepRule(t *testing.T) {
	w := func(f Feature) float64 {
		return map[string]float64{"franchise": 5, "genre": 1, "cast": 1, "other": 40}[f.Role]
	}
	self := map[Feature]float64{{"franchise", "f"}: 1, {"other", "o"}: 1}

	cands := []Candidate{
		// sim 5/46 = 0.109: below 0.15 but shares a franchise (omega 5): kept
		{ID: "strong-role", Features: map[Feature]float64{{"franchise", "f"}: 1}},
		// shares nothing: dropped
		{ID: "none", Features: map[Feature]float64{{"genre", "g"}: 1}},
	}
	got := Filter(self, cands, w)
	if len(got) != 1 || got[0].ID != "strong-role" || got[0].Sim != 0.11 {
		t.Fatalf("got %+v", got)
	}

	// weak role only (genre, omega 1): sim 1/(1+1+... ) below 0.15 dropped, above kept
	self2 := map[Feature]float64{
		{"genre", "g"}: 1, {"cast", "a"}: 1, {"cast", "b"}: 1, {"cast", "c"}: 1, {"cast", "d"}: 1,
		{"cast", "e"}: 1, {"cast", "f"}: 1, {"cast", "g"}: 1, {"cast", "h"}: 1, {"cast", "i"}: 1,
	}
	low := Candidate{ID: "low", Features: map[Feature]float64{{"genre", "g"}: 1}} // 1/10 = 0.10 < 0.15, no strong role
	if got := Filter(self2, []Candidate{low}, w); len(got) != 0 {
		t.Fatalf("weak-role 0.10 must be dropped, got %+v", got)
	}

	hi := Candidate{ID: "hi", Features: map[Feature]float64{{"genre", "g"}: 1, {"cast", "a"}: 1}} // 2/10 = 0.20
	if got := Filter(self2, []Candidate{hi}, w); len(got) != 1 {
		t.Fatalf("0.20 must be kept, got %+v", got)
	}

	tiny := Candidate{ID: "tiny", Features: map[Feature]float64{{"genre", "g"}: 1}}
	self3 := map[Feature]float64{{"genre", "g"}: 1}
	for i := range 30 {
		self3[Feature{"cast", strconv.Itoa(i)}] = 1
	}
	if got := Filter(self3, []Candidate{tiny}, w); len(got) != 0 { // 1/31 < 0.05
		t.Fatalf("below 0.05 must be dropped, got %+v", got)
	}
}

func TestSimilarOrder(t *testing.T) {
	in := []Scored{
		{ID: "d", Sim: 0.5, Cooccur: 1, Popularity: 9},
		{ID: "c", Sim: 0.5, Cooccur: 3, Popularity: 1},
		{ID: "b", Sim: 0.6},
		{ID: "a", Sim: 0.5, Cooccur: 1, Popularity: 9},
		{ID: "e", Sim: 0.5, Cooccur: 1, Popularity: 10},
	}
	var ids strings.Builder
	for _, s := range Order(in) {
		ids.WriteString(s.ID)
	}
	if ids.String() != "bcead" {
		t.Fatalf("order = %s, want bcead", ids.String())
	}

	many := make([]Scored, 50)
	for i := range many {
		many[i] = Scored{ID: fmt.Sprintf("%02d", i), Sim: 0.3}
	}
	if got := Order(many); len(got) != MaxNeighbours || got[0].ID != "00" {
		t.Fatalf("cap/tie-break: %d %s", len(got), got[0].ID)
	}
}
