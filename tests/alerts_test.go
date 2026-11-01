//go:build integration

package tests

import (
	"bytes"
	"log/slog"
	"strings"
	"sync"
	"testing"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/ssr"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.b.String()
}

// TestAlert_A5_SSRFallbackPublicRoute: with the SSR renderer unreachable, a public page still renders (client
// side) and the fallback logs alert=A5 at ERROR; a private page falls back without the alert (SDD 12 FR-12-3).
func TestAlert_A5_SSRFallbackPublicRoute(t *testing.T) {
	t.Parallel()

	var logs lockedBuffer

	failures := &ssr.Failures{Log: slog.New(slog.NewTextHandler(&logs, nil))}

	// Port 1 refuses connections; no Bundle, so the renderer is attempted. Same Except and OnFailure as bootstrap.
	dead := vukem.WithInertia(inertia.Config{Version: func() string { return "test" }, SSR: &inertia.SSRConfig{
		URL:       "http://127.0.0.1:1",
		OnFailure: failures.Observe,
		Except:    ssr.PrivatePrefixes,
	}})

	app := vukemtest.New(t, func(opts ...vukem.Option) (*vukem.App, error) {
		return bootstrap.NewWithSSRFailures(failures, opts...)
	}, vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}, AppOptions: []vukem.Option{dead}})

	c := app.Client(t)
	c.Get("/").AssertOK()

	if failures.Count() == 0 {
		t.Fatalf("no SSR fallback observed; the test cannot exercise A5 (logs: %s)", logs.String())
	}

	out := logs.String()
	if !strings.Contains(out, "alert=A5") || !strings.Contains(out, "level=ERROR") {
		t.Fatalf("public SSR fallback must log alert=A5 at ERROR, got: %s", out)
	}
}
