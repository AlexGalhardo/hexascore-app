package httpx

import (
	"context"
	"time"
)

// HostState is the part of a host's politeness state that must survive a restart (RFC-001 section 3, decision 4):
// without it a redeploy would forget an open breaker or a halved rate and hit the host again at full speed.
type HostState struct {
	Rate      float64       `json:"rate"`
	OpenFor   time.Duration `json:"open_for_ns"`
	OpenUntil time.Time     `json:"open_until"`
	Last429At time.Time     `json:"last_429_at"`
	// CallsToday counts requests sent on Day (UTC date, YYYY-MM-DD); a new day starts from zero.
	CallsToday int    `json:"calls_today"`
	Day        string `json:"day"`
}

// StateStore persists HostState by host name. Errors are logged by the caller's choice; the client never fails a
// request because the store did.
type StateStore interface {
	Load(ctx context.Context) (map[string]HostState, error)
	Save(ctx context.Context, host string, s HostState) error
}

const saveTimeout = 5 * time.Second

func day(t time.Time) string { return t.UTC().Format(time.DateOnly) }

// state snapshots the persisted fields. Callers hold h.mu.
func (h *host) state(now time.Time) HostState {
	return HostState{
		Rate:       h.rate,
		OpenFor:    h.br.openFor,
		OpenUntil:  h.br.openUntil,
		Last429At:  h.last429,
		CallsToday: h.callsOn(day(now)),
		Day:        day(now),
	}
}

func (h *host) callsOn(d string) int {
	if h.callsDay != d {
		return 0
	}

	return h.calls
}

// count registers one request sent at now. Callers hold h.mu.
func (h *host) count(now time.Time) {
	if d := day(now); h.callsDay != d {
		h.callsDay, h.calls = d, 0
	}

	h.calls++
}

// restore applies a stored state. The lower of the stored and configured rate wins, so a halving survives while
// a host that was never throttled keeps its configured rate. Callers hold h.mu.
func (h *host) restore(s HostState, now time.Time) {
	if s.Rate > 0 && (h.rate == 0 || s.Rate < h.rate) {
		h.rate = s.Rate
	}

	h.br.openFor = s.OpenFor
	h.br.openUntil = s.OpenUntil
	h.last429 = s.Last429At

	if s.Day == day(now) {
		h.callsDay, h.calls = s.Day, s.CallsToday
	}
}

// Restore loads the stored state of every configured host; call it once before the first request.
func (c *Client) Restore(ctx context.Context) error {
	if c.p.store == nil {
		return nil
	}

	states, err := c.p.store.Load(ctx)
	if err != nil {
		return err
	}

	now := c.p.clock.Now()

	for name, s := range states {
		if h, ok := c.p.hosts[name]; ok {
			h.mu.Lock()
			h.restore(s, now)
			h.mu.Unlock()
		}
	}

	return nil
}

// Flush stores the state of every host, calls_today included (it changes on every request, so transitions alone
// do not carry it). Returns the first error and still tries the other hosts.
func (c *Client) Flush(ctx context.Context) error {
	var first error

	for name, h := range c.p.hosts {
		if err := c.p.save(ctx, name, h); err != nil && first == nil {
			first = err
		}
	}

	return first
}

func (p *policy) save(ctx context.Context, name string, h *host) error {
	if p.store == nil {
		return nil
	}

	h.mu.Lock()
	s := h.state(p.clock.Now())
	h.mu.Unlock()

	return p.store.Save(ctx, name, s)
}

// saveSoon stores one host after a transition (breaker opened or closed, rate halved, first 429 of a while). It is
// best effort: a failed save only means a restart forgets this transition.
func (p *policy) saveSoon(name string, h *host) {
	if p.store == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), saveTimeout)
	defer cancel()

	_ = p.save(ctx, name, h)
}
