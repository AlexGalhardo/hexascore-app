// Package trending scores catalogue items for the home and API trending lists (SDD 08 §7.2).
package trending

import "math"

// HalfLifeDays is the decay half-life of a day's activity.
const HalfLifeDays = 7

// MinActors is the eligibility floor for A_i.
const MinActors = 5

// MinTrend is the eligibility floor for T_i.
const MinTrend = 30

// Day is one UTC day's activity for a single item. Weights are already the
// per-user maximum for that primitive (SDD 08 §7.1).
type Day struct {
	Age             int
	Views           float64
	SearchClicks    float64
	RatingsNewW     float64
	RatingsUpdatedW float64
	WantW           float64
	InProgressW     float64
	CompletedW      float64
	DroppedW        float64
	ListAddsW       float64
	SuggestionsW    float64
}

// Decay is D(a) = 2^(-a/7).
func Decay(ageDays int) float64 {
	return math.Pow(2, -float64(ageDays)/HalfLifeDays)
}

// Quality is the anonymous, damped part of a day: 10·ln(1+views+2·search_clicks).
func Quality(d Day) float64 {
	return 10 * math.Log(1+d.Views+2*d.SearchClicks)
}

// Engaged is the signed-in part of a day.
func Engaged(d Day) float64 {
	return 6*d.RatingsNewW + d.RatingsUpdatedW + 3*d.WantW + 4*d.InProgressW +
		5*d.CompletedW + d.DroppedW + 4*d.ListAddsW + 2*d.SuggestionsW
}

// Actors is the undamped sum that eligibility uses.
func Actors(d Day) float64 {
	return d.RatingsNewW + d.WantW + d.InProgressW + d.CompletedW + d.ListAddsW
}

// Score totals T_i and A_i over the supplied days.
func Score(days []Day) (trend, actors float64) {
	for _, d := range days {
		trend += Decay(d.Age) * (Quality(d) + Engaged(d))
		actors += Actors(d)
	}

	return trend, actors
}

// Eligible is A ≥ 5 ∧ T ≥ 30 ∧ released ∧ public ∧ not excluded.
func Eligible(trend, actors float64, released, public, excluded bool) bool {
	return actors >= MinActors && trend >= MinTrend && released && public && !excluded
}

// PriorFill is F_i for a slot the trend list did not fill.
// A release older than 180 days contributes only the popularity term.
func PriorFill(popularityPrior, daysSinceRelease float64) float64 {
	fresh := 1 - daysSinceRelease/180
	if fresh < 0 {
		fresh = 0
	}

	return math.Log(1+popularityPrior) + 3*fresh
}
