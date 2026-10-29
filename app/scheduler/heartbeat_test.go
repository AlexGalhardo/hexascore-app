package scheduler

import (
	"testing"
	"time"
)

func TestHeartbeat_AgeLimit(t *testing.T) {
	t.Parallel()

	now := time.Unix(1_000, 0)
	h := NewHeartbeat(func() time.Time { return now })

	if err := h.Check(t.Context()); err != nil {
		t.Fatalf("fresh heartbeat: %v", err)
	}

	now = now.Add(MaxAge - time.Second)
	if err := h.Check(t.Context()); err != nil {
		t.Fatalf("heartbeat %s old: %v", MaxAge-time.Second, err)
	}

	now = now.Add(time.Second)
	if err := h.Check(t.Context()); err == nil {
		t.Fatal("heartbeat MaxAge old: want error")
	}

	if err := h.Beat(t.Context()); err != nil {
		t.Fatal(err)
	}

	if err := h.Check(t.Context()); err != nil {
		t.Fatalf("after beat: %v", err)
	}
}
