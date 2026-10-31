package tests

import (
	"strings"
	"testing"
)

// TestCSRSkeleton: pages that skip SSR still ship a title and a busy placeholder in the HTML (P2.H.2.1).
func TestCSRSkeleton(t *testing.T) {
	t.Parallel()

	r := newKeyRig(t)
	c := r.app.Client(t)
	register(t, r.app, c, "skeleton@example.com")

	for _, tc := range []struct{ path, title string }{
		{"/profile", "Profile"},
		{"/profile/api", "API keys and usage"},
		{"/billing", "Plan and billing"},
		{"/me/notifications", "Notifications"},
		{"/library", "Your library"},
		{"/lists", "Your lists"},
	} {
		path, title := tc.path, tc.title
		body := c.Get(path).AssertOK().Body()
		if n := strings.Count(body, `aria-busy="true"`); n != 1 {
			t.Errorf("%s: aria-busy placeholders = %d, want 1", path, n)
		}

		if !strings.Contains(body, "<title>"+title+"</title>") {
			t.Errorf("%s: missing <title>%s</title>", path, title)
		}
	}

	if body := c.Get("/").AssertOK().Body(); strings.Contains(body, `aria-busy="true"`) {
		t.Error("home must not carry the CSR placeholder")
	}
}
