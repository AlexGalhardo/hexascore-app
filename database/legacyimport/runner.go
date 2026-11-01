// Package legacyimport moves the restored legacy dump (schema legacy_import,
// SDD 11 §7.6 steps 3-5) into the app tables, one stage at a time.
package legacyimport

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"slices"
	"sync"
	"time"

	"github.com/AlexGalhardo/vukem/auth/hash"
	"github.com/AlexGalhardo/vukem/database"
)

// Schema is where pg_restore puts the legacy tables.
const Schema = "legacy_import"

// DefaultBatch is the rows per transaction.
const DefaultBatch = 1000

// ErrNotImplemented marks a stage whose import is a later task.
var ErrNotImplemented = errors.New("legacyimport: stage not implemented yet")

// BatchFunc imports up to n source rows whose key sorts after `after` ("" for
// the first batch), inside tx, and returns the last key it imported and how
// many rows. Zero rows means the stage is complete. It must be idempotent on
// its target rows (upserts), because a batch can be retried after a crash
// between its commit and nothing else: the checkpoint commits with it.
type BatchFunc func(ctx context.Context, tx *database.Tx, after string, n int) (last string, rows int, err error)

// Stage is one import step.
type Stage struct {
	Name string
	// Tables are the legacy tables the stage reads, counted by the dry run.
	Tables []string
	// Batch is nil until the stage is implemented (P2.F.2.2 to P2.F.4.1).
	Batch BatchFunc
	// KeysSQL, when set, lets Runner.Workers split the stage: it takes $1 = checkpoint key ("" at the start)
	// and $2 = batch size, and returns the key ending every full batch after the checkpoint, in order. Such
	// a stage's Batch must be safe to run concurrently on disjoint key ranges.
	KeysSQL string
	// Run replaces Batch for a stage that is not a keyed copy (finalise): it
	// runs once, outside a batch transaction, must be idempotent and returns
	// the rows it touched.
	Run func(ctx context.Context, db *database.DB, now time.Time) (int64, error)
}

// Stages lists the stages in their mandatory order (SDD 11 §7.6 step 5):
// users before everything that references them, catalog before entities
// and social, finalise last.
func Stages() []Stage {
	return []Stage{
		{Name: "users", Tables: []string{"user", "account"}, Batch: usersBatch(hash.Default(), time.Now)},
		{Name: "twofactor", Tables: []string{"two_factor"}},
		{
			Name:   "catalog",
			Tables: []string{"items", "item_translations", "source_records", "affiliate_links"},
			Batch:  catalogBatch(time.Now),
		},
		{Name: "entities", Tables: []string{"items"}, Batch: entitiesBatch(), KeysSQL: entityKeysSQL},
		{
			Name:   "social",
			Tables: []string{"ratings", "lists", "list_items", "reports", "contact_messages"},
			Batch:  socialBatch(time.Now), KeysSQL: userKeysSQL,
		},
		{
			Name:    "library",
			Tables:  []string{"item_status", "time_to_beat"},
			Batch:   libraryBatch(time.Now),
			KeysSQL: userKeysSQL,
		},
		{Name: "billing", Tables: []string{
			"api_keys", "api_usage_daily", "subscriptions", "stripe_customers", "credit_ledger", "pix_charges",
		}, Batch: billingBatch(time.Now)},
		{Name: "finalise", Run: finalise},
	}
}

// Select returns the stages named by spec ("all" or a comma list), in the
// mandatory order whatever order spec uses.
func Select(all []Stage, names []string) ([]Stage, error) {
	if len(names) == 1 && names[0] == "all" {
		return all, nil
	}

	var out []Stage

	for _, n := range names {
		if !slices.ContainsFunc(all, func(s Stage) bool { return s.Name == n }) {
			return nil, fmt.Errorf("legacyimport: unknown stage %q", n)
		}
	}

	for _, s := range all {
		if slices.Contains(names, s.Name) {
			out = append(out, s)
		}
	}

	return out, nil
}

// TableCount is one line of the dry-run plan.
type TableCount struct {
	Stage, Table string
	Rows         int64
}

