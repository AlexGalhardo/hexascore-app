package similar

import (
	"math"
	"sort"
)

// MaxNeighbours caps item_similar.similar_ids.
const MaxNeighbours = 20

const (
	minSim          = 0.05
	strongSim       = 0.15
	strongRoleOmega = 1.5
)

// Candidate is another item that shares at least one feature with the subject.
type Candidate struct {
	ID         string
	Features   map[Feature]float64
	Popularity float64
}

// Scored is a kept neighbour. Sim is already rounded to two decimals, the precision the order uses.
type Scored struct {
	ID         string
	Sim        float64
	Cooccur    int
	Popularity float64
}

// Filter scores every candidate against the subject and applies the keep rule of SDD 08 §7.4: sim >= 0.05 and either a
// shared feature with omega >= 1.5 or sim >= 0.15.
func Filter(self map[Feature]float64, cands []Candidate, weight func(Feature) float64) []Scored {
	var out []Scored

	for _, c := range cands {
		s := Sim(self, c.Features, weight)
		if s < minSim {
			continue
		}

		if s < strongSim && !sharesStrongRole(self, c.Features) {
			continue
		}

		out = append(out, Scored{ID: c.ID, Sim: math.Round(s*100) / 100, Popularity: c.Popularity})
	}

	return out
}

func sharesStrongRole(a, b map[Feature]float64) bool {
	for f := range a {
		if _, ok := b[f]; ok && Omega(f.Role) >= strongRoleOmega {
			return true
		}
	}

	return false
}

// Order sorts by rounded sim, then cooccur, then popularity (all descending), then id, and keeps MaxNeighbours.
func Order(in []Scored) []Scored {
	sort.Slice(in, func(i, j int) bool {
		a, b := in[i], in[j]

		switch {
		case a.Sim != b.Sim:
			return a.Sim > b.Sim
		case a.Cooccur != b.Cooccur:
			return a.Cooccur > b.Cooccur
		case a.Popularity != b.Popularity:
			return a.Popularity > b.Popularity
		default:
			return a.ID < b.ID
		}
	})

	if len(in) > MaxNeighbours {
		in = in[:MaxNeighbours]
	}

	return in
}
