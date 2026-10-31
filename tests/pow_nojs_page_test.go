package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/cache"
)

// TestSecurity_A07_NoJS428Page: a plain browser form post past the PoW threshold gets an HTML 428 page
// that offers Google sign-in; script and API clients keep the problem+json body. Both stay fail closed.
func TestSecurity_A07_NoJS428Page(t *testing.T) {
	t.Parallel()

	store := &ttlStore{Memory: cache.NewMemory(0), ttls: map[string]time.Duration{}}
	h := testChallenger(store, nil).Require("signup")(okHandler)

	past := func(mutate func(*http.Request)) *httptest.ResponseRecorder {
		for range 3 {
			h.ServeHTTP(httptest.NewRecorder(), challengeFrom("198.51.100.7", ""))
		}

		r := challengeFrom("198.51.100.7", "")
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		mutate(r)

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)

		return rec
	}

	html := past(func(r *http.Request) { r.Header.Set("Accept", "text/html,application/xhtml+xml,*/*;q=0.8") })
	if html.Code != http.StatusPreconditionRequired {
		t.Fatalf("status %d, want 428", html.Code)
	}

	if ct := html.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("content-type %q, want text/html", ct)
	}

	if body := html.Body.String(); !strings.Contains(body, `href="/auth/google/redirect"`) || strings.Contains(body, "<script") {
		t.Fatalf("page must link Google sign-in and carry no script: %s", body)
	}

	store2 := &ttlStore{Memory: cache.NewMemory(0), ttls: map[string]time.Duration{}}
	h = testChallenger(store2, nil).Require("signup")(okHandler)

	for name, mutate := range map[string]func(*http.Request){
		"json accept": func(r *http.Request) { r.Header.Set("Accept", "application/json") },
		"inertia": func(r *http.Request) {
			r.Header.Set("Accept", "text/html, application/xhtml+xml")
			r.Header.Set("X-Inertia", "true")
		},
		"json body": func(r *http.Request) {
			r.Header.Set("Accept", "text/html")
			r.Header.Set("Content-Type", "application/json")
		},
	} {
		rec := past(mutate)
		if rec.Code != http.StatusPreconditionRequired || !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/problem+json") {
			t.Errorf("%s: status %d content-type %q, want 428 problem+json", name, rec.Code, rec.Header().Get("Content-Type"))
		}
	}
}
