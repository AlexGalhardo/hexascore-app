package trending

import (
	"math"
	"testing"
)

// RFC-002 section 3, worked example: T = 68.2, A = 9.
func TestScoreV2WorkedExample(t *testing.T) {
	t.Parallel()

	trend, actors := ScoreV2(Signals{
		Days: []Day{
			{Age: 0, Views: 40, RatingsNewW: 3, WantW: 2},
			{Age: 3, Views: 100, CompletedW: 4},
		},
		ExternalDelta:    0.4,
		DaysSinceRelease: 1000,
	})

	if !near(trend, 68.2, 0.2) || !near(actors, 9, 1e-9) {
		t.Fatalf("T=%v A=%v, want 68.2 and 9", trend, actors)
	}
}

// RFC-002 section 3, burst example: 300 views and 60 shadowed ratings give T about 44 and A 2.6, below the entry floor.
func TestScoreV2BurstStaysBelowEntry(t *testing.T) {
	t.Parallel()

	trend, actors := ScoreV2(Signals{
		Days:             []Day{{Age: 0, Views: 300, RatingsNewW: 2.6}},
		DaysSinceRelease: 1000,
	})

	if !near(trend, 44.1, 0.2) || !near(actors, 2.6, 1e-9) {
		t.Fatalf("T=%v A=%v, want 44.1 and 2.6", trend, actors)
	}

	if actors >= EntryActors {
		t.Fatalf("a burst with A=%v must not meet the entry floor %v", actors, EntryActors)
	}
}

// lcg is a tiny deterministic generator: the property tests need reproducible inputs, not randomness.
type lcg struct{ state uint64 }

func (g *lcg) next() uint64 {
	g.state = g.state*6364136223846793005 + 1442695040888963407

	return g.state >> 33
}

func (g *lcg) IntN(n int) int { return int(g.next() % uint64(n)) }

func (g *lcg) Float64() float64 { return float64(g.next()%(1<<30)) / float64(1<<30) }

func TestReleaseRecency(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		days, want float64
	}{{0, 3}, {45, 1.5}, {90, 0}, {400, 0}, {-5, 0}} {
		if got := ReleaseRecency(tc.days); !near(got, tc.want, 1e-9) {
			t.Fatalf("recency(%v) = %v, want %v", tc.days, got, tc.want)
		}
	}
}

func TestDecayHalfLife(t *testing.T) {
	t.Parallel()

	if got := DecayHalf(7, 3.5); !near(got, 0.25, 1e-9) {
		t.Fatalf("D(7) with half-life 3.5 = %v, want 0.25", got)
	}
}

func TestExternalDeltaIsClipped(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ today, before, want float64 }{
		{100, 100, 0}, {200, 100, math.Ln2}, {1000, 100, 2}, {1, 100, -1}, {50, 0, 0}, {0, 100, -1},
	} {
		if got := ExternalDelta(tc.today, tc.before); !near(got, tc.want, 1e-9) {
			t.Fatalf("delta(%v, %v) = %v, want %v", tc.today, tc.before, got, tc.want)
		}
	}
}

// Monotone: raising any one signal never lowers T or A.
func TestScoreV2MonotoneInEverySignal(t *testing.T) {
	t.Parallel()

	rng := &lcg{state: 1}
	bump := []func(*Day){
		func(d *Day) { d.Views++ }, func(d *Day) { d.SearchClicks++ }, func(d *Day) { d.RatingsNewW += 0.5 },
		func(d *Day) { d.RatingsUpdatedW += 0.5 }, func(d *Day) { d.WantW += 0.5 },
		func(d *Day) { d.InProgressW += 0.5 }, func(d *Day) { d.CompletedW += 0.5 },
		func(d *Day) { d.DroppedW += 0.5 }, func(d *Day) { d.ListAddsW += 0.5 },
	}

	for range 200 {
		days := make([]Day, 1+rng.IntN(5))
		for i := range days {
			days[i] = Day{
				Age: i, Views: float64(rng.IntN(300)), RatingsNewW: rng.Float64() * 5, WantW: rng.Float64() * 3,
				CompletedW: rng.Float64() * 2, ListAddsW: rng.Float64(),
			}
		}

		base := Signals{Days: days, ExternalDelta: rng.Float64(), DaysSinceRelease: float64(rng.IntN(200))}
		t0, a0 := ScoreV2(base)

		for i, f := range bump {
			more := base
			more.Days = append([]Day(nil), days...)
			f(&more.Days[rng.IntN(len(days))])

			if t1, a1 := ScoreV2(more); t1 < t0 || a1 < a0 {
				t.Fatalf("bump %d lowered the score: %v/%v -> %v/%v", i, t0, a0, t1, a1)
			}
		}
	}
}

// z does not change when every item's activity is scaled by the same factor.
func TestZScoresInvariantToScale(t *testing.T) {
	t.Parallel()

	rng := &lcg{state: 3}
	ts := make([]float64, 60)

	for i := range ts {
		ts[i] = 5 + rng.Float64()*200
	}

	base := ZScores(ts)

	for _, k := range []float64{0.01, 3, 1000} {
		scaled := make([]float64, len(ts))
		for i, v := range ts {
			scaled[i] = v * k
		}

		for i, z := range ZScores(scaled) {
			if !near(z, base[i], 1e-9) {
				t.Fatalf("scale %v: z[%d] = %v, want %v", k, i, z, base[i])
			}
		}
	}
}

func TestZScoresEdgeCases(t *testing.T) {
	t.Parallel()

	if got := ZScores(nil); len(got) != 0 {
		t.Fatalf("empty input: %v", got)
	}

	// every item equal: nobody stands out
	for _, z := range ZScores([]float64{7, 7, 7}) {
		if z != 0 {
			t.Fatalf("equal scores must give z = 0, got %v", z)
		}
	}

	// median 10, MAD 5: one clear leader
	z := ZScores([]float64{5, 10, 15, 40})
	if !(z[3] > z[2] && z[2] > z[1] && z[1] > z[0]) {
		t.Fatalf("z must keep the order of T: %v", z)
	}
}
