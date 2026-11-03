package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// recentRuns is how many ingestion runs the dashboard lists.
const recentRuns = 8

// workerStale is how old the worker heartbeat may be before the dashboard calls it stale.
const workerStale = ingest.WorkerStaleAfter

// OverviewHandlers serve the React dashboard (ADR 0026): GET /admin/overview renders the Inertia page with a first
// snapshot, GET /admin/overview/data returns the same snapshot as JSON for the page's React Query refresh.
type OverviewHandlers struct {
	DB        *database.DB
	Now       func() time.Time
	Authorize func(c *web.Ctx) error
}

// OverviewCounts are the headline numbers. Items and users are planner estimates once a table is large, because an
// exact count of a million rows on every refresh would cost more than the number is worth.
type OverviewCounts struct {
	Items          int64 `json:"items"`
	Users          int64 `json:"users"`
	OpenReports    int64 `json:"openReports"`
	RunningRuns    int64 `json:"runningRuns"`
	FailedRuns24h  int64 `json:"failedRuns24h"`
	ParkedFailures int64 `json:"parkedFailures"`
}

// OverviewWorker reports the ingestion worker's heartbeat.
type OverviewWorker struct {
	// Status is "ok", "stale" or "never".
	Status     string `json:"status"`
	AgeSeconds *int64 `json:"ageSeconds"`
}

// OverviewRun is one row of the recent runs table.
type OverviewRun struct {
	ID          int64   `json:"id"`
	Job         string  `json:"job"`
	Source      string  `json:"source"`
	Mode        string  `json:"mode"`
	Status      string  `json:"status"`
	Seen        int     `json:"seen"`
	Changed     int     `json:"changed"`
	Failed      int     `json:"failed"`
	StartedAt   *string `json:"startedAt"`
	DurationSec *int64  `json:"durationSeconds"`
}

// OverviewData is the whole snapshot.
type OverviewData struct {
	GeneratedAt string         `json:"generatedAt"`
	Counts      OverviewCounts `json:"counts"`
	Worker      OverviewWorker `json:"worker"`
	Runs        []OverviewRun  `json:"runs"`
}

// Index renders the dashboard page with the first snapshot.
func (h OverviewHandlers) Index(c *web.Ctx) error {
	if err := h.Authorize(c); err != nil {
		return err
	}

	data, err := h.snapshot(c.Context())
	if err != nil {
		return err
	}

	return inertia.Render(c, "Admin/Dashboard", inertia.Props{"overview": data})
}

// Data returns the snapshot as JSON; the page refreshes it with React Query.
func (h OverviewHandlers) Data(c *web.Ctx) error {
	if err := h.Authorize(c); err != nil {
		return err
	}

	data, err := h.snapshot(c.Context())
	if err != nil {
		return err
	}

	c.SetHeader("Cache-Control", "no-store")

	return c.JSON(http.StatusOK, data)
}

func (h OverviewHandlers) snapshot(ctx context.Context) (OverviewData, error) {
	now := h.Now().UTC()
	out := OverviewData{GeneratedAt: now.Format(time.RFC3339), Runs: []OverviewRun{}}

	var err error

	if out.Counts.Items, err = approxCount(ctx, h.DB, "items"); err != nil {
		return out, err
	}

	if out.Counts.Users, err = approxCount(ctx, h.DB, "users"); err != nil {
		return out, err
	}

	queries := []struct {
		dst *int64
		sql string
		arg []any
	}{
		{&out.Counts.OpenReports, `SELECT count(*) FROM reports WHERE status = 'open'`, nil},
		{&out.Counts.RunningRuns, `SELECT count(*) FROM job_runs WHERE status = $1`, []any{run.StatusRunning}},
		{&out.Counts.FailedRuns24h,
			`SELECT count(*) FROM job_runs WHERE status = $1 AND created_at > $2`,
			[]any{run.StatusFailed, now.Add(-24 * time.Hour)}},
		{&out.Counts.ParkedFailures, `SELECT count(*) FROM ingest_failures WHERE parked`, nil},
	}
	for _, q := range queries {
		if err := h.DB.QueryRowContext(ctx, q.sql, q.arg...).Scan(q.dst); err != nil {
			return out, fmt.Errorf("admin overview: %w", err)
		}
	}

	out.Worker, err = h.worker(ctx, now)
	if err != nil {
		return out, err
	}

	out.Runs, err = h.runs(ctx)

	return out, err
}

func (h OverviewHandlers) worker(ctx context.Context, now time.Time) (OverviewWorker, error) {
	at, ok, err := ingest.ReadWorkerHeartbeat(ctx, h.DB)
	if err != nil {
		return OverviewWorker{}, err
	}

	if !ok {
		return OverviewWorker{Status: "never"}, nil
	}

	age := int64(now.Sub(at).Seconds())
	status := "ok"

	if now.Sub(at) >= workerStale {
		status = "stale"
	}

	return OverviewWorker{Status: status, AgeSeconds: &age}, nil
}

func (h OverviewHandlers) runs(ctx context.Context) ([]OverviewRun, error) {
	rows, err := h.DB.QueryContext(ctx, `SELECT id, job, coalesce(source, ''), mode, status, seen, changed, failed,
  started_at, finished_at FROM job_runs ORDER BY id DESC LIMIT $1`, recentRuns)
	if err != nil {
		return nil, fmt.Errorf("admin overview runs: %w", err)
	}
	defer rows.Close()

	out := []OverviewRun{}

	for rows.Next() {
		var (
			r                 OverviewRun
			started, finished sql.NullTime
		)
		if err := rows.Scan(&r.ID, &r.Job, &r.Source, &r.Mode, &r.Status, &r.Seen, &r.Changed, &r.Failed,
			&started, &finished); err != nil {
			return nil, fmt.Errorf("admin overview runs: %w", err)
		}

		if started.Valid {
			s := started.Time.UTC().Format(time.RFC3339)
			r.StartedAt = &s

			if finished.Valid {
				d := int64(finished.Time.Sub(started.Time).Seconds())
				r.DurationSec = &d
			}
		}

		out = append(out, r)
	}

	return out, rows.Err()
}

// approxCount returns the planner's row estimate for a table, or an exact count while the estimate is unknown (a
// table never analysed reports 0 or -1). table is a constant of this package, never request input.
func approxCount(ctx context.Context, db *database.DB, table string) (int64, error) {
	var n int64

	err := db.QueryRowContext(ctx, `SELECT greatest(reltuples, 0)::bigint FROM pg_class WHERE oid = $1::regclass`, table).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("admin overview estimate %s: %w", table, err)
	}

	if n > 0 {
		return n, nil
	}

	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM `+table).Scan(&n); err != nil {
		return 0, fmt.Errorf("admin overview count %s: %w", table, err)
	}

	return n, nil
}
