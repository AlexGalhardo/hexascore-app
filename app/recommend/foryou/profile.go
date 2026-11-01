// Package foryou holds the pure scoring and re-ranking functions behind "For you" (SDD 08 section 7.5).
package foryou

import (
	"cmp"
	"math"
	"slices"
	"time"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/similar"
)

const (
	// DefaultMu is the neutral mean used until the user has MinRatedForMu ratings.
	DefaultMu = 60.0
	// MinRatedForMu is the smallest N_u for which the user's own mean replaces DefaultMu.
	MinRatedForMu = 5
	// HalfLifeDays is the half-life of an interaction's influence.
	HalfLifeDays = 180.0
	// MinPositives is the number of positive items below which the user gets the onboarding fallback.
	MinPositives = 3
)

// Interaction is one rated or statused item of the user. Overall is the 0-100 rating, nil when the item only has a status.
type Interaction struct {
	ItemID   string
	Title    string
	Overall  *float64
	Status   string
	At       time.Time
	Features map[similar.Feature]float64
}

// Profile is the user's taste vector p_u.
type Profile struct {
	Mu        float64
	P         map[similar.Feature]float64
	Positives int
}

// Mu is the user's mean overall rating when at least MinRatedForMu are given, else DefaultMu.
func Mu(overalls []float64) float64 {
	if len(overalls) < MinRatedForMu {
		return DefaultMu
	}

	var sum float64
	for _, o := range overalls {
		sum += o
	}

	return sum / float64(len(overalls))
}

// Rho is the signed preference in [-1, 1]: the rating's distance from mu over 50, or a fixed value from the status when unrated.
func Rho(overall *float64, status string, mu float64) float64 {
	if overall != nil {
		return max(-1, min(1, (*overall-mu)/50))
	}

	switch status {
	case "completed":
		return 0.3
	case "want":
		return 0.2
	case "dropped":
		return -0.5
	default:
		return 0
	}
}

// Tau is the recency factor 2^(-age/180).
func Tau(ageDays float64) float64 { return math.Pow(2, -ageDays/HalfLifeDays) }

func rhoTau(in Interaction, mu float64, now time.Time) float64 {
	return Rho(in.Overall, in.Status, mu) * Tau(max(0, now.Sub(in.At).Hours()/24))
}

func muOf(ints []Interaction) float64 {
	var o []float64

	for _, in := range ints {
		if in.Overall != nil {
			o = append(o, *in.Overall)
		}
	}

	return Mu(o)
}

// BuildProfile computes p_u[f] = sum over items of rho * tau * g_f for every feature of the item.
func BuildProfile(ints []Interaction, now time.Time, weight func(similar.Feature) float64) Profile {
	mu := muOf(ints)
	p := Profile{Mu: mu, P: map[similar.Feature]float64{}}

	for _, in := range ints {
		if Rho(in.Overall, in.Status, mu) > 0 {
			p.Positives++
		}

		rt := rhoTau(in, mu, now)
		for f := range in.Features {
			p.P[f] += rt * weight(f)
		}
	}

	return p
}

// TopItems returns the ids of the n interactions with the largest positive rho*tau (candidate source 1: their item_similar rows).
func TopItems(ints []Interaction, now time.Time, n int) []string {
	mu := muOf(ints)
	s := slices.Clone(ints)

	slices.SortFunc(s, func(a, b Interaction) int {
		return cmp.Or(cmp.Compare(rhoTau(b, mu, now), rhoTau(a, mu, now)), cmp.Compare(a.ItemID, b.ItemID))
	})

	var out []string

	for _, in := range s[:min(n, len(s))] {
		if rhoTau(in, mu, now) > 0 {
			out = append(out, in.ItemID)
		}
	}

	return out
}

// TopFeatures returns the n features with the largest positive p_u (candidate source 2: the best items carrying them).
func TopFeatures(p Profile, n int) []similar.Feature {
	var fs []similar.Feature

	for f, v := range p.P {
		if v > 0 {
			fs = append(fs, f)
		}
	}

	slices.SortFunc(fs, func(a, b similar.Feature) int {
		return cmp.Or(cmp.Compare(p.P[b], p.P[a]), cmp.Compare(a.Role, b.Role), cmp.Compare(a.Entity, b.Entity))
	})

	return fs[:min(n, len(fs))]
}