// Plan counts the source rows each stage would read. It only reads.
func Plan(ctx context.Context, db *database.DB, stages []Stage) ([]TableCount, error) {
	var out []TableCount

	for _, s := range stages {
		for _, t := range s.Tables {
			var n int64
			// Table names are constants from Stages, never input.
			q := fmt.Sprintf(`SELECT count(*) FROM %s.%q`, Schema, t)
			if err := db.QueryRowContext(ctx, q).Scan(&n); err != nil {
				return nil, fmt.Errorf("legacyimport: count %s.%s: %w", Schema, t, err)
			}

			out = append(out, TableCount{Stage: s.Name, Table: t, Rows: n})
		}
	}

	return out, nil
}

// Runner runs stages in order with a checkpoint per stage.
type Runner struct {
	DB    *database.DB
	Batch int
	Out   io.Writer
	Now   func() time.Time
	// Workers is how many batches of a stage with KeysSQL run at once, each in its own transaction on its own
	// connection (the pool must allow that many). 0 or 1 is sequential.
	Workers int
}

type checkpoint struct {
	status, lastKey string
	rows            int64
}

// Run imports the stages in order. A stage already done is skipped; a stage
// interrupted earlier resumes after its last committed key.
func (r *Runner) Run(ctx context.Context, stages []Stage) error {
	for _, s := range stages {
		if err := r.runStage(ctx, s); err != nil {
			r.fail(ctx, s.Name, err)
			return err
		}
	}

	return nil
}

func (r *Runner) runStage(ctx context.Context, s Stage) error {
	cp, err := r.load(ctx, s.Name)
	if err != nil {
		return err
	}

	if cp.status == "done" {
		_, _ = fmt.Fprintf(r.Out, "%s: done earlier (%d rows), skipped\n", s.Name, cp.rows)
		return nil
	}

	if s.Batch == nil && s.Run == nil {
		return fmt.Errorf("%w: %s", ErrNotImplemented, s.Name)
	}

	now := r.Now().UTC()
	if _, err := r.DB.ExecContext(ctx, `INSERT INTO legacy_import_runs (stage, status, started_at, updated_at)
VALUES ($1, 'running', $2, $2) ON CONFLICT (stage) DO UPDATE SET status = 'running', error = NULL, updated_at = $2`,
		s.Name, now); err != nil {
		return fmt.Errorf("legacyimport: checkpoint %s: %w", s.Name, err)
	}

	if cp.lastKey != "" {
		_, _ = fmt.Fprintf(r.Out, "%s: resuming after key %s (%d rows so far)\n", s.Name, cp.lastKey, cp.rows)
	}

	after, total := cp.lastKey, cp.rows

	if s.Run != nil {
		n, err := s.Run(ctx, r.DB, now)
		if err != nil {
			return fmt.Errorf("legacyimport: %s: %w", s.Name, err)
		}

		total = n
		if _, err := r.DB.ExecContext(
			ctx,
			`UPDATE legacy_import_runs SET rows = $2 WHERE stage = $1`,
			s.Name,
			n,
		); err != nil {
			return fmt.Errorf("legacyimport: checkpoint %s: %w", s.Name, err)
		}
	}

	if s.Run == nil && r.Workers > 1 && s.KeysSQL != "" {
		var n int64

		if after, n, err = r.runParallel(ctx, s, after); err != nil {
			return fmt.Errorf("legacyimport: %s after key %q: %w", s.Name, after, err)
		}

		total += n
	}

	// With workers this only runs the tail (a last short batch) and the empty batch that ends the stage.
	for s.Run == nil {
		var n int

		err := r.DB.Tx(ctx, func(tx *database.Tx) error {
			last, rows, err := s.Batch(ctx, tx, after, r.Batch)
			if err != nil || rows == 0 {
				n = 0
				return err
			}

			n = rows

			_, err = tx.ExecContext(ctx, `UPDATE legacy_import_runs SET last_key = $2, rows = rows + $3, updated_at = $4
WHERE stage = $1`, s.Name, last, rows, r.Now().UTC())
			if err == nil {
				after = last
			}

			return err
		})
		if err != nil {
			return fmt.Errorf("legacyimport: %s after key %q: %w", s.Name, after, err)
		}

		if n == 0 {
			break
		}

		total += int64(n)
	}

	if _, err := r.DB.ExecContext(ctx, `UPDATE legacy_import_runs SET status = 'done', finished_at = $2, updated_at = $2
WHERE stage = $1`, s.Name, r.Now().UTC()); err != nil {
		return fmt.Errorf("legacyimport: checkpoint %s: %w", s.Name, err)
	}

	_, _ = fmt.Fprintf(r.Out, "%s: done, %d rows\n", s.Name, total)

	return nil
}

