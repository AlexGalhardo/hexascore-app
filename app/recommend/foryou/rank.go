package foryou

import (
	"cmp"
	"math"
	"slices"
	"time"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/similar"
)

const (
	// MaxItems is the length of a "For you" list (FR-08-9).
	MaxItems = 20
	// MaxPerFranchise is the per-franchise diversity cap (FR-08-9).
	MaxPerFranchise = 3
	// MaxPerGenre is the per-primary-genre diversity cap (FR-08-9).
	MaxPerGenre = 4
	// BayesWeight scales the item's Bayesian score B_j in s(u,j).
	BayesWeight = 0.1

	// BasisPersonal labels a personalised Result.
	BasisPersonal = "personal"
	// BasisOnboarding labels the fallback for users with fewer than MinPositives positive items.
	BasisOnboarding = "onboarding"
)

// Candidate is an item that may be recommended. Bayes is B_j (0-100), nil when the item has fewer than 5 ratings.
// Franchise and PrimaryGenre (genre at position 0) are empty when unknown and then count against no cap.
type Candidate struct {
	ID           string
	Title        string
	Features     map[similar.Feature]float64
	Bayes        *float64
	Released     bool
	Public       bool
	Franchise    string
	PrimaryGenre string
}

// Pick is one recommended item with its explanation.
type Pick struct {
	ID     string
	Score  float64
	Reason string
	// ReasonID is the item of the user that caused the pick.
	ReasonID string
}

// Result is a computed list. Items is empty when Basis is BasisOnboarding.
type Result struct {
	Basis string
	Items []Pick
}

// Score is s(u,j): the profile dotted with the candidate's feature weights, normalised by the candidate's weight norm,
// plus a small Bayesian-score bonus.
func Score(p Profile, c Candidate, weight func(similar.Feature) float64) float64 {
	var dot, norm float64

	for f := range c.Features {
		g := weight(f)
		dot += p.P[f] * g
		norm += g * g
	}

	var s float64
	if norm > 0 {
		s = dot / math.Sqrt(norm)
	}

	if c.Bayes != nil {
		s += BayesWeight * *c.Bayes / 100
	}

	return s
}

// Recommend scores every eligible candidate and returns the re-ranked list. ints must contain every item the user has
// rated or given a status, since those are removed from the output. Fewer than MinPositives positive items yields the
// onboarding result; the caller then shows trending (SDD 08 section 7.5). A candidate with no positive interaction
// sharing a feature has no explanation and is dropped.
func Recommend(ints []Interaction, cands []Candidate, now time.Time, weight func(similar.Feature) float64) Result {
	p := BuildProfile(ints, now, weight)
	if p.Positives < MinPositives {
		return Result{Basis: BasisOnboarding}
	}

	seen := make(map[string]bool, len(ints))
	for _, in := range ints {
		seen[in.ItemID] = true
	}

	type scored struct {
		c Candidate
		p Pick
	}

	var all []scored

	for _, c := range cands {
		if seen[c.ID] || !c.Released || !c.Public {
			continue
		}

		s := Score(p, c, weight)
		if s <= 0 {
			continue
		}

		reason, reasonID, ok := explain(ints, c, p.Mu, now, weight)
		if !ok {
			continue
		}

		all = append(all, scored{c, Pick{ID: c.ID, Score: s, Reason: reason, ReasonID: reasonID}})
	}

	slices.SortFunc(all, func(a, b scored) int {
		return cmp.Or(cmp.Compare(b.p.Score, a.p.Score), cmp.Compare(a.c.ID, b.c.ID))
	})

	res := Result{Basis: BasisPersonal}
	fr, ge := map[string]int{}, map[string]int{}

	for _, s := range all {
		if len(res.Items) == MaxItems {
			break
		}

		if (s.c.Franchise != "" && fr[s.c.Franchise] >= MaxPerFranchise) ||
			(s.c.PrimaryGenre != "" && ge[s.c.PrimaryGenre] >= MaxPerGenre) {
			continue
		}

		fr[s.c.Franchise]++
		ge[s.c.PrimaryGenre]++

		res.Items = append(res.Items, s.p)
	}

	return res
}
