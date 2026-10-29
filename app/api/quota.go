package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

const (
	// FlushInterval is how often the meter writes its deltas (SDD 05 §7.2).
	FlushInterval = time.Minute
	// MaxUnflushed is how long deltas may stay unwritten before the meter fails closed.
	MaxUnflushed = 10 * time.Minute
	// flushBatch is the number of users written per transaction.
	flushBatch = 500
)

// ErrMeterStale is returned by Admit when deltas have not been flushed for MaxUnflushed: the API answers 503
// instead of serving requests it may never count.
var ErrMeterStale = errors.New("api: quota meter not flushed")

// QuotaStore loads and saves the meter's counters (PGQuotaStore in production).
type QuotaStore interface {
	// Load returns the used count of userID in period and the user's remaining legacy credits.
	Load(ctx context.Context, userID int64, period string) (used, credits int, err error)
	// Save adds every delta of one user batch in a single transaction.
	Save(ctx context.Context, batch []UserDelta) error
}

// UserDelta is what one user consumed since the last flush.
type UserDelta struct {
	UserID int64
	// Used adds to api_quota_counters per period; Credits is subtracted from api_credit_balances.
	Used    map[string]int
	Credits int
	// Count and Refunded add to api_usage_daily per UTC day ("2006-01-02").
	Count    map[string]int
	Refunded map[string]int
	// Keys maps key ids to their last use, for api_keys.last_used_at.
	Keys map[string]time.Time
}

func (d *UserDelta) empty() bool {
	for _, m := range []map[string]int{d.Used, d.Count, d.Refunded} {
		for _, n := range m {
			if n != 0 {
				return false
			}
		}
	}

	return d.Credits == 0 && len(d.Keys) == 0
}

// userState is one user's counters: base is the stored value when loaded or last flushed, pending the delta.
type userState struct {
	used        map[string]int // period -> stored used
	credits     int            // stored remaining credits
	pending     UserDelta
	creditsSeen bool
}

// Usage is the admission result that feeds the X-Quota-* headers.
type Usage struct {
	Plan      Plan
	Remaining int
	ResetIn   int
	Exhausted bool
	ticket    ticket
}

type ticket struct {
	userID int64
	period string
	day    string
	credit bool
}

// Meter meters counted requests in memory and flushes deltas every FlushInterval (ADR 0003: counters in
// Postgres, no Redis). A crash loses at most one interval of deltas, so it undercounts and never overcharges.
// shortcut: single instance (D-08) and idle users are never evicted from memory; shard or evict when the app
// scales out or memory shows it.
type Meter struct {
	Store QuotaStore
	Now   func() time.Time
	// Log receives flush failures (SDD 12 alerts on repeated ones); nil discards them.
	Log *slog.Logger

	mu        sync.Mutex
	users     map[int64]*userState
	lastFlush time.Time
}

func newDelta(userID int64) UserDelta {
	return UserDelta{
		UserID: userID, Used: map[string]int{}, Count: map[string]int{}, Refunded: map[string]int{},
		Keys: map[string]time.Time{},
	}
}

// state returns the user's state with the period loaded, reading the store outside the lock.
func (m *Meter) state(ctx context.Context, userID int64, period string) (*userState, error) {
	m.mu.Lock()
	if m.users == nil {
		m.users = map[int64]*userState{}
		m.lastFlush = m.Now()
	}

	s := m.users[userID]
	if s != nil {
		if _, ok := s.used[period]; ok {
			return s, nil // still locked
		}
	}
	m.mu.Unlock()

	used, credits, err := m.Store.Load(ctx, userID, period)
	if err != nil {
		return nil, fmt.Errorf("load quota of user %d: %w", userID, err)
	}

	m.mu.Lock()

	s = m.users[userID]
	if s == nil {
		s = &userState{used: map[string]int{}, pending: newDelta(userID)}
		m.users[userID] = s
	}

	if _, ok := s.used[period]; !ok {
		s.used[period] = used
	}

	if !s.creditsSeen {
		s.credits, s.creditsSeen = credits, true
	}

	return s, nil
}

// Admit counts one request of userID with keyID under plan, or reports the quota exhausted. Legacy credits are
// consumed before the plan quota (D-23). The returned usage must go to Refund when the response is not counted.
func (m *Meter) Admit(ctx context.Context, userID int64, keyID string, plan Plan) (Usage, error) {
	now := m.Now()
	period := plan.Period(now)

	s, err := m.state(ctx, userID, period)
	if err != nil {
		return Usage{}, err
	}
	defer m.mu.Unlock()

	if now.Sub(m.lastFlush) > MaxUnflushed && m.pendingLocked() {
		return Usage{}, ErrMeterStale
	}

	u := Usage{Plan: plan, ResetIn: plan.ResetIn(now)}
	credits := s.credits - s.pending.Credits
	used := s.used[period] + s.pending.Used[period]

	switch {
	case credits > 0:
		s.pending.Credits++
		credits--
		u.ticket.credit = true
	case used < plan.Quota:
		s.pending.Used[period]++
		used++
	default:
		u.Exhausted = true
		return u, nil
	}

	day := now.UTC().Format(time.DateOnly)
	s.pending.Count[day]++

	if keyID != "" {
		s.pending.Keys[keyID] = now
	}

	u.Remaining = max(0, plan.Quota-used) + credits
	u.ticket = ticket{userID: userID, period: period, day: day, credit: u.ticket.credit}

	return u, nil
}

