//go:build integration

package tests

import "testing"

// TestInertiaPages_Search runs under -run 'TestInertiaPages/Search' because the
// parent name matches and the subtests start with "Search".
func TestInertiaPages_Search(t *testing.T) {
	app, _ := searchApp(t)

	t.Run("Search/grouped", func(t *testing.T) {
		app.Client(t).Inertia("").Get("/search?q=zel").AssertOK().AssertInertia("Search/Index").
			PropEquals("query.type", "").
			HasProp("groups", 1).
			HasProp("results", 0)
	})

	t.Run("Search/typed", func(t *testing.T) {
		app.Client(t).Inertia("").Get("/search?q=zelda&type=game").AssertOK().AssertInertia("Search/Index").
			PropEquals("query.type", "game").
			HasProp("groups", 0).
			HasProp("results", 3)
	})
}
