package foryou

import (
	"fmt"
	"strconv"
	"time"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/similar"
)

var statusVerb = map[string]string{"completed": "completed", "want": "wanted"}

// explain names the positive interaction that contributes most to the candidate (sum of rho*tau*g_f^2 over shared
// features); ties go to the lower item id. ok is false when no positive interaction shares a feature.
func explain(
	ints []Interaction,
	c Candidate,
	mu float64,
	now time.Time,
	weight func(similar.Feature) float64,
) (string, bool) {
	var (
		best      Interaction
		bestScore float64
		found     bool
	)

	for _, in := range ints {
		rt := rhoTau(in, mu, now)
		if rt <= 0 {
			continue
		}

		var s float64

		for f := range in.Features {
			if _, ok := c.Features[f]; ok {
				g := weight(f)
				s += rt * g * g
			}
		}

		if s > 0 && (!found || s > bestScore || (s == bestScore && in.ItemID < best.ItemID)) {
			best, bestScore, found = in, s, true
		}
	}

	if !found {
		return "", false
	}

	if best.Overall != nil {
		stars := strconv.FormatFloat(*best.Overall/20, 'f', -1, 64)

		return fmt.Sprintf("Because you rated %s %s/5", best.Title, stars), true
	}

	if verb, ok := statusVerb[best.Status]; ok {
		return fmt.Sprintf("Because you %s %s", verb, best.Title), true
	}

	return "", false
}
