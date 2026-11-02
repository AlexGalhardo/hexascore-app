package run

import (
	"testing"
	"time"
)

// seeded is a tiny deterministic generator (xorshift) so the test needs no math/rand.
func seeded(seed uint64) func() float64 {
	x := seed

	return func() float64 {
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17

		return float64(x>>11) / (1 << 53)
	}
}

func TestSecureFloatInRange(t *testing.T) {
	for range 100 {
		if f := secureFloat(); f < 0 || f >= 1 {
			t.Fatalf("secureFloat = %v", f)
		}
	}
}

func TestJitterBackoffDiffersAndStaysUnderTheCeiling(t *testing.T) {
	b := JitterBackoff(RetryBase, RetryCap, seeded(7))

	first, second := b(3), b(3)
	if first == second {
		t.Errorf("two attempts computed the same backoff %v", first)
	}

	for attempt := 1; attempt <= 40; attempt++ {
		ceil := RetryCap
		if attempt <= 6 {
			ceil = min(RetryCap, RetryBase<<(attempt-1))
		}

		if d := b(attempt); d < 0 || d > ceil {
			t.Errorf("attempt %d: %v outside [0, %v]", attempt, d, ceil)
		}
	}

	if got := JitterBackoff(RetryBase, RetryCap, seeded(7))(3); got != first {
		t.Errorf("same seed gave %v and %v", first, got)
	}

	if d := JitterBackoff(time.Second, time.Minute, func() float64 { return 0.5 })(1); d != 500*time.Millisecond {
		t.Errorf("attempt 1 with rnd 0.5 = %v, want 500ms", d)
	}
}

func TestKeyedTriggers(t *testing.T) {
	for by, want := range map[string]bool{
		"schedule": true, RetryTrigger: true, "chain:12": true, "admin:1": false, "cli": false, "resume:3": false,
	} {
		if keyed(by) != want {
			t.Errorf("keyed(%q) = %v, want %v", by, !want, want)
		}
	}
}
