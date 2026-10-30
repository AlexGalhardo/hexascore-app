package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	apiv1 "github.com/AlexGalhardo/hexascore-app/app/api/v1"
)

// TestAPIDeprecationHeaders checks the RFC headers on a test-only route and
// rejects a sunset that is under 183 days after Deprecation. No v1 route is
// deprecated, so this handler is not mounted on the app router.
func TestAPIDeprecationHeaders(t *testing.T) {
	t.Parallel()

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	sunset := since.Add(apiv1.MinSunset)
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	h, err := apiv1.Deprecate(since, sunset, "/docs/v2", next)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/fixture-deprecated", http.NoBody))

	if !called || rec.Code != http.StatusNoContent {
		t.Fatalf("fixture status = %d called = %v", rec.Code, called)
	}

	got := rec.Header()
	wantDep := "@" + strconv.FormatInt(since.Unix(), 10)
	wantSunset := sunset.UTC().Format(http.TimeFormat)
	wantLink := `</docs/v2>; rel="deprecation"; type="text/html"`
	if got.Get("Deprecation") != wantDep || got.Get("Sunset") != wantSunset || got.Get("Link") != wantLink {
		t.Fatalf("Deprecation=%q Sunset=%q Link=%q", got.Get("Deprecation"), got.Get("Sunset"), got.Get("Link"))
	}

	short := since.Add(apiv1.MinSunset - time.Second)
	if _, err := apiv1.Deprecate(since, short, "/docs/v2", next); !errors.Is(err, apiv1.ErrSunsetTooSoon) {
		t.Fatalf("short window err = %v", err)
	}
}
