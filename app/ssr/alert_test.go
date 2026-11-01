package ssr

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/inertia"
)

func TestAlert_SSRFallbackPublic(t *testing.T) {
	cases := []struct {
		url   string
		alert bool
	}{
		{"/items/the-witcher", true},
		{"/", true},
		{"/profile", false},
		{"/admin/users?x=1", false},
		{"/profiles/ana", true},
	}

	for _, c := range cases {
		var b bytes.Buffer
		f := &Failures{Log: slog.New(slog.NewTextHandler(&b, nil))}
		f.Observe(context.Background(), inertia.SSRError{Type: "connection", Component: "Home", URL: c.url})

		got := strings.Contains(b.String(), "alert=A5")
		if got != c.alert {
			t.Errorf("%s: alert=%v, want %v (%s)", c.url, got, c.alert, b.String())
		}
		if got && !strings.Contains(b.String(), "level=ERROR") {
			t.Errorf("%s: alert must log at ERROR: %s", c.url, b.String())
		}
		if strings.Contains(b.String(), c.url) && c.url != "/" {
			t.Errorf("%s: raw path logged: %s", c.url, b.String())
		}
		if f.Count() != 1 {
			t.Errorf("%s: Count() = %d, want 1", c.url, f.Count())
		}
	}
}
