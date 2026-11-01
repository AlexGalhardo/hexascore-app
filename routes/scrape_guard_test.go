package routes

import (
	"testing"

	"github.com/AlexGalhardo/vukem/auth"
)

type stubUser struct{}

func (stubUser) AuthID() string { return "1" }

// The scrape guard's signed-in check reads the "auth_web" session key directly (VF-065); this pins it to
// what vukem's web session guard writes on login.
func TestScrapeGuardSignedInKeyMatchesVukem(t *testing.T) {
	if _, ok := auth.SessionValues("web", stubUser{})["auth_web"]; !ok {
		t.Fatalf("auth.SessionValues(\"web\", u) = %v, want an auth_web key", auth.SessionValues("web", stubUser{}))
	}
}
