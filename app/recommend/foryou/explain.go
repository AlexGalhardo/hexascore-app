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
) (string, string, bool) {
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
		return "", "", false
	}

	reason, ok := Reason(best.Title, best.Overall, best.Status)

	return reason, best.ItemID, ok
}

// Reason renders the explanation of a recommendation from the interaction that caused it. ok is false for a status
// with no wording (in_progress, dropped).
func Reason(title string, overall *float64, status string) (string, bool) {
	if overall != nil {
		return fmt.Sprintf("Because you rated %s %s/5", title, strconv.FormatFloat(*overall/20, 'f', -1, 64)), true
	}

	if verb, ok := statusVerb[status]; ok {
		return fmt.Sprintf("Because you %s %s", verb, title), true
	}

	return "", false
}
