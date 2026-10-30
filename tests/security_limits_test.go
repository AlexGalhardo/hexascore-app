package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/ratelimit"

	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
)

func memoryLimiters(_ string, limit int, window time.Duration) ratelimit.Limiter {
	return ratelimit.NewMemory(limit, window)
}

func limited(l ratelimit.Limiter) http.Handler {
	return ratelimit.Middleware(l, ratelimit.ByIP)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
}

// TestSecurity_A06_HexaLimits walks every row of SDD 13 §7.1: the budget passes, the next request gets
// 429 with Retry-After and RateLimit-Remaining 0, names are unique, and no route builds its own limiter.
func TestSecurity_A06_HexaLimits(t *testing.T) {
	t.Parallel()

	names := map[string]bool{}

	for _, rule := range limits.All() {
		t.Run(rule.Name, func(t *testing.T) {
			t.Parallel()

			h := limited(rule.From(memoryLimiters))

			for i := range rule.Limit {
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))

				if rec.Code != http.StatusOK {
					t.Fatalf("request %d of %d: status %d", i+1, rule.Limit, rec.Code)
				}
			}

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))

			if rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") == "" ||
				rec.Header().Get("RateLimit-Remaining") != "0" {
				t.Fatalf("over budget: status %d, Retry-After %q, RateLimit-Remaining %q", rec.Code,
					rec.Header().Get("Retry-After"), rec.Header().Get("RateLimit-Remaining"))
			}
		})

		if names[rule.Name] {
			t.Errorf("rule name %q is used twice: the rows would share buckets", rule.Name)
		}

		names[rule.Name] = true
	}

	files, err := filepath.Glob("../routes/*.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("routes: %v", err)
	}

	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}

		for _, bad := range []string{`limiters("`, "app.RateLimiters()", "ratelimit.NewMemory"} {
			if strings.Contains(string(src), bad) {
				t.Errorf("%s uses %s: take the rule from app/http/limits", f, bad)
			}
		}
	}
}

type brokenLimiter struct{}

func (brokenLimiter) Allow(context.Context, string) (ratelimit.Result, error) {
	return ratelimit.Result{}, errors.New("store down")
}

// TestSecurity_A06_HexaLimitsFailClosed: when the store fails, every limited route answers 503.
func TestSecurity_A06_HexaLimitsFailClosed(t *testing.T) {
	t.Parallel()

	broken := func(string, int, time.Duration) ratelimit.Limiter { return brokenLimiter{} }

	for _, rule := range limits.All() {
		rec := httptest.NewRecorder()
		limited(rule.From(broken)).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))

		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("%s with a failing store: status %d, want 503", rule.Name, rec.Code)
		}
	}
}
