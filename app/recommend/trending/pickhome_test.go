package trending

import (
	"fmt"
	"slices"
	"testing"
	"time"
)

var t0 = time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)

func ids(es []HomeEntry) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.ID
	}

	return out
}

func strong(id string, z float64) HomeCandidate {
	return HomeCandidate{ID: id, Trend: 100 + z, Actors: 8, Z: z, Prior: 1, DaysSinceRelease: 500}
}

func TestPickHomeEntryRule(t *testing.T) {
	t.Parallel()

	got := PickHome([]HomeCandidate{
		strong("ok", 0.5),
		{ID: "low-z", Trend: 90, Actors: 8, Z: 0.49, Prior: 0, DaysSinceRelease: 500},
		{ID: "low-a", Trend: 90, Actors: 4.99, Z: 3, Prior: 0, DaysSinceRelease: 500},
	}, t0)

	if got[0].ID != "ok" || got[0].Basis != "trend" {
		t.Fatalf("first = %+v", got[0])
	}

	for _, e := range got[1:] {
		if e.Basis != "prior" {
			t.Fatalf("%s entered on the trend: %+v", e.ID, e)
		}
	}
}

func TestPickHomeCooldownKeepsAFadingItem(t *testing.T) {
	t.Parallel()

	fading := HomeCandidate{
		ID: "fading", Trend: 40, Actors: 6, Z: 0.1, Prior: 1, DaysSinceRelease: 500,
		Prev: &Prev{EnteredAt: t0.Add(-100 * time.Hour), MetAt: t0.Add(-71 * time.Hour)},
	}

	if got := PickHome(
		[]HomeCandidate{fading},
		t0,
	); got[0].Basis != "trend" ||
		got[0].EnteredAt != fading.Prev.EnteredAt {
		t.Fatalf("inside the 72 h cooldown the item stays on the trend with its entry time: %+v", got[0])
	}

	fading.Prev.MetAt = t0.Add(-73 * time.Hour)
	if got := PickHome([]HomeCandidate{fading}, t0); got[0].Basis != "prior" {
		t.Fatalf("after the cooldown the item leaves the trend: %+v", got[0])
	}
}

func TestPickHomeTwoWeakRunsEvict(t *testing.T) {
	t.Parallel()

	c := HomeCandidate{
		ID: "x", Trend: 10, Actors: 6, Z: -0.2, Prior: 1, DaysSinceRelease: 500,
		Prev: &Prev{EnteredAt: t0.Add(-30 * time.Hour), MetAt: t0.Add(-10 * time.Hour), WeakRuns: 0},
	}

	first := PickHome([]HomeCandidate{c}, t0)
	if first[0].Basis != "trend" || first[0].WeakRuns != 1 {
		t.Fatalf("one run below zero is forgiven: %+v", first[0])
	}

	c.Prev.WeakRuns = first[0].WeakRuns
	if second := PickHome([]HomeCandidate{c}, t0.Add(time.Hour)); second[0].Basis != "prior" {
		t.Fatalf("two consecutive runs below zero evict: %+v", second[0])
	}
}

func TestPickHomeDiversity(t *testing.T) {
	t.Parallel()

	var cands []HomeCandidate

	for i := range 5 { // five entries of one franchise: two may stay
		c := strong(fmt.Sprintf("saga-%d", i), 9-float64(i))
		c.Franchise = "saga"
		cands = append(cands, c)
	}

	for i := range 6 { // six of one genre: four may stay
		c := strong(fmt.Sprintf("noir-%d", i), 4-float64(i)/10)
		c.Genre = "noir"
		cands = append(cands, c)
	}

	got := PickHome(cands, t0)

	saga, noir := 0, 0

	for _, e := range got {
		switch {
		case len(e.ID) > 4 && e.ID[:4] == "saga":
			saga++
		case len(e.ID) > 4 && e.ID[:4] == "noir":
			noir++
		}
	}

	if saga != MaxPerFranchise || noir != MaxPerGenre {
		t.Fatalf("saga=%d noir=%d, want %d and %d: %v", saga, noir, MaxPerFranchise, MaxPerGenre, ids(got))
	}
}

