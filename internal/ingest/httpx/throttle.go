package httpx

import (
	"context"
	"io"
	"sync"
	"time"
)

// Clock is the time source of the client; tests drive it by hand.
type Clock interface {
	Now() time.Time
	Sleep(ctx context.Context, d time.Duration) error
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

func (realClock) Sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}

	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// HostLimit is the politeness budget of one host (SDD 06 §7.2).
type HostLimit struct {
	// PerSecond is the sustained rate; 0 means unthrottled.
	PerSecond float64
	// InFlight bounds requests whose body is still open; 0 means unbounded.
	InFlight int
}

// minRate stops repeated halving from stalling a host forever.
const minRate = 0.05

// halveAfter is the number of rate-limit hits that halves a host's rate.
const halveAfter = 20

type host struct {
	mu          sync.Mutex
	rate        float64
	next        time.Time // earliest start of the next reserved slot
	pausedUntil time.Time
	hits        int
	sem         chan struct{}
	br          breakerState
}

func newHost(l HostLimit) *host {
	h := &host{rate: l.PerSecond}
	if l.InFlight > 0 {
		h.sem = make(chan struct{}, l.InFlight)
	}

	return h
}

// wait reserves a slot (slot reservation: concurrent callers get distinct,
// spaced start times) and sleeps until it, honouring a host-wide pause that
// may have been set while sleeping.
func (h *host) wait(ctx context.Context, c Clock) error {
	for {
		h.mu.Lock()

		now := c.Now()
		at := now

		if h.next.After(at) {
			at = h.next
		}

		if h.pausedUntil.After(at) {
			at = h.pausedUntil
		}

		if h.rate > 0 {
			h.next = at.Add(time.Duration(float64(time.Second) / h.rate))
		}

		h.mu.Unlock()

		if err := c.Sleep(ctx, at.Sub(now)); err != nil {
			return err
		}

		h.mu.Lock()
		paused := h.pausedUntil.After(c.Now())
		h.mu.Unlock()

		if !paused {
			return nil
		}
	}
}

// pause stops every request to the host until now+d (Retry-After).
func (h *host) pause(c Clock, d time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if until := c.Now().Add(d); until.After(h.pausedUntil) {
		h.pausedUntil = until
	}
}

// rateLimited counts a 429 and halves the rate on every 20th.
func (h *host) rateLimited() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.hits++
	if h.hits%halveAfter == 0 && h.rate > 0 {
		h.rate = max(h.rate/2, minRate)
	}
}

func (h *host) acquire(ctx context.Context) (release func(), err error) {
	if h.sem == nil {
		return func() {}, nil
	}

	select {
	case h.sem <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	return sync.OnceFunc(func() { <-h.sem }), nil
}

// releaseBody frees the in-flight slot when the body is closed.
type releaseBody struct {
	io.ReadCloser
	release func()
}

func (b *releaseBody) Close() error {
	err := b.ReadCloser.Close()
	b.release()

	return err
}
