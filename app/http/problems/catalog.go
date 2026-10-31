// Package problems is the catalogue of public API problem codes (SDD 05 §6.2).
package problems

import "net/http"

// DocsPath is where every problem type URL redirects, with the code as fragment.
const DocsPath = "/docs/errors"

// Problem is one catalogue entry: the kebab-case code and its HTTP status.
type Problem struct {
	Code   string
	Status int
	// Specific codes are raised explicitly by a handler; ForStatus never picks them.
	Specific bool
}

// Catalog lists every code of SDD 05 §6.2. Several codes may share a status;
// ForStatus picks the first.
var Catalog = []Problem{
	{"unauthorized", http.StatusUnauthorized, false},
	{"invalid-query", http.StatusBadRequest, false},
	{"invalid-cursor", http.StatusBadRequest, false},
	{"changes-window-exceeded", http.StatusBadRequest, false},
	{"quota-exhausted", http.StatusPaymentRequired, false},
	{"plan-required", http.StatusForbidden, false},
	{"account-banned", http.StatusForbidden, false},
	{"not-found", http.StatusNotFound, false},
	{"rate-limited", http.StatusTooManyRequests, false},
	{"internal-error", http.StatusInternalServerError, false},
	{"service-unavailable", http.StatusServiceUnavailable, false},
	// Community edits (SDD 04), served on /app/v1 and the community pages.
	{"community-submissions-closed", http.StatusForbidden, true},
	{"suggestion-not-eligible", http.StatusForbidden, true},
	{"suggestion-duplicate", http.StatusConflict, true},
	{"stale-before", http.StatusConflict, true},
	{"invalid-transition", http.StatusConflict, true},
	{"field-locked", http.StatusLocked, true},
	{"suggestion-quota", http.StatusTooManyRequests, true},
	{"suggestion-cooldown", http.StatusTooManyRequests, true},
	{"validation", http.StatusUnprocessableEntity, true},
}

// Known reports whether code is in the catalogue.
func Known(code string) bool {
	for _, p := range Catalog {
		if p.Code == code {
			return true
		}
	}

	return false
}

// ForStatus is the default code for a status, used when a problem carries none.
// Statuses outside the catalogue map to "" (the problem is left as it is).
func ForStatus(status int) string {
	switch status {
	case http.StatusMethodNotAllowed, http.StatusNotAcceptable, http.StatusUnsupportedMediaType:
		return "invalid-query"
	case http.StatusBadGateway, http.StatusGatewayTimeout:
		return "service-unavailable"
	}

	for _, p := range Catalog {
		if p.Status == status && !p.Specific {
			return p.Code
		}
	}

	return ""
}
