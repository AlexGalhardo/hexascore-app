//go:build integration

package tests

import (
	"testing"
)

// TestSearchTimeoutDegrades holds an exclusive lock on item_translations so the
// item query hits the 800 ms statement timeout.
func TestSearchTimeoutDegrades(t *testing.T) {
	app, _ := searchApp(t)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.SQL().BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = tx.Rollback() })

	if _, err := tx.ExecContext(t.Context(), `LOCK TABLE item_translations IN ACCESS EXCLUSIVE MODE`); err != nil {
		t.Fatal(err)
	}

	t.Run("html degrades to an empty page with a notice", func(t *testing.T) {
		app.Client(t).Inertia("").Get("/search?q=zelda").AssertOK().AssertInertia("Search/Index").
			PropEquals("unavailable", true).
			HasProp("groups", 0).
			HasProp("results", 0)
	})

	t.Run("html response asks the client to retry", func(t *testing.T) {
		app.Client(t).Inertia("").Get("/search?q=zelda").AssertHeader("Retry-After", "5")
	})

	t.Run("app json is a 503 problem with Retry-After", func(t *testing.T) {
		app.Client(t).GetJSON("/app/v1/search?q=zelda").AssertProblem(503).AssertHeader("Retry-After", "5")
	})
}