// runParallel imports the full batches after `after` with r.Workers goroutines. The checkpoint moves only
// to the end of the longest prefix of batches that are all committed, so a crash re-runs (idempotently)
// the batches finished out of order and never skips one. It returns the checkpoint it reached.
func (r *Runner) runParallel(ctx context.Context, s Stage, after string) (string, int64, error) {
	bounds, err := r.boundaries(ctx, s, after)
	if err != nil || len(bounds) == 0 {
		return after, 0, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		firstErr error
		next     int
		done     = make([]int, len(bounds))
		reached  = after
		moved    int64
		jobs     = make(chan int)
	)

	finish := func(j, rows int) error {
		mu.Lock()
		defer mu.Unlock()

		done[j] = rows

		var add int64

		for next < len(bounds) && done[next] > 0 {
			add += int64(done[next])
			reached = bounds[next]
			next++
		}

		if add == 0 {
			return nil
		}

		moved += add

		_, err := r.DB.ExecContext(ctx, `UPDATE legacy_import_runs SET last_key = $2, rows = rows + $3, updated_at = $4
WHERE stage = $1`, s.Name, reached, add, r.Now().UTC())

		return err
	}

	fail := func(err error) {
		mu.Lock()
		defer mu.Unlock()

		if firstErr == nil {
			firstErr = err
		}

		cancel()
	}

	for range r.Workers {
		wg.Go(func() {
			for j := range jobs {
				lo := after
				if j > 0 {
					lo = bounds[j-1]
				}

				var rows int

				err := r.DB.Tx(ctx, func(tx *database.Tx) error {
					last, n, err := s.Batch(ctx, tx, lo, r.Batch)
					if err == nil && last != bounds[j] {
						err = fmt.Errorf("batch after %q ended at %q, expected %q", lo, last, bounds[j])
					}

					rows = n

					return err
				})
				if err == nil {
					err = finish(j, rows)
				}

				if err != nil {
					fail(err)
					return
				}
			}
		})
	}

feed:
	for j := range bounds {
		select {
		case jobs <- j:
		case <-ctx.Done():
			break feed
		}
	}

	close(jobs)
	wg.Wait()

	return reached, moved, firstErr
}

func (r *Runner) boundaries(ctx context.Context, s Stage, after string) ([]string, error) {
	rows, err := r.DB.QueryContext(ctx, s.KeysSQL, after, r.Batch)
	if err != nil {
		return nil, fmt.Errorf("legacyimport: %s batch boundaries: %w", s.Name, err)
	}

	defer func() { _ = rows.Close() }()

	var out []string

	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}

		out = append(out, k)
	}

	return out, rows.Err()
}

func (r *Runner) load(ctx context.Context, stage string) (checkpoint, error) {
	var (
		cp  checkpoint
		key sql.NullString
	)

	err := r.DB.QueryRowContext(ctx, `SELECT status, last_key, rows FROM legacy_import_runs WHERE stage = $1`, stage).
		Scan(&cp.status, &key, &cp.rows)
	if errors.Is(err, sql.ErrNoRows) {
		return checkpoint{}, nil
	}

	if err != nil {
		return cp, fmt.Errorf("legacyimport: load checkpoint %s: %w", stage, err)
	}

	cp.lastKey = key.String

	return cp, nil
}

// fail records the error; the checkpoint keeps the last committed key.
func (r *Runner) fail(ctx context.Context, stage string, cause error) {
	msg := cause.Error()
	_, _ = r.DB.ExecContext(context.WithoutCancel(ctx), `UPDATE legacy_import_runs SET status = 'failed', error = $2,
updated_at = $3 WHERE stage = $1`, stage, msg[:min(len(msg), 2000)], r.Now().UTC())
}
