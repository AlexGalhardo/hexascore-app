package limits_test

import (
	"errors"
	"testing"
	"testing/fstest"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/ratelimit"

	"github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
)

func appWith(t *testing.T, env map[string]string) *vukem.App {
	t.Helper()

	app, err := vukem.New(vukem.WithEnv(fstest.MapFS{}, func(k string) (string, bool) {
		v, ok := env[k]
		return v, ok
	}))
	if err != nil {
		t.Fatal(err)
	}

	return app
}

// TestSecurity_A06_ProductionRefusesMemoryLimits: per-process budgets in production are a boot error.
func TestSecurity_A06_ProductionRefusesMemoryLimits(t *testing.T) {
	t.Parallel()

	key := "base64:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

	for _, store := range []string{"", "memory"} {
		app := appWith(t, map[string]string{"APP_ENV": "production", "APP_KEY": key, "RATE_LIMIT_STORE": store})
		if _, err := limits.Factory(app); !errors.Is(err, limits.ErrMemoryStore) {
			t.Errorf("production with RATE_LIMIT_STORE=%q: err = %v, want ErrMemoryStore", store, err)
		}
	}

	if _, err := limits.Factory(appWith(t, map[string]string{"APP_ENV": "development"})); err != nil {
		t.Errorf("local memory store: %v", err)
	}
}

// TestAPIRateLimitFreeBurst: the free plan allows 5 immediate requests (Q-A2, D-33 review) and then
// sustains 1 request/second: the 6th in the same instant is denied with a Retry-After of at most one second.
func TestAPIRateLimitFreeBurst(t *testing.T) {
	t.Parallel()

	p, _ := api.PlanByID(api.PlanFree)
	factory := func(_ string, limit int, window time.Duration) ratelimit.Limiter {
		return ratelimit.NewMemory(limit, window)
	}
	l := limits.APIPlan(api.PlanFree, p.RPS, p.Burst).From(factory)

	for i := range 5 {
		if r, err := l.Allow(t.Context(), "u"); err != nil || !r.Allowed {
			t.Fatalf("request %d: %+v, %v", i+1, r, err)
		}
	}

	r, err := l.Allow(t.Context(), "u")
	if err != nil || r.Allowed || r.RetryAfter <= 0 || r.RetryAfter > time.Second {
		t.Fatalf("6th request: %+v, %v; want denied, Retry-After <= 1s", r, err)
	}
}
