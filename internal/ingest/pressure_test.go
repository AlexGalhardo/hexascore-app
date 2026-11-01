package ingest

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time { return c.t }

func newTestPressure(c *fakeClock, onSleep func()) *Pressure {
	p := NewPressure(c.now)
	p.Sleep = func(_ context.Context, d time.Duration) error {
		c.t = c.t.Add(d)
		if onSleep != nil {
			onSleep()
		}

		return nil
	}

	return p
}

// A batch waits while more than 16 requests are in flight and goes on at 16.
func TestPressureGateWaits(t *testing.T) {
	c := &fakeClock{t: time.Unix(1000, 0)}

	var finish []func(bool)

	p := newTestPressure(c, func() {
		if len(finish) > 16 {
			finish[len(finish)-1](false)
			finish = finish[:len(finish)-1]
		}
	})

	for range 17 {
		finish = append(finish, p.Begin())
	}

	if err := p.Wait(t.Context()); err != nil {
		t.Fatal(err)
	}

	if got := p.InFlight(); got != 16 {
		t.Errorf("in flight after the wait = %d, want 16", got)
	}

	if c.t.Equal(time.Unix(1000, 0)) {
		t.Error("the gate did not wait at 17 requests in flight")
	}
}

// Item-page p95 above 300 ms over the last minute holds the batch until the slow samples age out.
func TestPressureGateWaitsOnSlowItemPages(t *testing.T) {
	c := &fakeClock{t: time.Unix(1000, 0)}
	p := newTestPressure(c, nil)

	for range 30 {
		p.Record(400 * time.Millisecond)
	}

	if p.P95() <= MaxP95 {
		t.Fatalf("p95 = %v, want over %v", p.P95(), MaxP95)
	}

	start := c.t
	if err := p.Wait(t.Context()); err != nil {
		t.Fatal(err)
	}

	if waited := c.t.Sub(start); waited < Window {
		t.Errorf("waited %v, want at least the %v window", waited, Window)
	}
}

// Too few samples say nothing about the p95, and an idle site never holds a batch.
func TestPressureGateIdleAndFewSamples(t *testing.T) {
	c := &fakeClock{t: time.Unix(1000, 0)}
	p := newTestPressure(c, nil)

	p.Record(900 * time.Millisecond)

	if err := p.Wait(t.Context()); err != nil || !c.t.Equal(time.Unix(1000, 0)) {
		t.Errorf("idle gate waited: %v, %v", c.t, err)
	}
}

// The gate gives up after MaxWait so a stuck gauge cannot freeze a run, and honours cancellation.
func TestPressureGateGivesUpAndCancels(t *testing.T) {
	c := &fakeClock{t: time.Unix(1000, 0)}
	p := newTestPressure(c, nil)

	for range 20 {
		p.Begin()
	}

	if err := p.Wait(t.Context()); err != nil {
		t.Fatal(err)
	}

	if waited := c.t.Sub(time.Unix(1000, 0)); waited < MaxWait {
		t.Errorf("waited %v, want at least %v before giving up", waited, MaxWait)
	}

	p.Sleep = func(ctx context.Context, _ time.Duration) error { return ctx.Err() }

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if err := p.Wait(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

// Samples older than Window are dropped on Record, so the slice stays bounded while no run calls P95.
func TestPressureRecordTrimsOldSamples(t *testing.T) {
	c := &fakeClock{t: time.Unix(1_700_000_000, 0)}
	p := NewPressure(c.now)

	for range 100 {
		p.Record(time.Millisecond)
		c.t = c.t.Add(Window + time.Second)
	}

	if n := len(p.samples); n != 1 {
		t.Errorf("samples = %d, want 1", n)
	}
}
