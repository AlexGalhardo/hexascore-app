// Package ssr observes Inertia server-side rendering health.
package ssr

import (
	"context"
	"log/slog"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/AlexGalhardo/vukem/inertia"

	"github.com/AlexGalhardo/hexascore-app/app/alerts"
)

// PrivatePrefixes are the route prefixes that skip SSR (SDD 09 §6.1); a fallback there is not an SEO loss and
// never raises A5.
var PrivatePrefixes = []string{"/admin", "/profile", "/me", "/billing", "/suggest", "/suggestions"}

// Failures counts renders that fell back to client-side rendering. vukem's
// default OnFailure only logs (VF-002), so the count is what alerts read.
type Failures struct {
	// Log receives the fallback lines; nil uses slog.Default.
	Log *slog.Logger

	n atomic.Int64
}

// Observe is an inertia.SSRConfig.OnFailure hook. A fallback on a public route logs alert A5 at ERROR (SDD 12
// FR-12-3); on a private one it is a plain WARN without the alert key.
func (f *Failures) Observe(ctx context.Context, err inertia.SSRError) {
	f.n.Add(1)

	log := f.Log
	if log == nil {
		log = slog.Default()
	}

	public := isPublic(err.URL)
	if public {
		alerts.Error(ctx, log, alerts.A5, "inertia ssr fallback", "public", true,
			"type", err.Type, "component", err.Component, "error", err.Error)
		return
	}

	log.WarnContext(ctx, "inertia ssr fallback", "public", false,
		"type", err.Type, "component", err.Component, "error", err.Error)
}

func isPublic(raw string) bool {
	path := raw
	if u, err := url.Parse(raw); err == nil {
		path = u.Path
	}

	for _, p := range PrivatePrefixes {
		if path == p || strings.HasPrefix(path, p+"/") {
			return false
		}
	}

	return true
}

// Count returns the number of fallbacks since boot.
func (f *Failures) Count() int64 {
	return f.n.Load()
}
