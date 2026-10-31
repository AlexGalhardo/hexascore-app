//go:build integration

package moderation

import (
	"context"
	"testing"
	"time"
)

// TestMetricsQuery checks FR-04-16 on a fixture whose percentiles are computed by hand (percentile_cont
// interpolates: p95 of 1,3,5,9 h is 5 + 0.85 * 4 = 8.4 h).
func TestMetricsQuery(t *testing.T) {
	db := openDB(t)
	q := Queue{DB: db, Now: func() time.Time { return base }}

	insert(t, db, seed{ageHours: 2})
	insert(t, db, seed{ageHours: 30, status: "in_review"})
	insert(t, db, seed{ageHours: 5})

	decide := func(status string, decidedAgo, ttd float64) {
		t.Helper()
		id := insert(t, db, seed{status: status, ageHours: decidedAgo + ttd})
		if _, err := db.ExecContext(context.Background(), `UPDATE suggestions SET decided_at = $2 WHERE id = $1`,
			id, base.Add(-time.Duration(decidedAgo*float64(time.Hour)))); err != nil {
			t.Fatal(err)
		}
	}
	decide("approved", 24, 1)
	decide("applied", 48, 3)
	decide("rejected", 72, 5)
	decide("reverted", 100, 9)
	decide("rejected", 240, 20)
	decide("applied", 960, 2)  // older than 30 days
	decide("withdrawn", 24, 4) // not a moderator decision
	decide("expired", 24, 6)

	m, err := q.Metrics(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if m.Open != 3 || m.OldestOpen != 30*time.Hour {
		t.Fatalf("open %d oldest %v", m.Open, m.OldestOpen)
	}

	hours := func(h float64) time.Duration { return time.Duration(h * float64(time.Hour)) }
	near := func(got, want time.Duration) bool { d := got - want; return d > -time.Second && d < time.Second }
	for _, c := range []struct {
		w                 Window
		days              int
		decided, approved int64
		p50, p95          time.Duration
		rate              float64
	}{
		{m.Week, 7, 4, 3, hours(4), hours(8.4), 0.75},
		{m.Month, 30, 5, 3, hours(5), hours(17.8), 0.6},
	} {
		if c.w.Days != c.days || c.w.Decided != c.decided || c.w.Approved != c.approved ||
			!near(c.w.P50, c.p50) || !near(c.w.P95, c.p95) || c.w.ApprovalRate() != c.rate {
			t.Fatalf("window %d: %+v rate %v", c.days, c.w, c.w.ApprovalRate())
		}
	}
}

func TestMetricsQueryEmpty(t *testing.T) {
	m, err := Queue{DB: openDB(t), Now: func() time.Time { return base }}.Metrics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if m.Open != 0 || m.OldestOpen != 0 || m.Week.Decided != 0 || m.Week.P95 != 0 || m.Month.ApprovalRate() != 0 {
		t.Fatalf("empty metrics = %+v", m)
	}
}
