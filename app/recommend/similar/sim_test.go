package similar

import (
	"math"
	"testing"
)

func TestSimilarityZeldaExample(t *testing.T) {
	// SDD 08 section 7.4: franchise 16.0, developer 12.3 and genre 0.9 are shared, 12.3 of other features are not.
	g := map[Feature]float64{
		{"franchise", "zelda"}:        16.0,
		{"developer", "nintendo-epd"}: 12.3,
		{"genre", "adventure"}:        0.9,
		{"cast", "a"}:                 6.0,
		{"cast", "b"}:                 6.3,
	}
	weight := func(f Feature) float64 { return g[f] }

	a := map[Feature]float64{
		{"franchise", "zelda"}: 1, {"developer", "nintendo-epd"}: 1, {"genre", "adventure"}: 1, {"cast", "a"}: 1,
	}
	b := map[Feature]float64{
		{"franchise", "zelda"}: 1, {"developer", "nintendo-epd"}: 1, {"genre", "adventure"}: 1, {"cast", "b"}: 1,
	}

	// shared 29.2, union 29.2 + 6.0 + 6.3 = 41.5
	got := Sim(a, b, weight)
	if math.Round(got*100)/100 != 0.70 {
		t.Fatalf("sim = %v, want 0.70", got)
	}
	if rev := Sim(b, a, weight); rev != got {
		t.Fatalf("not symmetric: %v vs %v", got, rev)
	}
}

func TestSimilarityBounds(t *testing.T) {
	w := func(Feature) float64 { return 2 }
	f := Feature{"genre", "rpg"}
	same := map[Feature]float64{f: 1}
	if got := Sim(same, same, w); got != 1 {
		t.Fatalf("identical = %v, want 1", got)
	}
	if got := Sim(same, map[Feature]float64{{"genre", "fps"}: 1}, w); got != 0 {
		t.Fatalf("disjoint = %v, want 0", got)
	}
	if got := Sim(nil, nil, w); got != 0 {
		t.Fatalf("empty = %v, want 0", got)
	}
}

func TestSimilarityParentHalfWeight(t *testing.T) {
	w := func(Feature) float64 { return 4 }
	f := Feature{"franchise", "zelda"}
	// i holds the franchise directly, j only through its parent (half weight): min 0.5 over max 1.
	got := Sim(map[Feature]float64{f: 1}, map[Feature]float64{f: 0.5}, w)
	if got != 0.5 {
		t.Fatalf("sim = %v, want 0.5", got)
	}
}

func TestSimilarityWeight(t *testing.T) {
	// ln(1000/(1+9)) = ln 100
	if got, want := Weight("franchise", 1000, 9), 5*math.Log(100); math.Abs(got-want) > 1e-9 {
		t.Fatalf("franchise = %v, want %v", got, want)
	}
	// a feature on almost every item floors at 0.1 before omega
	if got, want := Weight("genre", 100, 99), 0.1; math.Abs(got-want) > 1e-9 {
		t.Fatalf("floor = %v, want %v", got, want)
	}
	if got := Weight("country", 100, 99); math.Abs(got-0.03) > 1e-9 {
		t.Fatalf("country floor = %v, want 0.03", got)
	}
	if got := Weight("unknown_role", 1000, 1); got != 0 {
		t.Fatalf("unknown role = %v, want 0", got)
	}
	if Omega("developer") != 3 || Omega("publisher") != 1.5 || Omega("writer") != 2 || Omega("platform") != 0.5 {
		t.Fatal("omega table mismatch")
	}
}

func TestSimilarityExtractFeatures(t *testing.T) {
	entries := []Entry{
		{Role: "cast", Entity: "c0", Position: 0},
		{Role: "cast", Entity: "c4", Position: 4},
		{Role: "cast", Entity: "c5", Position: 5},
		{Role: "franchise", Entity: "zelda"},
		{Role: "franchise", Entity: "nintendo-franchise"},
		{Role: "genre", Entity: "adventure"},
	}
	parents := map[string][]string{
		"zelda":              {"nintendo-franchise", "mythos"},
		"nintendo-franchise": {"mythos"},
	}
	got := ExtractFeatures(entries, parents)

	want := map[Feature]float64{
		{"cast", "c0"}:                      1,
		{"cast", "c4"}:                      1,
		{"franchise", "zelda"}:              1,
		{"franchise", "nintendo-franchise"}: 1, // direct beats half-weight parent
		{"franchise", "mythos"}:             0.5,
		{"genre", "adventure"}:              1,
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for f, m := range want {
		if got[f] != m {
			t.Fatalf("%v = %v, want %v", f, got[f], m)
		}
	}
}
