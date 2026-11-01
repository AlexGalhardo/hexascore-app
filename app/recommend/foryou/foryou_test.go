package foryou_test

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/foryou"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/similar"
)

var now = time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)

func feat(role, ent string) similar.Feature { return similar.Feature{Role: role, Entity: ent} }

func fs(fts ...similar.Feature) map[similar.Feature]float64 {
	m := map[similar.Feature]float64{}
	for _, f := range fts {
		m[f] = 1
	}

	return m
}

func w(similar.Feature) float64 { return 2 }

func rated(id string, overall float64, daysAgo int, fts ...similar.Feature) foryou.Interaction {
	return foryou.Interaction{
		ItemID: id, Title: "T" + id, Overall: &overall, At: now.AddDate(0, 0, -daysAgo), Features: fs(fts...),
	}
}

func TestProfileMuAndRho(t *testing.T) {
	if got := foryou.Mu([]float64{90, 90, 90, 90}); got != 60 {
		t.Fatalf("N<5 must use 60, got %v", got)
	}

	if got := foryou.Mu([]float64{80, 80, 80, 80, 90}); got != 82 {
		t.Fatalf("mean of 5, got %v", got)
	}

	hi, lo := 100.0, 0.0
	if got := foryou.Rho(&hi, "", 20); got != 1 {
		t.Fatalf("clamp high, got %v", got)
	}

	if got := foryou.Rho(&lo, "", 90); got != -1 {
		t.Fatalf("clamp low, got %v", got)
	}

	mid := 70.0
	if got := foryou.Rho(&mid, "", 60); math.Abs(got-0.2) > 1e-9 {
		t.Fatalf("(70-60)/50, got %v", got)
	}

	for st, want := range map[string]float64{"completed": 0.3, "want": 0.2, "dropped": -0.5, "in_progress": 0} {
		if got := foryou.Rho(nil, st, 60); got != want {
			t.Fatalf("status %s: got %v want %v", st, got, want)
		}
	}
}

func TestProfileTau(t *testing.T) {
	if got := foryou.Tau(0); got != 1 {
		t.Fatalf("tau(0) = %v", got)
	}

	if got := foryou.Tau(180); math.Abs(got-0.5) > 1e-9 {
		t.Fatalf("tau(180) = %v", got)
	}
}

func TestProfileWeights(t *testing.T) {
	a := feat("genre", "a")
	p := foryou.BuildProfile([]foryou.Interaction{rated("1", 100, 0, a), rated("2", 20, 0, a)}, now, w)

	// rho = +0.8 and -0.8 at tau 1 cancel out.
	if p.P[a] != 0 {
		t.Fatalf("p[a] = %v", p.P[a])
	}

	p = foryou.BuildProfile([]foryou.Interaction{rated("1", 100, 180, a)}, now, w)
	// mu = 60 (N<5): rho = 0.8, tau = 0.5, g = 2.
	if math.Abs(p.P[a]-0.8) > 1e-9 {
		t.Fatalf("p[a] = %v", p.P[a])
	}
}

func TestProfileOnboardingBelowThreePositives(t *testing.T) {
	a := feat("genre", "a")
	ints := []foryou.Interaction{rated("1", 90, 1, a), rated("2", 95, 1, a), rated("3", 10, 1, a)}
	cands := []foryou.Candidate{{ID: "c", Title: "C", Features: fs(a), Released: true, Public: true}}

	res := foryou.Recommend(ints, cands, now, w)
	if res.Basis != "onboarding" || len(res.Items) != 0 {
		t.Fatalf("got %+v", res)
	}
}

func TestProfileCapsAndExplanation(t *testing.T) {
	a := feat("genre", "a")
	ints := []foryou.Interaction{rated("1", 90, 1, a), rated("2", 95, 1, a), rated("3", 99, 1, a)}

	var cands []foryou.Candidate

	for i := range 40 {
		c := foryou.Candidate{
			ID: fmt.Sprintf("c%02d", i), Title: "C", Features: fs(a), Released: true, Public: true,
			Franchise: fmt.Sprintf("f%d", i/5), PrimaryGenre: fmt.Sprintf("g%d", i/8),
		}
		cands = append(cands, c)
	}

	// Rated, statused, unreleased and hidden items never appear.
	cands = append(cands,
		foryou.Candidate{ID: "1", Features: fs(a), Released: true, Public: true},
		foryou.Candidate{ID: "u", Features: fs(a), Released: false, Public: true},
		foryou.Candidate{ID: "h", Features: fs(a), Released: true, Public: false},
	)

	ints = append(ints, foryou.Interaction{ItemID: "wanted", Status: "want", At: now})
	cands = append(cands, foryou.Candidate{ID: "wanted", Features: fs(a), Released: true, Public: true})

	res := foryou.Recommend(ints, cands, now, w)
	if res.Basis != "personal" {
		t.Fatalf("basis %q", res.Basis)
	}

	if len(res.Items) > 20 || len(res.Items) == 0 {
		t.Fatalf("len %d", len(res.Items))
	}

	fr, ge := map[string]int{}, map[string]int{}
	byID := map[string]foryou.Candidate{}

	for _, c := range cands {
		byID[c.ID] = c
	}

	for _, it := range res.Items {
		switch it.ID {
		case "1", "u", "h", "wanted":
			t.Fatalf("filtered item %s present", it.ID)
		}

		c := byID[it.ID]
		fr[c.Franchise]++
		ge[c.PrimaryGenre]++

		if !strings.HasPrefix(it.Reason, "Because you rated T") || !strings.HasSuffix(it.Reason, "/5") {
			t.Fatalf("reason %q", it.Reason)
		}
	}

	for k, n := range fr {
		if n > 3 {
			t.Fatalf("franchise %s has %d", k, n)
		}
	}

	for k, n := range ge {
		if n > 4 {
			t.Fatalf("genre %s has %d", k, n)
		}
	}
}

func TestExplainPicksTopContributor(t *testing.T) {
	a, b := feat("franchise", "a"), feat("genre", "b")
	ints := []foryou.Interaction{
		rated("1", 90, 0, b), rated("2", 100, 0, a, b), rated("3", 80, 0, feat("genre", "z")),
	}
	c := foryou.Candidate{ID: "c", Features: fs(a, b), Released: true, Public: true}

	res := foryou.Recommend(ints, []foryou.Candidate{c}, now, w)
	if len(res.Items) != 1 || res.Items[0].Reason != "Because you rated T2 5/5" {
		t.Fatalf("got %+v", res)
	}
}

func TestTopSources(t *testing.T) {
	a, b := feat("genre", "a"), feat("genre", "b")
	ints := []foryou.Interaction{rated("1", 100, 0, a), rated("2", 70, 0, b), rated("3", 0, 0, b)}
	p := foryou.BuildProfile(ints, now, w)

	items := foryou.TopItems(ints, now, 1)
	if len(items) != 1 || items[0] != "1" {
		t.Fatalf("top items %v", items)
	}

	if f := foryou.TopFeatures(p, 1); len(f) != 1 || f[0] != a {
		t.Fatalf("top features %v", f)
	}
}
