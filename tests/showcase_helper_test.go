//go:build integration

package tests

// showEverythingEnv is the test environment for fixtures that have no cover or score: it turns the landing and
// category showcase filters off.
func showEverythingEnv(databaseURL string) map[string]string {
	return map[string]string{
		"DATABASE_URL":        databaseURL,
		"FLAG_HIDE_UNCOVERED": "false",
	}
}
