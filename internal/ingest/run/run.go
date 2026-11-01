// Package run is the ingestion run framework: the job_runs lifecycle, heartbeat, chunked continuation,
// per-target dead letters and the fail-closed kill switches (SDD 06 §6, §7.6).
package run

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"
)

// Run statuses (job_runs.status) and modes (job_runs.mode).
const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusPartial   = "partial"
	StatusFailed    = "failed"
	StatusSkipped   = "skipped"
	StatusCancelled = "cancelled"

	ModeBootstrap   = "bootstrap"
	ModeIncremental = "incremental"
	ModeFull        = "full"
	ModeFollowup    = "followup"
	ModeManual      = "manual"
	ModeBackfill    = "backfill"
	ModeRollback    = "rollback"
)

// ErrRunActive is returned by Create when job_runs_one_active already holds a queued or running run of the
// same job and source; the admin route maps it to 409 run-active.
var ErrRunActive = errors.New("ingest: a run of this job is already active")

// Job is the queue payload: parameters live in job_runs.params (SDD 06 §6).
type Job struct {
	RunID int64 `json:"run_id"`
}

// Kind names the job on the queue.
func (Job) Kind() string { return "ingest.run" }

// Dispatcher is the DispatchTx half of *queue.Queue.
type Dispatcher interface {
	DispatchTx(ctx context.Context, ex database.Executor, j queue.Job, opts ...queue.Option) error
}

// Spec describes a run to create.
type Spec struct {
	Job         string
	Source      string
	Mode        string
	Params      json.RawMessage
	Cursor      json.RawMessage
	TriggeredBy string
}

// Run is a job_runs row.
type Run struct {
	ID          int64
	Job         string
	Source      string
	Mode        string
	Status      string
	Params      json.RawMessage
	Cursor      json.RawMessage
	Attempt     int
	Seen        int
	Changed     int
	Failed      int
	Skipped     int
	RateLimited int
}

const selectRun = `SELECT id, job, coalesce(source, ''), mode, status, params, coalesce(cursor, 'null'::jsonb), attempt,
seen, changed, failed, skipped, rate_limited FROM job_runs WHERE id = $1`

func scanRun(row *sql.Row) (Run, error) {
	var r Run

	err := row.Scan(&r.ID, &r.Job, &r.Source, &r.Mode, &r.Status, &r.Params, &r.Cursor, &r.Attempt,
		&r.Seen, &r.Changed, &r.Failed, &r.Skipped, &r.RateLimited)
	if string(r.Cursor) == "null" {
		r.Cursor = nil
	}

	return r, err
}

// Load reads one run.
func Load(ctx context.Context, ex database.Executor, id int64) (Run, error) {
	r, err := scanRun(ex.QueryRowContext(ctx, selectRun, id))
	if err != nil {
		return Run{}, fmt.Errorf("ingest run %d: %w", id, err)
	}

	return r, nil
}

// Create inserts a queued run and dispatches its job in the same transaction, so a queued row always has a job.
// It returns ErrRunActive, dispatching nothing, when the unique index refuses the row.
func Create(ctx context.Context, db *database.DB, d Dispatcher, s Spec) (Run, error) {
	params := s.Params
	if len(params) == 0 {
		params = json.RawMessage(`{}`)
	}

	var cursor any
	if len(s.Cursor) > 0 {
		cursor = string(s.Cursor)
	}

	var id int64

	err := db.Tx(ctx, func(tx *database.Tx) error {
		err := tx.QueryRowContext(ctx, `INSERT INTO job_runs (job, source, mode, status, params, cursor, triggered_by)
VALUES ($1, NULLIF($2, ''), $3, 'queued', $4::jsonb, $5::jsonb, $6) ON CONFLICT DO NOTHING RETURNING id`,
			s.Job, s.Source, s.Mode, string(params), cursor, s.TriggeredBy).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRunActive
		}

		if err != nil {
			return fmt.Errorf("ingest run: insert: %w", err)
		}

		return d.DispatchTx(ctx, tx, Job{RunID: id})
	})
	if err != nil {
		return Run{}, err
	}

	return Load(ctx, db, id)
}

// Finish closes a run that is not yet over. A run that already ended keeps its status.
func Finish(ctx context.Context, ex database.Executor, id int64, status string, cause error) error {
	msg := sql.NullString{}
	if cause != nil {
		msg = sql.NullString{String: cause.Error(), Valid: true}
	}

	_, err := ex.ExecContext(ctx, `UPDATE job_runs SET status = $2, error = $3, finished_at = now()
WHERE id = $1 AND status IN ('queued', 'running')`, id, status, msg)

	return err
}

// StaleAfter is how long a running run may go without a heartbeat (SDD 06 §7.6).
const StaleAfter = "15 minutes"

// ReapStale fails the running runs whose heartbeat is older than StaleAfter and reports how many. Their cursor
// is kept, so Resume can continue them.
func ReapStale(ctx context.Context, ex database.Executor) (int, error) {
	res, err := ex.ExecContext(
		ctx,
		`UPDATE job_runs SET status = 'failed', error = 'stale heartbeat', finished_at = now()
WHERE status = 'running' AND heartbeat_at < now() - $1::interval`,
		StaleAfter,
	)
	if err != nil {
		return 0, err
	}

	n, err := res.RowsAffected()

	return int(n), err
}

// Resume queues a followup run that continues a failed run from its saved cursor.
func Resume(ctx context.Context, db *database.DB, d Dispatcher, id int64) (Run, error) {
	old, err := Load(ctx, db, id)
	if err != nil {
		return Run{}, err
	}

	if old.Status != StatusFailed {
		return Run{}, fmt.Errorf("ingest run %d is %s, only a failed run resumes", id, old.Status)
	}

	return Create(ctx, db, d, Spec{
		Job: old.Job, Source: old.Source, Mode: ModeFollowup, Params: old.Params,
		Cursor: old.Cursor, TriggeredBy: fmt.Sprintf("resume:%d", id),
	})
}
