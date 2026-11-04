package catalog

import (
	"context"

	"github.com/AlexGalhardo/vukem/flags"
)

// HideUncoveredFlag keeps items without a cover image off the landing page and the four category pages (owner
// 2026-11-04). Env FLAG_HIDE_UNCOVERED. The item page itself, search and the public API are not affected.
var HideUncoveredFlag = flags.Bool("hide_uncovered", true).
	Describe("Hide items without a cover from the landing page and the category listings")

// HideUnscoredFlag keeps items without a visible score (fewer than 5 ratings) off the landing page and the four
// category pages; they come back as soon as they have one (owner 2026-11-04). Env FLAG_HIDE_UNSCORED.
var HideUnscoredFlag = flags.Bool("hide_unscored", true).
	Describe("Hide items without a visible score from the landing page and the category listings")

// Showcase says which items the public shop windows (landing, category pages) may show.
type Showcase struct{ Cover, Score bool }

// ShowcaseFrom reads the two flags. Unreadable flags fall back to showing everything.
func ShowcaseFrom(ctx context.Context, ff *flags.Flags) Showcase {
	if ff == nil {
		return Showcase{}
	}

	return Showcase{Cover: ff.Enabled(ctx, HideUncoveredFlag), Score: ff.Enabled(ctx, HideUnscoredFlag)}
}

// Keep reports whether a loaded card passes the showcase.
func (s Showcase) Keep(h Highlight) bool {
	return (!s.Cover || h.Cover != nil) && (!s.Score || h.Score != nil)
}

// Filter drops the cards that fail the showcase, keeping the order.
func (s Showcase) Filter(in []Highlight) []Highlight {
	if !s.Cover && !s.Score {
		return in
	}

	out := make([]Highlight, 0, len(in))

	for _, h := range in {
		if s.Keep(h) {
			out = append(out, h)
		}
	}

	return out
}

// where is the SQL condition for items aliased i; empty when nothing is hidden.
func (s Showcase) where() string {
	cond := ""

	if s.Cover {
		cond += " AND coalesce(i.cover_key, '') <> ''"
	}

	if s.Score {
		cond += " AND " + visibleScore
	}

	return cond
}
