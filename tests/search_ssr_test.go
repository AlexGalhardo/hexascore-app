//go:build integration

package tests

import "testing"

// TestSearchFirstResponseHasResults: the first HTML response carries the result links (W-05), whether the page was
// server-rendered (a checkout with a built bootstrap/ssr bundle) or fell back to the client shell (no bundle, as in
// CI); the <noscript> list itself is covered by TestSearchShell.
func TestSearchFirstResponseHasResults(t *testing.T) {
	app, _ := searchApp(t)

	app.Client(t).Get("/search?q=zelda&type=game").AssertOK().
		AssertSee(`href="/games/game-0"`)
}
