//go:build integration

package tests

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/httpclient"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/httpx"
)

type handClock struct{ now time.Time }

func (c *handClock) Now() time.Time { return c.now }

func (c *handClock) Sleep(ctx context.Context, d time.Duration) error {
	c.now = c.now.Add(d)

	return ctx.Err()
}

type status int

func (s status) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: int(s),
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader("x")),
		Request:    req,
	}, nil
}

func ingestDB(t *testing.T) *vukemtest.App {
	t.Helper()

	return vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)
}

// The worker's gate reads the web's signal from the table: it waits while in-flight > 16 or p95 > 300 ms, goes on
// when the signal quiets, and ignores a stale or missing one.
func TestIngestPressure_PausesAboveThresholds(t *testing.T) {
	app := ingestDB(t)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	clk := &handClock{now: time.Now()}
	sleeps := 0
	gate := &ingest.SignalGate{DB: db, Now: func() time.Time { return clk.now }}
	gate.Sleep = func(ctx context.Context, d time.Duration) error {
		sleeps++
		clk.now = clk.now.Add(d)

		if sleeps == 3 { // the site calms down
			return ingest.WriteSignal(ctx, db, ingest.Signal{InFlight: 2, P95MS: 40, At: clk.now})
		}

		return nil
	}

	for name, busy := range map[string]ingest.Signal{
		"inflight": {InFlight: 17, P95MS: 40},
		"p95":      {InFlight: 2, P95MS: 301},
	} {
		sleeps = 0
		busy.At = clk.now

		if err := ingest.WriteSignal(t.Context(), db, busy); err != nil {
			t.Fatal(err)
		}

		if err := gate.Wait(t.Context()); err != nil {
			t.Fatal(err)
		}

		if sleeps != 3 {
			t.Errorf("%s: slept %d times, want 3", name, sleeps)
		}
	}

	sleeps = 0

	if err := ingest.WriteSignal(
		t.Context(),
		db,
		ingest.Signal{InFlight: 99, P95MS: 900, At: clk.now.Add(-ingest.StaleAfter - time.Minute)},
	); err != nil {
		t.Fatal(err)
	}

	if err := gate.Wait(t.Context()); err != nil || sleeps != 0 {
		t.Errorf("stale signal: err %v, slept %d times, want no wait", err, sleeps)
	}
}

// An open breaker is stored when it opens and a fresh client built from the table (a restart) still refuses.
func TestBreakerStateSurvivesRestart(t *testing.T) {
	app := ingestDB(t)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	clk := &handClock{now: time.Now()}
	build := func(code int) *httpx.Client {
		return httpx.New(httpx.Config{
			ContactURL: "https://hexascore.app",
			Hosts:      map[string]httpx.HostLimit{"api.test": {PerSecond: 100}},
			Clock:      clk, Base: status(code), State: ingest.HostStore{DB: db},
		})
	}

	first := build(http.StatusInternalServerError)

	for range 12 {
		req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.test/x", nil)
		if resp, err := first.JSON.Do(req); err == nil {
			_ = resp.Body.Close()
		}
	}

	second := build(http.StatusOK)
	if err := second.Restore(t.Context()); err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodGet, "https://api.test/x", nil)
	_, err = second.JSON.Do(req) //nolint:bodyclose // the breaker refuses before any response exists
	if !errors.Is(err, httpclient.ErrCircuitOpen) {
		t.Fatalf("after restart err = %v, want ErrCircuitOpen", err)
	}

	clk.now = clk.now.Add(6 * time.Minute)

	resp, err := second.JSON.Do(req)
	if err != nil {
		t.Fatalf("probe after the open period: %v", err)
	}

	_ = resp.Body.Close()
}
