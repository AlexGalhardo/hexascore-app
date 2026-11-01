package related

import (
	"context"
	"crypto/rand"
	"log/slog"
	"math/big"
)

// Events writes the item-page recommendation events as structured INFO lines (SDD 08 section 8, D-35): Railway log
// search counts `event=reco_impression` and `event=reco_click` per surface and type, since the app exports no metrics.
type Events struct {
	Log *slog.Logger
	// ImpressionEvery logs one impression in N (0 or 1: all). Each line carries sample_rate=N, so a count times
	// sample_rate estimates the real volume. Clicks are rare and always logged.
	ImpressionEvery int
	pick            func(n int) int
}

// Impression logs one served block, sampled by ImpressionEvery.
func (e Events) Impression(ctx context.Context, surface, typ, itemID string) {
	n := max(e.ImpressionEvery, 1)
	pick := e.pick
	if pick == nil {
		pick = cryptoPick
	}

	if n > 1 && pick(n) != 0 {
		return
	}

	e.Log.InfoContext(ctx, "reco impression", "event", "reco_impression",
		"surface", surface, "type", typ, "item", itemID, "sample_rate", n)
}

// Click logs one visit that arrived through the block; surfaces other than similar and crossmedia are ignored.
func (e Events) Click(ctx context.Context, surface, typ, itemID string) {
	if FromSurface(surface) == "" {
		return
	}

	e.Log.InfoContext(ctx, "reco click", "event", "reco_click", "surface", surface, "type", typ, "item", itemID)
}

// FromSurface returns the ?from= value when it names an item-page block, else "".
func FromSurface(from string) string {
	if from == "similar" || from == "crossmedia" {
		return from
	}

	return ""
}

// cryptoPick is a uniform number in [0, n); a failing entropy source keeps the line (0), never drops it.
func cryptoPick(n int) int {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}

	return int(v.Int64())
}
