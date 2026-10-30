// Package scheduler holds the scheduler's own liveness signal.
package scheduler

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// MaxAge is the oldest heartbeat the readiness check accepts: three missed
// minutes of an EveryMinute task (SDD 12 §7.1).
const MaxAge = 180 * time.Second

// Heartbeat is beaten by the "heartbeat" task and read by the scheduler
// readiness check. It lives in memory, so the check is only meaningful in the
// process that runs the schedule (SCHEDULE_IN_SERVE=true).
type Heartbeat struct {
	now  func() time.Time
	last atomic.Int64
}

// NewHeartbeat starts the clock at boot, so a fresh process is ready before
// the first minute tick instead of failing its deploy healthcheck.
func NewHeartbeat(now func() time.Time) *Heartbeat {
	h := &Heartbeat{now: now}
	h.last.Store(now().UnixNano())

	return h
}

// Beat records a scheduler tick.
func (h *Heartbeat) Beat(context.Context) error {
	h.last.Store(h.now().UnixNano())

	return nil
}

// Age is how long ago Beat last ran.
func (h *Heartbeat) Age() time.Duration {
	return h.now().Sub(time.Unix(0, h.last.Load()))
}

// Check fails when the last beat is MaxAge old or older.
func (h *Heartbeat) Check(context.Context) error {
	if age := h.now().Sub(time.Unix(0, h.last.Load())); age >= MaxAge {
		return fmt.Errorf("scheduler: last heartbeat %s ago", age.Round(time.Second))
	}

	return nil
}
