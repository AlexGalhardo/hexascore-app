package v1

import (
	"errors"
	"net/http"
	"time"

	vapi "github.com/AlexGalhardo/vukem/api"
)

// MinSunset is the shortest gap between Deprecation and Sunset (SDD 05 §6.5).
// A shorter window is rejected. 183 days is the six-month floor the policy uses.
const MinSunset = 183 * 24 * time.Hour

// ErrSunsetTooSoon means Sunset is less than MinSunset after Deprecation.
var ErrSunsetTooSoon = errors.New("api: sunset is under 183 days after deprecation")

// Deprecate checks the sunset window and wraps the handler with vukem
// api.Deprecate (Deprecation, Sunset, Link). No v1 route is deprecated;
// callers mount the handler only on a route that is actually going away.
func Deprecate(since, sunset time.Time, link string, next http.Handler) (http.Handler, error) {
	if sunset.Sub(since) < MinSunset {
		return nil, ErrSunsetTooSoon
	}

	return vapi.Deprecate(since, sunset, link)(next), nil
}