func TestPickHomePinsFirstInOrder(t *testing.T) {
	t.Parallel()

	a, b := strong("a", 5), strong("b", 6)
	pinned2 := HomeCandidate{ID: "pin2", Pinned: true, PinOrder: 2, Prior: 0, DaysSinceRelease: 500}
	pinned1 := HomeCandidate{ID: "pin1", Pinned: true, PinOrder: 1, Prior: 0, DaysSinceRelease: 500}

	got := PickHome([]HomeCandidate{a, b, pinned2, pinned1}, t0)
	if !slices.Equal(ids(got)[:4], []string{"pin1", "pin2", "b", "a"}) || got[0].Basis != "pinned" {
		t.Fatalf("order = %v", ids(got))
	}
}

func TestPickHomeFallbackIsLabelledPrior(t *testing.T) {
	t.Parallel()

	got := PickHome([]HomeCandidate{
		{ID: "new", Prior: 80, DaysSinceRelease: 30},
		{ID: "classic", Prior: 200, DaysSinceRelease: 9000},
	}, t0)

	if len(got) != 2 || got[0].ID != "new" || got[0].Basis != "prior" || got[1].Basis != "prior" {
		t.Fatalf("got %+v", got)
	}

	if got[0].Rank != 1 || got[1].Rank != 2 {
		t.Fatalf("ranks are not gapless: %+v", got)
	}
}

func TestPickHomeIsAPureFunctionOfTheSet(t *testing.T) {
	t.Parallel()

	var cands []HomeCandidate

	for i := range 30 {
		c := strong(fmt.Sprintf("id-%02d", i), float64(i%7))
		c.Franchise = fmt.Sprintf("f%d", i%9)
		cands = append(cands, c)
	}

	want := ids(PickHome(slices.Clone(cands), t0))
	rev := slices.Clone(cands)
	slices.Reverse(rev)

	if got := ids(PickHome(rev, t0)); !slices.Equal(got, want) {
		t.Fatalf("input order changed the list:\n%v\n%v", got, want)
	}

	if len(want) > HomeListSize {
		t.Fatalf("%d entries", len(want))
	}
}

// profile drives the synthetic suite: activity per day for one item.
type profile struct {
	id        string
	franchise string
	activity  func(day int) Day // Age is set by the caller
}

