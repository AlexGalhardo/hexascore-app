package trending

import "testing"

func TestTrendScore(t *testing.T) {
	t.Parallel()

	if got := []float64{Decay(0), Decay(7), Decay(14)}; got[0] != 1 || got[1] != 0.5 || got[2] != 0.25 {
		t.Fatalf("D(0,7,14) = %v, want 1, 0.5, 0.25", got)
	}

	days := []Day{
		{Age: 0, Views: 40, RatingsNewW: 3, WantW: 2},
		{Age: 7, Views: 100, CompletedW: 4, ListAddsW: 2},
		{Age: 14, Views: 200, RatingsNewW: 5},
	}
	trend, actors := Score(days)
	if !near(trend, 118.97, 0.01) || !near(actors, 16, 0.01) {
		t.Fatalf("item X: T=%v A=%v, want 118.97 and 16", trend, actors)
	}

	if !Eligible(trend, actors, true, true, false) {
		t.Fatal("item X should be eligible")
	}

	burst := []Day{{Age: 0, Views: 300, RatingsNewW: 2.625}}
	trend, actors = Score(burst)
	// SDD 08 prints 72.8; 10·ln(301)+6·2.625 is 72.82, inside 0.03 of that rounding.
	if !near(trend, 72.8, 0.03) || !near(actors, 2.625, 0.01) {
		t.Fatalf("item Y: T=%v A=%v, want 72.8 and 2.625", trend, actors)
	}

	if Eligible(trend, actors, true, true, false) {
		t.Fatal("item Y should not be eligible")
	}

	if Eligible(30, 4.99, true, true, false) {
		t.Fatal("A 4.99 is below the floor")
	}

	if !Eligible(30, 5, true, true, false) {
		t.Fatal("A 5.0 on the floor is eligible")
	}
}

func TestFallback(t *testing.T) {
	t.Parallel()

	if got := PriorFill(80, 30); !near(got, 6.89, 0.01) {
		t.Fatalf("80 sitelinks at 30 days: F=%v, want 6.89", got)
	}

	if got := PriorFill(200, 1000); !near(got, 5.30, 0.01) {
		t.Fatalf("classic with 200 sitelinks: F=%v, want 5.30", got)
	}
}

func near(got, want, tol float64) bool {
	d := got - want
	if d < 0 {
		d = -d
	}

	return d <= tol
}
