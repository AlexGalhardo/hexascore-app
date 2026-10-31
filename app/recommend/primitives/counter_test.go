package primitives

import (
	"strconv"
	"testing"
)

// TestCounterDropsOldestAbove100k: above the cap the oldest keys go and the drop is counted (SDD 08 section 8).
func TestCounterDropsOldestAbove100k(t *testing.T) {
	dropped := 0
	c := NewCounter(MaxKeys, func(n int) { dropped += n })

	for i := range MaxKeys + 1 {
		c.Add(Key{Type: "game", Ref: strconv.Itoa(i), Day: "2026-10-31"}, 1, 0)
	}

	if c.Len() != MaxKeys || dropped != 1 || c.Dropped() != 1 {
		t.Fatalf("len=%d dropped=%d total=%d, want %d/1/1", c.Len(), dropped, c.Dropped(), MaxKeys)
	}

	got := c.Swap()
	if _, ok := got[Key{Type: "game", Ref: "0", Day: "2026-10-31"}]; ok {
		t.Fatal("the oldest key survived")
	}
}

// TestCounterMergeBackKeepsTotals: a failed flush merges into the counts that arrived meanwhile.
func TestCounterMergeBackKeepsTotals(t *testing.T) {
	c := NewCounter(10, nil)
	k := Key{Type: "game", Ref: "zelda", Day: "2026-10-31"}
	c.Add(k, 2, 1)

	failed := c.Swap()
	c.Add(k, 3, 0)
	c.Merge(failed)

	if got := c.Swap()[k]; got != (Tally{Views: 5, Clicks: 1}) {
		t.Fatalf("got %+v", got)
	}
}