// RFC-002 section 6, synthetic: 1,000 items over 30 days, hourly runs, four kinds of item.
func TestSyntheticHomeListBehaviour(t *testing.T) {
	t.Parallel()

	const (
		days  = 30
		items = 1000
	)

	var profiles []profile

	for i := range items {
		id := fmt.Sprintf("i%04d", i)

		switch {
		case i < 8: // rising from day 12: activity grows each day (fewer than the 12 slots, so room is never the reason)
			profiles = append(profiles, profile{id: id, activity: func(d int) Day {
				if d < 12 {
					return Day{}
				}

				k := float64(d - 11)

				return Day{Views: 60 * k, RatingsNewW: 2 * k, WantW: k, CompletedW: k}
			}})
		case i < 28: // burst on day 20: huge traffic, few real accounts
			profiles = append(profiles, profile{id: id, activity: func(d int) Day {
				if d != 20 {
					return Day{}
				}

				return Day{Views: 400, RatingsNewW: 2.6}
			}})
		case i < 48: // fading: steady until day 10, then nothing
			profiles = append(profiles, profile{id: id, activity: func(d int) Day {
				if d > 10 {
					return Day{}
				}

				return Day{Views: 150, RatingsNewW: 3, WantW: 2, CompletedW: 1}
			}})
		default: // background noise: small steady activity
			profiles = append(profiles, profile{id: id, activity: func(int) Day {
				return Day{Views: float64(5 + i%11), RatingsNewW: float64(i % 2), WantW: 0.5 * float64(i%3)}
			}})
		}

		if i >= 100 && i < 110 { // ten near-identical franchise entries, strong enough to compete
			p := &profiles[len(profiles)-1]
			p.franchise = "saga"
			base := p.activity
			p.activity = func(d int) Day { x := base(d); x.Views += 200; x.RatingsNewW += 3; x.WantW += 2; return x }
		}
	}

	prev := map[string]*Prev{}
	firstIn := map[string]time.Time{}     // first run the item was on the trend
	lastMet := map[string]time.Time{}     // last run the item met the entry rule
	thresholdAt := map[string]time.Time{} // first run the item met the entry rule
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	for h := range days * 24 {
		now := start.Add(time.Duration(h) * time.Hour)
		day := h / 24

		type scored struct {
			p      profile
			trend  float64
			actors float64
		}

		all := make([]scored, 0, len(profiles))
		trends := make([]float64, 0, len(profiles))

		for _, p := range profiles {
			var ds []Day

			for age := 0; age < HomeWindowDays && day-age >= 0; age++ {
				d := p.activity(day - age)
				d.Age = age
				ds = append(ds, d)
			}

			tr, ac := ScoreV2(Signals{Days: ds, DaysSinceRelease: 1000})
			all = append(all, scored{p: p, trend: tr, actors: ac})
			trends = append(trends, tr)
		}

		zs := ZScores(trends)

		cands := make([]HomeCandidate, 0, len(all))

		for i, s := range all {
			c := HomeCandidate{
				ID: s.p.id, Trend: s.trend, Actors: s.actors, Z: zs[i], Prior: 1, DaysSinceRelease: 1000,
				Franchise: s.p.franchise, Prev: prev[s.p.id],
			}
			cands = append(cands, c)

			if s.actors >= EntryActors && zs[i] >= EntryZ {
				lastMet[s.p.id] = now

				if _, ok := thresholdAt[s.p.id]; !ok {
					thresholdAt[s.p.id] = now
				}
			}
		}

		entries := PickHome(cands, now)
		clear(prev)

		saga := 0

		for _, e := range entries {
			if e.Basis == "trend" {
				if _, ok := firstIn[e.ID]; !ok {
					firstIn[e.ID] = now
				}

				if m, ok := lastMet[e.ID]; !ok || now.Sub(m) > CooldownHours*time.Hour {
					t.Fatalf(
						"%s on the trend at %v, last met the entry rule at %v (over %d h)",
						e.ID,
						now,
						m,
						CooldownHours,
					)
				}
			}

			prev[e.ID] = &Prev{EnteredAt: e.EnteredAt, MetAt: e.MetAt, WeakRuns: e.WeakRuns}

			if len(e.ID) == 5 && e.ID >= "i0100" && e.ID <= "i0109" {
				saga++
			}
		}

		if saga > MaxPerFranchise {
			t.Fatalf("hour %d: %d saga entries", h, saga)
		}

		if len(entries) > HomeListSize {
			t.Fatalf("hour %d: %d entries", h, len(entries))
		}
	}

	for i := range 8 { // rising items enter within 24 h of meeting the entry rule
		id := fmt.Sprintf("i%04d", i)
		at, ok := thresholdAt[id]

		if !ok {
			continue // a franchise cap or a full list may keep a rising item out; it must not be a rule failure
		}

		if in, ok := firstIn[id]; ok && in.Sub(at) > 24*time.Hour {
			t.Fatalf("%s met the entry rule at %v but entered at %v", id, at, in)
		}
	}

	for i := 8; i < 28; i++ { // bursts never enter
		if at, ok := firstIn[fmt.Sprintf("i%04d", i)]; ok {
			t.Fatalf("burst item i%04d entered the trend at %v", i, at)
		}
	}
}
