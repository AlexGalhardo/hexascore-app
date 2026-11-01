package ingest

import (
	"context"
	"slices"
	"sync"
	"time"
)

// The pressure gate thresholds (SDD 06 §7.5, NFR-06-1): ingestion batches wait while the app serves more than
// MaxInFlight requests or the last minute's item-page p95 is over MaxP95.
const (
	MaxInFlight = 16
	MaxP95      = 300 * time.Millisecond
	Window      = time.Minute
	// MinSamples is how many item-page samples a p95 needs; an idle site says nothing about latency.
	MinSamples = 20
	// MaxWait is the longest one Wait holds a batch. shortcut: the gate fails open after it, so a stuck gauge
	// cannot freeze a run past its heartbeat; revisit if sustained load should pause runs for longer.
	MaxWait = 2 * time.Minute
	// PollEvery is how often Wait re-reads the gauges.
	PollEvery = 500 * time.Millisecond
)

type sample struct {
	at time.Time
	d  time.Duration
}

// Pressure counts in-flight requests and item-page latency, and holds ingestion batches while the site is busy.
// The HTTP middleware feeds it (Begin); the run framework asks it (Wait).
type Pressure struct {
	// Sleep replaces the real wait; tests only.
	Sleep func(ctx context.Context, d time.Duration) error

	mu       sync.Mutex
	now      func() time.Time
	inflight int
	samples  []sample
}

// NewPressure returns a gate reading time from now.
func NewPressure(now func() time.Time) *Pressure {
	return &Pressure{now: now, Sleep: sleep}
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// Begin counts a request and returns the function that ends it; finish(true) also records its duration as an
// item-page sample.
func (p *Pressure) Begin() (finish func(itemPage bool)) {
	p.mu.Lock()
	p.inflight++
	start := p.now()
	p.mu.Unlock()

	return func(itemPage bool) {
		p.mu.Lock()
		p.inflight--
		p.mu.Unlock()

		if itemPage {
			p.Record(p.now().Sub(start))
		}
	}
}

// Record adds one item-page latency sample.
func (p *Pressure) Record(d time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Trim here too: P95 runs only while an ingestion run waits, so between runs the slice would grow per request.
	now := p.now()
	cut := now.Add(-Window)
	p.samples = slices.DeleteFunc(p.samples, func(s sample) bool { return s.at.Before(cut) })
	p.samples = append(p.samples, sample{at: now, d: d})
}

// InFlight is the number of requests being served.
func (p *Pressure) InFlight() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.inflight
}

// P95 is the item-page 95th percentile of the last Window, or 0 with fewer than MinSamples samples.
func (p *Pressure) P95() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()

	cut := p.now().Add(-Window)
	p.samples = slices.DeleteFunc(p.samples, func(s sample) bool { return s.at.Before(cut) })

	if len(p.samples) < MinSamples {
		return 0
	}

	ds := make([]time.Duration, len(p.samples))
	for i, s := range p.samples {
		ds[i] = s.d
	}

	slices.Sort(ds)

	return ds[(len(ds)*95+99)/100-1]
}

func (p *Pressure) busy() bool { return p.InFlight() > MaxInFlight || p.P95() > MaxP95 }

// Wait returns when the site is quiet, after MaxWait, or when ctx ends (then with its error).
func (p *Pressure) Wait(ctx context.Context) error {
	deadline := p.now().Add(MaxWait)

	for p.busy() && p.now().Before(deadline) {
		if err := p.Sleep(ctx, PollEvery); err != nil {
			return err
		}
	}

	return nil
}
