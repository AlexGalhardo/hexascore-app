package trending

import (
	"cmp"
	"slices"
	"time"
)

// Home list rules (RFC-002 section 4).
const (
	// HomeListSize is the number of ranks per type in the home window.
	HomeListSize = 12
	// CooldownHours is how long an item keeps its slot after it last met the entry rule.
	CooldownHours = 72
	// WeakRunsToExit is how many consecutive runs with z below 0 evict an item that is still inside the cooldown.
	WeakRunsToExit = 2
	// MaxPerFranchise and MaxPerGenre cap how many of the 12 slots one franchise or primary genre may take.
	MaxPerFranchise = 2
	MaxPerGenre     = 4
)

// Prev is what the previous 7-day snapshot knew about an item that was on its list.
type Prev struct {
	EnteredAt time.Time
	// MetAt is the last run at which the item met the entry rule; zero when it was on the list as "prior".
	MetAt    time.Time
	WeakRuns int
}

// HomeCandidate is an eligible item (public, released, quality floor met, not banned) with its v2 numbers. Franchise
// and Genre are empty when the item has none.
type HomeCandidate struct {
	ID               string
	Trend, Actors, Z float64
	Prior            float64
	DaysSinceRelease float64
	Franchise, Genre string
	Pinned           bool
	PinOrder         int
	Prev             *Prev
}

// HomeEntry is one rank of the home list with the state the next run reads back.
type HomeEntry struct {
	Entry
	Z         float64
	EnteredAt time.Time
	MetAt     time.Time
	WeakRuns  int
}

type stateful struct {
	HomeCandidate
	entered, met time.Time
	weak         int
	trend        bool
}

// PickHome builds the home list of one type: pins first, then the items that meet the entry rule or are inside their
// cooldown, then the prior fill (basis "prior"). Franchise and genre caps skip an item and take the next one. The
// result depends on the set of candidates, not their order, so reruns are identical.
func PickHome(cands []HomeCandidate, now time.Time) []HomeEntry {
	var pins, trend, rest []stateful

	cooldown := CooldownHours * time.Hour

	for _, c := range cands {
		s := stateful{HomeCandidate: c, entered: now}

		if c.Prev != nil {
			s.entered, s.met = c.Prev.EnteredAt, c.Prev.MetAt
		}

		if c.Z < 0 {
			s.weak = 1
			if c.Prev != nil {
				s.weak = c.Prev.WeakRuns + 1
			}
		}

		qualifies := c.Actors >= EntryActors && c.Z >= EntryZ
		if qualifies {
			s.met = now
		}

		retained := !qualifies && !s.met.IsZero() && now.Sub(s.met) <= cooldown && s.weak < WeakRunsToExit

		switch {
		case c.Pinned:
			pins = append(pins, s)
		case qualifies || retained:
			s.trend = true
			trend = append(trend, s)
		default:
			rest = append(rest, s)
		}
	}

	slices.SortFunc(pins, func(a, b stateful) int {
		return cmp.Or(cmp.Compare(a.PinOrder, b.PinOrder), cmp.Compare(a.ID, b.ID))
	})
	slices.SortFunc(trend, func(a, b stateful) int {
		return cmp.Or(cmp.Compare(b.Z, a.Z), cmp.Compare(b.Trend, a.Trend), cmp.Compare(b.Prior, a.Prior),
			cmp.Compare(a.ID, b.ID))
	})

	fill := func(c stateful) float64 { return PriorFill(c.Prior, c.DaysSinceRelease) }

	slices.SortFunc(rest, func(a, b stateful) int {
		return cmp.Or(cmp.Compare(fill(b), fill(a)), cmp.Compare(b.Prior, a.Prior), cmp.Compare(a.ID, b.ID))
	})

	var (
		out        []HomeEntry
		franchises = map[string]int{}
		genres     = map[string]int{}
	)

	take := func(s stateful, basis string, score float64, force bool) {
		if len(out) == HomeListSize {
			return
		}

		if !force && ((s.Franchise != "" && franchises[s.Franchise] >= MaxPerFranchise) ||
			(s.Genre != "" && genres[s.Genre] >= MaxPerGenre)) {
			return
		}

		franchises[s.Franchise]++
		genres[s.Genre]++

		out = append(out, HomeEntry{
			Rank:      len(out) + 1,
			ID:        s.ID,
			Score:     score,
			Basis:     basis,
			Z:         s.Z,
			EnteredAt: s.entered,
			MetAt:     s.met,
			WeakRuns:  s.weak,
		})
	}

	for _, s := range pins {
		take(s, "pinned", s.Trend, true)
	}

	for _, s := range trend {
		take(s, "trend", s.Trend, false)
	}

	for _, s := range rest {
		take(s, "prior", fill(s), false)
	}

	return out
}
