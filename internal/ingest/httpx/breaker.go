package httpx

import "time"

const (
	breakerConsecutive = 10
	breakerWindow      = 50
	breakerBaseOpen    = 5 * time.Minute
	breakerMaxOpen     = time.Hour
)

// breakerState opens after 10 consecutive failures or >= 50% of the last
// 50 outcomes, stays open 5 min doubling to 1 h, then lets one probe
// through (SDD 06 §7.2). Guarded by host.mu.
type breakerState struct {
	consecutive int
	window      [breakerWindow]bool // ring of failure flags
	filled      int
	pos         int
	openFor     time.Duration
	openUntil   time.Time
	probing     bool
}

// allow reports whether a request may go; it marks the caller as the probe
// when the open period has elapsed.
func (b *breakerState) allow(now time.Time) bool {
	if b.openFor == 0 {
		return true
	}

	if now.Before(b.openUntil) || b.probing {
		return false
	}

	b.probing = true

	return true
}

// record registers an outcome; cancelled requests must not be recorded.
func (b *breakerState) record(now time.Time, failed bool) {
	wasProbe := b.probing
	b.probing = false

	if !failed {
		b.consecutive = 0
		b.openFor = 0
		b.push(false)

		return
	}

	b.consecutive++
	b.push(true)

	if wasProbe {
		b.open(now)

		return
	}

	if b.consecutive >= breakerConsecutive || b.failureRatio() {
		b.open(now)
	}
}

func (b *breakerState) open(now time.Time) {
	if b.openFor == 0 {
		b.openFor = breakerBaseOpen
	} else {
		b.openFor = min(b.openFor*2, breakerMaxOpen)
	}

	b.openUntil = now.Add(b.openFor)
	b.consecutive = 0
	b.filled = 0
	b.pos = 0
}

func (b *breakerState) push(failed bool) {
	b.window[b.pos] = failed
	b.pos = (b.pos + 1) % breakerWindow
	b.filled = min(b.filled+1, breakerWindow)
}

func (b *breakerState) failureRatio() bool {
	if b.filled < breakerWindow {
		return false
	}

	n := 0
	for _, f := range b.window {
		if f {
			n++
		}
	}

	return n*2 >= breakerWindow
}
