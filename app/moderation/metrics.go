package moderation

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Window holds the decisions of the last Days days: moderator outcomes only (approved, applied, reverted count as
// approvals; rejected as a rejection); withdrawn and expired suggestions are system transitions.
type Window struct {
	Days     int
	Decided  int64
	Approved int64
	P50, P95 time.Duration
}

// ApprovalRate is Approved over Decided, 0 without decisions.
func (w Window) ApprovalRate() float64 {
	if w.Decided == 0 {
		return 0
	}

	return float64(w.Approved) / float64(w.Decided)
}

// Metrics is the FR-04-16 dashboard: queue size, oldest open age and the 7 and 30 day decision windows.
type Metrics struct {
	Open       int64
	OldestOpen time.Duration
	Week       Window
	Month      Window
}

// Metrics computes the dashboard in one query (SDD 04 FR-04-16, NFR-04-2); time-to-decision is decided_at minus
// submitted_at, by decision date.
func (q Queue) Metrics(ctx context.Context) (Metrics, error) {
	m := Metrics{Week: Window{Days: 7}, Month: Window{Days: 30}}
	var oldest, wp50, wp95, mp50, mp95 sql.NullFloat64

	err := q.DB.QueryRowContext(ctx, `WITH s AS (
  SELECT status IN ('submitted','in_review') AS open, submitted_at,
    status IN ('approved','applied','reverted','rejected') AND decided_at >= $1::timestamptz - interval '30 days'
      AS decided,
    decided_at >= $1::timestamptz - interval '7 days' AS week,
    status <> 'rejected' AS approved,
    EXTRACT(EPOCH FROM decided_at - submitted_at) AS ttd
  FROM suggestions
  WHERE status IN ('submitted','in_review')
     OR (decided_at >= $1::timestamptz - interval '30 days' AND submitted_at IS NOT NULL))
SELECT count(*) FILTER (WHERE open),
  EXTRACT(EPOCH FROM $1::timestamptz - min(submitted_at) FILTER (WHERE open)),
  count(*) FILTER (WHERE decided AND week), count(*) FILTER (WHERE decided AND week AND approved),
  percentile_cont(0.5) WITHIN GROUP (ORDER BY ttd) FILTER (WHERE decided AND week),
  percentile_cont(0.95) WITHIN GROUP (ORDER BY ttd) FILTER (WHERE decided AND week),
  count(*) FILTER (WHERE decided), count(*) FILTER (WHERE decided AND approved),
  percentile_cont(0.5) WITHIN GROUP (ORDER BY ttd) FILTER (WHERE decided),
  percentile_cont(0.95) WITHIN GROUP (ORDER BY ttd) FILTER (WHERE decided)
FROM s`, q.now()).Scan(&m.Open, &oldest, &m.Week.Decided, &m.Week.Approved, &wp50, &wp95,
		&m.Month.Decided, &m.Month.Approved, &mp50, &mp95)
	if err != nil {
		return m, fmt.Errorf("moderation: metrics: %w", err)
	}

	m.OldestOpen, m.Week.P50, m.Week.P95 = seconds(oldest), seconds(wp50), seconds(wp95)
	m.Month.P50, m.Month.P95 = seconds(mp50), seconds(mp95)

	return m, nil
}

func seconds(v sql.NullFloat64) time.Duration {
	return time.Duration(v.Float64 * float64(time.Second)).Round(time.Second)
}
