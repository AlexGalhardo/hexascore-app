// Package ssr observes Inertia server-side rendering health.
package ssr

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/AlexGalhardo/vukem/inertia"
)

// Failures counts renders that fell back to client-side rendering. vukem's
// default OnFailure only logs (VF-002), so the count is what alerts read.
type Failures struct {
	n atomic.Int64
}

// Observe is an inertia.SSRConfig.OnFailure hook.
func (f *Failures) Observe(ctx context.Context, err inertia.SSRError) {
	f.n.Add(1)
	slog.WarnContext(ctx, "inertia ssr fallback", "type", err.Type, "component", err.Component, "error", err.Error)
}

// Count returns the number of fallbacks since boot.
func (f *Failures) Count() int64 {
	return f.n.Load()
}
