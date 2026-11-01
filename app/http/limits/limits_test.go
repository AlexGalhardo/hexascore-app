package limits_test

import (
	"errors"
	"testing"
	"testing/fstest"

	"github.com/AlexGalhardo/vukem"

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
