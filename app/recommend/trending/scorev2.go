package trending

import (
	"math"
	"slices"
)

// Score v2 (RFC-002 section 3) ranks the 7-day home window. The 30-day API snapshot keeps score v1 (score.go).
const (
	// HomeWindowDays is how many UTC days feed the home score.
	HomeWindowDays = 7
	// HomeHalfLifeDays is the decay half-life of a day's activity in the home window.
	HomeHalfLifeDays = 3.5
	// EntryActors is the least distinct-account activity (A) an item needs to enter the home list.
	EntryActors = 5
	// EntryZ is the least per-type z-score an item needs to enter the home list.
	EntryZ = 0.5

	externalWeight   = 8 * 0.6 // weight 8, reliability 0.6 (RFC-002 section 2)
	listAddsWeight   = 4 * 0.8 // weight 4, reliability 0.8
	viewsReliability = 0.5
	recencyDays      = 90
	maxZ             = 10
)

// Signals is everything the v2 score reads for one item.
type Signals struct {
	// Days is the item's activity for the last HomeWindowDays UTC days; days with no row are absent.
	Days []Day
	// ExternalDelta is ln(value today / value 7 days ago) of the external popularity, clipped (ExternalDelta).
	ExternalDelta float64
	// DaysSinceRelease is negative for a release date in the future.
	DaysSinceRelease float64
}

// DecayHalf is D(a) = 2^(-a/half).
func DecayHalf(ageDays int, halfLifeDays float64) float64 {
	return math.Pow(2, -float64(ageDays)/halfLifeDays)
}

// ReleaseRecency is 3·max(0, 1 − days/90); an unreleased item (negative days) counts 0.
func ReleaseRecency(daysSinceRelease float64) float64 {
	if daysSinceRelease < 0 {
		return 0
	}

	return 3 * max(0, 1-daysSinceRelease/recencyDays)
}

// ExternalDelta is the 7-day relative change of an external popularity value, ln(today/before) clipped to [−1, 2].
// A source with no value 7 days ago (or switched off) gives 0 so it never moves the rank.
func ExternalDelta(today, before float64) float64 {
	if before <= 0 {
		return 0
	}

	if today <= 0 {
		return -1
	}

	return min(2, max(-1, math.Log(today/before)))
}

// engagedV2 is U_d without the external signal: the signed-in part of a day.
func engagedV2(d Day) float64 {
	return 6*d.RatingsNewW + d.RatingsUpdatedW + 3*d.WantW + 4*d.InProgressW + 5*d.CompletedW + d.DroppedW +
		listAddsWeight*d.ListAddsW
}

// qualityV2 is Q_d: anonymous traffic, damped by ln and weighted 0.5.
func qualityV2(d Day) float64 {
	return viewsReliability * Quality(d)
}

// ScoreV2 totals T_i and A_i over the home window: decayed daily activity plus today's external delta and the release
// recency. A is the undecayed distinct-account activity that the entry rule reads.
func ScoreV2(s Signals) (trend, actors float64) {
	for _, d := range s.Days {
		if d.Age < 0 || d.Age >= HomeWindowDays {
			continue
		}

		trend += DecayHalf(d.Age, HomeHalfLifeDays) * (engagedV2(d) + qualityV2(d))
		actors += Actors(d)
	}

	return trend + externalWeight*s.ExternalDelta + ReleaseRecency(s.DaysSinceRelease), actors
}

// ZBasis is the median and spread of one type's scores, from which any item's z-score follows.
type ZBasis struct{ Median, Scale float64 }

// NewZBasis measures the scores of the items that count for the type (A ≥ 1 on the day, RFC-002 section 3). The scale
// is the MAD; when the MAD is 0 but the scores differ, the mean absolute deviation stands in so a lone leader still
// stands out. A zero Scale means nobody stands out and every z is 0.
func NewZBasis(trends []float64) ZBasis {
	if len(trends) == 0 {
		return ZBasis{}
	}

	med := median(trends)
	dev := make([]float64, len(trends))

	for i, t := range trends {
		dev[i] = math.Abs(t - med)
	}

	scale := median(dev)
	if scale == 0 {
		for _, d := range dev {
			scale += d
		}

		scale /= float64(len(dev))
	}

	return ZBasis{Median: med, Scale: scale}
}

// Z is (T − median)/scale, clipped to ±10 so one runaway item cannot push everyone else below the entry floor.
func (b ZBasis) Z(trend float64) float64 {
	if b.Scale == 0 {
		return 0
	}

	return min(maxZ, max(-maxZ, (trend-b.Median)/b.Scale))
}

// ZScores is the z-score of every score against the basis of the same slice, in the order given.
func ZScores(trends []float64) []float64 {
	b := NewZBasis(trends)
	out := make([]float64, len(trends))

	for i, t := range trends {
		out[i] = b.Z(t)
	}

	return out
}

func median(xs []float64) float64 {
	s := slices.Clone(xs)
	slices.Sort(s)

	n := len(s)
	if n%2 == 1 {
		return s[n/2]
	}

	return (s[n/2-1] + s[n/2]) / 2
}