// Used is the plan quota userID consumed in the current period: the stored counter plus the unflushed delta
// (credits excluded), so it agrees with X-Quota-Remaining of the same request.
func (m *Meter) Used(ctx context.Context, userID int64, plan Plan) (int, error) {
	period := plan.Period(m.Now())

	s, err := m.state(ctx, userID, period)
	if err != nil {
		return 0, err
	}
	defer m.mu.Unlock()

	return s.used[period] + s.pending.Used[period], nil
}

// Refund uncounts an admitted request whose response is not counted (FR-05-4): a query 400, or a 5xx, which is
// also recorded as refunded.
func (m *Meter) Refund(u Usage, serverError bool) {
	t := u.ticket
	if t.userID == 0 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	s := m.users[t.userID]
	if s == nil {
		return
	}

	// A flush may have written the admission already: the delta goes negative and the next flush subtracts it.
	if t.credit {
		s.pending.Credits--
	} else {
		s.pending.Used[t.period]--
	}

	s.pending.Count[t.day]--

	if serverError {
		s.pending.Refunded[t.day]++
	}
}

func (m *Meter) pendingLocked() bool {
	for _, s := range m.users {
		if !s.pending.empty() {
			return true
		}
	}

	return false
}

// Flush writes every pending delta, flushBatch users per transaction. A failed batch keeps its deltas for the
// next flush; batches before it stay written. Snapshotted deltas move into the base under the lock before the
// write, so requests admitted while it runs still see them (no over-admission during a flush); restore takes
// them back out on failure.
func (m *Meter) Flush(ctx context.Context) error {
	m.mu.Lock()

	var batch []UserDelta

	for id, s := range m.users {
		if s.pending.empty() {
			continue
		}

		for p, n := range s.pending.Used {
			s.used[p] += n
		}

		s.credits -= s.pending.Credits
		batch = append(batch, s.pending)
		s.pending = newDelta(id)
	}

	if len(batch) == 0 {
		m.lastFlush = m.Now()
	}
	m.mu.Unlock()

	for start := 0; start < len(batch); start += flushBatch {
		chunk := batch[start:min(start+flushBatch, len(batch))]

		if err := m.Store.Save(ctx, chunk); err != nil {
			m.restore(batch[start:])
			return fmt.Errorf("flush api usage: %w", err)
		}

		m.applied(chunk)
	}

	if len(batch) > 0 {
		m.mu.Lock()
		m.lastFlush = m.Now()
		m.mu.Unlock()
	}

	return nil
}

// applied drops periods that are no longer current once their deltas are written.
func (m *Meter) applied(chunk []UserDelta) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.Now().UTC().Format("200601")

	for _, d := range chunk {
		s := m.users[d.UserID]
		for p := range s.used {
			if p != "life" && p != now && s.pending.Used[p] == 0 {
				delete(s.used, p)
			}
		}
	}
}

// restore takes unwritten deltas out of the base and puts them back in front of whatever arrived during the flush.
func (m *Meter) restore(unwritten []UserDelta) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, d := range unwritten {
		s := m.users[d.UserID]
		p := &s.pending
		p.Credits += d.Credits
		s.credits += d.Credits

		for k, n := range d.Used {
			p.Used[k] += n
			s.used[k] -= n
		}

		for k, n := range d.Count {
			p.Count[k] += n
		}

		for k, n := range d.Refunded {
			p.Refunded[k] += n
		}

		for k, at := range d.Keys {
			if at.After(p.Keys[k]) {
				p.Keys[k] = at
			}
		}
	}
}

// Run flushes every FlushInterval until ctx ends, then once more with a fresh context so shutdown keeps the last
// interval (vukem.Runner).
func (m *Meter) Run(ctx context.Context) error {
	tick := time.NewTicker(FlushInterval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// Retried next tick; Admit fails closed after MaxUnflushed.
			if err := m.Flush(ctx); err != nil && m.Log != nil {
				m.Log.ErrorContext(ctx, "api usage flush failed", "error", err)
			}
		case <-ctx.Done():
			final, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			defer cancel()

			return m.Flush(final)
		}
	}
}
