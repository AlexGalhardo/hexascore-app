// Package problems is the catalogue of public API problem codes (SDD 05 §6.2).
package problems

import "net/http"

// DocsPath is where every problem type URL redirects, with the code as fragment.
const DocsPath = "/docs/errors"

// Problem is one catalogue entry: the kebab-case code and its HTTP status.
type Problem struct {
	Code   string
	Status int
}

// Catalog lists every code of SDD 05 §6.2. Several codes may share a status;
// ForStatus picks the first.
var Catalog = []Problem{
	{"unauthorized", http.StatusUnauthorized},
	{"invalid-query", http.StatusBadRequest},
	{"invalid-cursor", http.StatusBadRequest},
	{"changes-window-exceeded", http.StatusBadRequest},
	{"quota-exhausted", http.StatusPaymentRequired},
	{"plan-required", http.StatusForbidden},
	{"account-banned", http.StatusForbidden},
	{"not-found", http.StatusNotFound},
	{"rate-limited", http.StatusTooManyRequests},
	{"internal-error", http.StatusInternalServerError},
	{"service-unavailable", http.StatusServiceUnavailable},
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
		if p.Status == status {
			return p.Code
		}
	}

	return ""
}
