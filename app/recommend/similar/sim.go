// Package similar computes item-to-item similarity from shared catalogue features (SDD 08 §7.4).
package similar

import (
	"math"
	"sort"
)

var omega = map[string]float64{
	"franchise": 5, "series": 5,
	"author": 3, "director": 3, "creator": 3, "developer": 3,
	"writer":    2,
	"publisher": 1.5, "production_company": 1.5, "network": 1.5,
	"cast": 1, "genre": 1,
	"platform": 0.5, "game_mode": 0.5,
	"country": 0.3, "original_language": 0.3,
}

// Omega is the role weight ω_role; an unknown role weighs 0 and so never contributes.
func Omega(role string) float64 { return omega[role] }

// Weight is g_f = ω_role · max(0.1, ln(N_t / (1 + n_f))): total is N_t (public items of the type) and n the items carrying the feature.
func Weight(role string, total, n int) float64 {
	return Omega(role) * math.Max(0.1, math.Log(float64(total)/float64(1+n)))
}

// Sim is the weighted Jaccard of two feature sets. Each set maps a feature to its multiplier (1, or 0.5 for a franchise parent);
// weight returns g_f. Shared mass takes the smaller scaled weight and the union the larger, which reduces to Σ g_f(shared) / Σ g_f(union) when every multiplier is 1.
func Sim(a, b map[Feature]float64, weight func(Feature) float64) float64 {
	// Sorted iteration keeps the float sum independent of map order, so Sim(a,b) == Sim(b,a) bit for bit.
	keys := make([]Feature, 0, len(a)+len(b))
	for f := range a {
		keys = append(keys, f)
	}

	for f := range b {
		if _, ok := a[f]; !ok {
			keys = append(keys, f)
		}
	}

	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Role != keys[j].Role {
			return keys[i].Role < keys[j].Role
		}

		return keys[i].Entity < keys[j].Entity
	})

	var shared, union float64

	for _, f := range keys {
		w := weight(f)
		ma, inA := a[f]
		mb, inB := b[f]

		switch {
		case inA && inB:
			shared += w * math.Min(ma, mb)
			union += w * math.Max(ma, mb)
		case inA:
			union += w * ma
		default:
			union += w * mb
		}
	}

	if union == 0 {
		return 0
	}

	return shared / union
}
