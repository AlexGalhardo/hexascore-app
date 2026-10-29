// Package scoring implements HexaScore's rating aggregate (SDD 02).
package scoring

import (
	"math"
	"time"
)

// Categories is the number of rating categories (c1..c6).
const Categories = 6

// minVisible is the ratingCount below which scores are not shown (legacy
// MIN_RATINGS_FOR_SCORE); stored values stay the raw aggregate (SDD 02 §5).
const minVisible = 5

// Rating is one counted input: category values 5..100 and its weight.
type Rating struct {
	C      [Categories]float64
	Weight float64
}

// Overall is the unrounded mean of the six categories (D-13).
func (r Rating) Overall() float64 {
	var sum float64
	for _, c := range r.C {
		sum += c
	}
	return sum / Categories
}

// Params are the per-type formula inputs: m_t and the priors C_{t,k}.
type Params struct {
	M     float64
	Prior [Categories]float64
}

// Confidence levels (SDD 02 §7.1); empty means fewer than 5 ratings.
const (
	ConfidenceLow    = "low"
	ConfidenceMedium = "medium"
	ConfidenceHigh   = "high"
)

// Result is the aggregate of one item.
type Result struct {
	N          int
	W          float64
	B          float64
	BK         [Categories]float64
	SE         float64
	Confidence string
}

// Stars is B on the 0–10 scale with one decimal, rounded half away from
// zero; Visible reports whether it may be shown (N ≥ 5).
func (r Result) Stars() (stars float64, visible bool) {
	return math.Round(r.B/10*10) / 10, r.N >= minVisible
}

// Compute applies formula v1: B_k = (S_k + m·C_k) / (W + m), B = mean of B_k.
// Ratings with weight 0 are not counted.
func Compute(ratings []Rating, p Params) Result {
	var res Result
	var s [Categories]float64
	var sumWO float64
	for _, r := range ratings {
		if r.Weight <= 0 {
			continue
		}
		res.N++
		res.W += r.Weight
		for k, c := range r.C {
			s[k] += r.Weight * c
		}
		sumWO += r.Weight * r.Overall()
	}

	for k := range Categories {
		res.BK[k] = (s[k] + p.M*p.Prior[k]) / (res.W + p.M)
		res.B += res.BK[k]
	}
	res.B /= Categories

	if res.W > 0 {
		mean := sumWO / res.W
		var v float64
		for _, r := range ratings {
			if r.Weight > 0 {
				d := r.Overall() - mean
				v += r.Weight * d * d
			}
		}
		res.SE = math.Sqrt(v/res.W) / math.Sqrt(res.W)
	}

	switch {
	case res.N < minVisible:
	case res.N >= 50 && res.SE <= 2.0:
		res.Confidence = ConfidenceHigh
	case res.N >= 15 && res.SE <= 4.0:
		res.Confidence = ConfidenceMedium
	default:
		res.Confidence = ConfidenceLow
	}
	return res
}

// Factors are the inputs of one rating's weight w_u = a·v·s·r (SDD 02 §7.1).
type Factors struct {
	AccountAge    time.Duration
	Status        string // "", want, in_progress, completed, dropped
	IsGame        bool
	HasTimeToBeat bool
	Shadow        float64 // rating_flags.shadow_weight, 1 when unflagged
	Banned        bool
	BeforeRelease bool // rated before the effective release date (§7.4)
}

// Weight returns w_u for formula v1.
func Weight(f Factors) float64 {
	if f.Banned || f.BeforeRelease {
		return 0
	}
	days := f.AccountAge.Hours() / 24
	var a float64
	switch {
	case days < 7:
		a = 0.25
	case days < 30:
		a = 0.5
	case days < 90:
		a = 0.75
	default:
		a = 1
	}
	var v float64
	switch f.Status {
	case "in_progress", "dropped":
		v = 0.85
	case "completed":
		v = 1
		if f.IsGame && !f.HasTimeToBeat {
			v = 0.9
		}
	default:
		v = 0.7
	}
	return a * v * f.Shadow
}
