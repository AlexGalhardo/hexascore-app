package backup

import (
	"slices"
	"testing"
	"time"
)

func TestPruneSelection(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)

	keys := func(kind string, n int, step func(time.Time, int) time.Time) []string {
		var out []string
		for i := range n {
			out = append(out, ObjectKey(kind, "production", step(start, i)))
		}
		// List order is not trusted: shuffle by reversing.
		slices.Reverse(out)

		return out
	}

	daily := keys("daily", 40, func(t time.Time, i int) time.Time { return t.AddDate(0, 0, i) })
	monthly := keys("monthly", 15, func(t time.Time, i int) time.Time { return t.AddDate(0, i, 0) })

	for _, c := range []struct {
		name       string
		keys       []string
		keep, want int
		oldestKept time.Time
	}{
		{"daily", daily, Retention["daily"], 10, start.AddDate(0, 0, 10)},
		{"monthly", monthly, Retention["monthly"], 3, start.AddDate(0, 3, 0)},
		{"few", daily[:5], Retention["daily"], 0, time.Time{}},
	} {
		del := PruneSelection(c.keys, c.keep)
		if len(del) != c.want {
			t.Fatalf("%s: deleted %d, want %d", c.name, len(del), c.want)
		}

		if c.want == 0 {
			continue
		}

		kind := c.name
		if slices.Contains(del, ObjectKey(kind, "production", c.oldestKept)) {
			t.Errorf("%s: deleted the oldest dump that must be kept", c.name)
		}

		if !slices.Contains(del, ObjectKey(kind, "production", c.oldestKept.AddDate(0, 0, -1))) &&
			!slices.Contains(del, ObjectKey(kind, "production", c.oldestKept.AddDate(0, -1, 0))) {
			t.Errorf("%s: kept a dump older than the retention", c.name)
		}
	}
}
