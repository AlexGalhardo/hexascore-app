//go:build integration

package tests

import "testing"

// TestSearchFirstResponseHasResults: the tests run without an SSR bundle, so /search takes the client-shell path;
// the first HTML response must still carry the result links (W-05).
func TestSearchFirstResponseHasResults(t *testing.T) {
	app, _ := searchApp(t)

	app.Client(t).Get("/search?q=zelda&type=game").AssertOK().
		AssertSee(`<noscript>`).
		AssertSee(`href="/games/game-0"`)
}
