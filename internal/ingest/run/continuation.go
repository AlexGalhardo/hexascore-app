package run

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/flags"
	"github.com/AlexGalhardo/vukem/queue"
)

// Outcome is how one Execute call ended.
type Outcome string

const (
	// Finished means the run reached a final status.
	Finished Outcome = "finished"
	// Continued means the budget ran out and the successor job was dispatched with the saved cursor.
	Continued Outcome = "continued"
	// Skipped means a kill switch was off; the source was never touched.
	Skipped Outcome = "skipped"
	// Cancelled means an operator cancelled the run.
	Cancelled Outcome = "cancelled"
)

// Limits bound one job execution (SDD 06 §6: at most 500 targets or 5 minutes).
type Limits struct {
	Targets  int
	Duration time.Duration
}

// Result is what one Step did. Cursor is saved in the same transaction as the Step's writes.
type Result struct {
	Cursor      json.RawMessage
	Done        bool
	Targets     int
	Seen        int
	Changed     int
	Failed      int
	Skipped     int
	RateLimited int
}

// Step processes one chunk of targets inside tx, starting at r.Cursor. Everything it writes commits together
// with the new cursor, so an interrupted run replays at most one chunk.
type Step func(ctx context.Context, tx *database.Tx, r Run) (Result, error)

// Runner executes runs by job name.
type Runner struct {
	DB     *database.DB
	Queue  Dispatcher
	Flags  *flags.Flags
	Limits Limits
	Steps  map[string]Step
	Now    func() time.Time
}

func (r *Runner) limits() Limits {
	l := r.Limits
	if l.Targets <= 0 {
		l.Targets = 500
	}

	if l.Duration <= 0 {
		l.Duration = 5 * time.Minute
	}

	return l
}

// Execute runs chunks of run id until it is done, cancelled, or the chunk budget is spent, in which case the
// successor job is dispatched in the transaction that saved the cursor. It is safe to call again for the same
// id (a redelivered or retried job resumes from the cursor).
func (r *Runner) Execute(ctx context.Context, id int64) (Outcome, error) {
	run, err := Load(ctx, r.DB, id)
	if err != nil {
		return "", fmt.Errorf("%w: %w", err, queue.ErrPermanent)
	}

	switch run.Status {
	case StatusQueued, StatusRunning:
	case StatusCancelled:
		return Cancelled, nil
	default:
		return Finished, nil
	}

	step, ok := r.Steps[run.Job]
	if !ok {
		return "", fmt.Errorf("ingest run %d: unknown job %q: %w", id, run.Job, queue.ErrPermanent)
	}

	if !Allowed(ctx, r.Flags, run.Source) {
		return Skipped, Finish(ctx, r.DB, id, StatusSkipped, nil)
	}

	if _, err := r.DB.ExecContext(ctx, `UPDATE job_runs SET status = 'running', attempt = attempt + 1,
started_at = coalesce(started_at, now()), heartbeat_at = now() WHERE id = $1`, id); err != nil {
		return "", err
	}

	lim := r.limits()
	deadline := r.Now().Add(lim.Duration)
	processed := 0

	for {
		var out Outcome

		err := r.DB.Tx(ctx, func(tx *database.Tx) error {
			var status string
			if err := tx.QueryRowContext(ctx, `SELECT status FROM job_runs WHERE id = $1 FOR UPDATE`, id).
				Scan(&status); err != nil {
				return err
			}

			if status == StatusCancelled {
				out = Cancelled

				return nil
			}

			res, err := step(ctx, tx, run)
			if err != nil {
				return err
			}

			processed += res.Targets
			run.Cursor = res.Cursor
			run.Failed += res.Failed

			if _, err := tx.ExecContext(ctx, `UPDATE job_runs SET cursor = $2::jsonb, heartbeat_at = now(),
seen = seen + $3, changed = changed + $4, failed = failed + $5, skipped = skipped + $6, rate_limited = rate_limited + $7
WHERE id = $1`, id, cursorArg(res.Cursor), res.Seen, res.Changed, res.Failed, res.Skipped, res.RateLimited); err != nil {
				return err
			}

			switch {
			case res.Done:
				final := StatusSucceeded
				if run.Failed > 0 {
					final = StatusPartial
				}

				out = Finished

				return Finish(ctx, tx, id, final, nil)
			case processed >= lim.Targets || !r.Now().Before(deadline):
				out = Continued

				return r.Queue.DispatchTx(ctx, tx, Job{RunID: id})
			}

			return nil
		})
		if err != nil {
			return "", err
		}

		if out != "" {
			return out, nil
		}
	}
}

// HandleAttempt executes run id and closes it as failed when the error is permanent or last is true, so a run
// never stays running after its retries are spent.
func HandleAttempt(ctx context.Context, r *Runner, id int64, last bool) error {
	_, err := r.Execute(ctx, id)
	if err != nil && (last || errors.Is(err, queue.ErrPermanent)) {
		if ferr := Finish(context.WithoutCancel(ctx), r.DB, id, StatusFailed, err); ferr != nil {
			return errors.Join(err, ferr)
		}
	}

	return err
}

// MaxConcurrent caps simultaneous ingestion jobs whatever QUEUE_CONCURRENCY is (NFR-06-2).
// workaround(vukem): VF-018, no per-queue concurrency limit inside serve.
const MaxConcurrent = 2

// Registration is the queue handler of Job: queue ingest, 10 minute timeout, 3 retries (four attempts).
func Registration(r *Runner) queue.Registration {
	slots := make(chan struct{}, MaxConcurrent)

	return queue.Handle(func(ctx context.Context, j Job) error {
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
		case <-ctx.Done():
			return ctx.Err()
		}

		a, ok := queue.AttemptFrom(ctx)

		return HandleAttempt(ctx, r, j.RunID, !ok || a.Last())
	}, queue.OnQueue("ingest"), queue.Retries(3), queue.Timeout(10*time.Minute))
}

func cursorArg(c json.RawMessage) any {
	if len(c) == 0 {
		return nil
	}

	return string(c)
}
