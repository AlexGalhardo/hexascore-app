//go:build integration

package ingest_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

func TestStaleRunFailsAndResumesFromCursor(t *testing.T) {
	db := newDB(t)
	rec := &recorder{}
	r := newRunner(
		db,
		rec,
		newFlags(t, allOn),
		run.Limits{Targets: 20},
		map[string]run.Step{"fetch": upsertStep(50, 10, nil)},
	)

	cr, err := run.Create(
		t.Context(),
		db,
		rec,
		run.Spec{Job: "fetch", Source: "wikidata", Mode: run.ModeFull, TriggeredBy: "test"},
	)
	if err != nil {
		t.Fatal(err)
	}

	if out, err := r.Execute(t.Context(), cr.ID); err != nil || out != run.Continued {
		t.Fatalf("Execute = %v, %v", out, err)
	}

	if n, err := run.ReapStale(t.Context(), db); err != nil || n != 0 {
		t.Fatalf("fresh heartbeat reaped: %d, %v", n, err)
	}

	if _, err := db.ExecContext(
		t.Context(),
		`UPDATE job_runs SET heartbeat_at = now() - interval '16 minutes'`,
	); err != nil {
		t.Fatal(err)
	}

	if n, err := run.ReapStale(t.Context(), db); err != nil || n != 1 {
		t.Fatalf("ReapStale = %d, %v; want 1", n, err)
	}

	if n := count(t, db, `SELECT count(*) FROM job_runs WHERE id = $1 AND status = 'failed'`, cr.ID); n != 1 {
		t.Fatal("stale run not failed")
	}

	next, err := run.Resume(t.Context(), db, rec, cr.ID)
	if err != nil {
		t.Fatal(err)
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM job_runs WHERE id = $1 AND mode = 'followup' AND cursor->>'next' = '21'`,
		next.ID,
	); n != 1 {
		t.Error("resumed run does not carry the cursor")
	}

	drive(t, r, next.ID)

	if n := count(t, db, `SELECT count(*) FROM source_records`); n != 50 {
		t.Errorf("records = %d, want 50", n)
	}
}

func TestPermanentErrorFailsRunWithoutRetry(t *testing.T) {
	db := newDB(t)
	rec := &recorder{}
	boom := fmt.Errorf("tmdb 401: %w", queue.ErrPermanent)
	step := func(context.Context, *database.Tx, run.Run) (run.Result, error) { return run.Result{}, boom }
	r := newRunner(db, rec, newFlags(t, allOn), run.Limits{}, map[string]run.Step{"fetch": step})

	cr, err := run.Create(
		t.Context(),
		db,
		rec,
		run.Spec{Job: "fetch", Source: "wikidata", Mode: run.ModeFull, TriggeredBy: "test"},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = run.HandleAttempt(t.Context(), r, cr.ID, false)
	if !errors.Is(err, queue.ErrPermanent) {
		t.Fatalf("Handle err = %v, want ErrPermanent", err)
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM job_runs WHERE id = $1 AND status = 'failed' AND error LIKE '%401%'`,
		cr.ID,
	); n != 1 {
		t.Error("run not closed as failed with the error")
	}
}

func TestTransientErrorKeepsRunOpenUntilLastAttempt(t *testing.T) {
	db := newDB(t)
	rec := &recorder{}
	step := func(context.Context, *database.Tx, run.Run) (run.Result, error) {
		return run.Result{}, errors.New("boom")
	}
	r := newRunner(db, rec, newFlags(t, allOn), run.Limits{}, map[string]run.Step{"fetch": step})

	cr, err := run.Create(
		t.Context(),
		db,
		rec,
		run.Spec{Job: "fetch", Source: "wikidata", Mode: run.ModeFull, TriggeredBy: "test"},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := run.HandleAttempt(t.Context(), r, cr.ID, false); err == nil {
		t.Fatal("want error")
	}

	if n := count(t, db, `SELECT count(*) FROM job_runs WHERE id = $1 AND status = 'running'`, cr.ID); n != 1 {
		t.Fatal("run closed before the last attempt")
	}

	if err := run.HandleAttempt(t.Context(), r, cr.ID, true); err == nil {
		t.Fatal("want error")
	}

	if n := count(t, db, `SELECT count(*) FROM job_runs WHERE id = $1 AND status = 'failed'`, cr.ID); n != 1 {
		t.Error("run not closed on the last attempt")
	}
}

func TestFailuresParkAfterFiveRuns(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()

	for runID := int64(1); runID <= 5; runID++ {
		// Two failures inside one run count once.
		for range 2 {
			if _, err := run.RecordFailure(ctx, db, runID, "wikidata", "Q1", "http", errors.New("500")); err != nil {
				t.Fatal(err)
			}
		}

		parked, err := run.IsParked(ctx, db, "wikidata", "Q1")
		if err != nil {
			t.Fatal(err)
		}

		if want := runID == 5; parked != want {
			t.Fatalf("after run %d parked = %v, want %v", runID, parked, want)
		}
	}

	if err := run.ClearFailure(ctx, db, "wikidata", "Q1"); err != nil {
		t.Fatal(err)
	}

	if parked, _ := run.IsParked(ctx, db, "wikidata", "Q1"); parked {
		t.Error("cleared failure still parked")
	}
}
