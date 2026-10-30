//go:build integration

package tests

import (
	"net/http"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestLimitsSurviveRestart: budgets live in the shared database cache (SDD 13 §7.1), so a new process (a
// deploy or a crash) does not hand the client a fresh login budget.
func TestLimitsSurviveRestart(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"DATABASE_URL":     pgtest.Database(t, ""),
		"RATE_LIMIT_STORE": "cache",
		"CACHE_STORE":      "database",
	}
	login := map[string]any{"identifier": "nobody@example.com", "password": "wrong-password-123"}

	before := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})
	c := before.Client(t)

	for i := range 5 {
		if code := c.PostJSON("/login", login).Status(); code == http.StatusTooManyRequests {
			t.Fatalf("attempt %d already limited", i+1)
		}
	}

	after := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})
	res := after.Client(t).PostJSON("/login", login).AssertStatus(http.StatusTooManyRequests).
		AssertHeader("RateLimit-Remaining", "0")
	if res.Header("Retry-After") == "" {
		t.Error("429 without Retry-After")
	}
}
