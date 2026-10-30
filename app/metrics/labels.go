// Package metrics registers the platform instruments from SDD 12 §6.2.
package metrics

import (
	"log/slog"
	"strings"
	"sync"
)

// ErrorTypeCap is the number of distinct error.type values kept before other (SDD 12 §6.2).
const ErrorTypeCap = 50

// Cap keeps the first max distinct values and maps the rest to other.
type Cap struct {
	max  int
	seen map[string]struct{}
	mu   sync.Mutex
	once sync.Once
}

// NewCap returns a cap that keeps max distinct values.
func NewCap(limit int) *Cap {
	return &Cap{max: limit, seen: make(map[string]struct{}, limit)}
}

// Observe returns v when it is already known or still fits, and other after the cap.
func (c *Cap) Observe(v string) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.seen[v]; ok {
		return v
	}

	if len(c.seen) >= c.max {
		c.once.Do(func() {
			slog.Warn("metric label cap reached", "cap", c.max)
		})

		return "other"
	}

	c.seen[v] = struct{}{}

	return v
}

// RouteGroup maps an http.route pattern to the SDD 12 §6.2 group.
// Groups are derived at read time; they are not metric labels.
func RouteGroup(route string) string {
	route = strings.TrimSpace(route)
	if i := strings.IndexByte(route, ' '); i > 0 && !strings.Contains(route[:i], "/") {
		route = route[i+1:]
	}

	if i := strings.IndexByte(route, '/'); i > 0 {
		route = route[i:]
	}

	switch {
	case route == "/up" || strings.HasPrefix(route, "/up/"):
		return "health"
	case strings.HasPrefix(route, "/search"), strings.HasPrefix(route, "/api/search"):
		return "search"
	case strings.HasPrefix(route, "/admin"):
		return "admin"
	case strings.HasPrefix(route, "/api/v1"), strings.HasPrefix(route, "/v1/"):
		return "api.v1"
	case strings.HasPrefix(route, "/app/v1"):
		return "app.v1"
	case isAppRoute(route):
		return "web.app"
	default:
		return "web.public"
	}
}

// PublicPage reports whether a fallback on route should page (SDD 12 §8).
func PublicPage(route string) bool {
	switch RouteGroup(route) {
	case "web.app", "admin", "api.v1", "app.v1":
		return false
	default:
		return true
	}
}

func isAppRoute(route string) bool {
	for _, p := range []string{
		"/dashboard", "/profile", "/me", "/billing", "/onboarding",
		"/lists", "/notifications", "/library", "/settings",
	} {
		if route == p || strings.HasPrefix(route, p+"/") {
			return true
		}
	}

	return false
}
