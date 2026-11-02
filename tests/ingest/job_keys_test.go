//go:build integration

package ingest_test

import (
	"errors"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// A second automatic dispatch of the same stage on the same day is refused (and leaves no job_runs row), a manual
// one is not, a failed run frees its key and a done key older than 24 h frees it too.
func TestIngestJobKeys_DuplicateDispatchIsNoop(t *testing.T) {
	db := newDB(t)
	rec := &recorder{}
	spec := run.Spec{Job: "fetch", Source: "wikidata", Mode: run.ModeIncremental, TriggeredBy: "schedule"}

	first, err := run.Create(t.Context(), db, rec, spec)
	if err != nil {
		t.Fatal(err)
	}

	// job_runs_one_active already refuses a twin of a live run; the key covers the finished ones.
	if _, err := run.Create(t.Context(), db, rec, spec); !errors.Is(err, run.ErrRunActive) {
		t.Fatalf("queued twin: err = %v, want ErrRunActive", err)
	}

	if err := run.Finish(t.Context(), db, first.ID, run.StatusSucceeded, nil); err != nil {
		t.Fatal(err)
	}

	if _, err := run.Create(t.Context(), db, rec, spec); !errors.Is(err, run.ErrDuplicate) {
		t.Fatalf("done twin within 24 h: err = %v, want ErrDuplicate", err)
	}

	if n := count(t, db, `SELECT count(*) FROM job_runs`); n != 1 {
		t.Fatalf("job_runs rows = %d, want 1 (the refused twin must roll back)", n)
	}

	if len(rec.jobs) != 1 {
		t.Fatalf("dispatched %d jobs, want 1", len(rec.jobs))
	}

	manual := spec
	manual.TriggeredBy = "admin:1"

	if _, err := run.Create(t.Context(), db, rec, manual); err != nil {
		t.Fatalf("manual start: %v", err)
	}

	if _, err := db.ExecContext(
		t.Context(),
		`UPDATE job_runs SET status = 'succeeded', finished_at = now()`,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(
		t.Context(),
		`UPDATE ingest_job_keys SET done_at = now() - interval '25 hours'`,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := run.Create(t.Context(), db, rec, spec); err != nil {
		t.Fatalf("done key older than 24 h: %v", err)
	}
}

func TestIngestJobKeys_FailedRunFreesKeyAndPruneDropsOld(t *testing.T) {
	db := newDB(t)
	rec := &recorder{}
	spec := run.Spec{Job: "fetch", Source: "wikidata", Mode: run.ModeFull, TriggeredBy: "chain:1"}

	r, err := run.Create(t.Context(), db, rec, spec)
	if err != nil {
		t.Fatal(err)
	}

	if err := run.Finish(t.Context(), db, r.ID, run.StatusFailed, errors.New("boom")); err != nil {
		t.Fatal(err)
	}

	if _, err := run.Create(t.Context(), db, rec, spec); err != nil {
		t.Fatalf("after a failed run: %v", err)
	}

	if _, err := db.ExecContext(
		t.Context(),
		`UPDATE ingest_job_keys SET created_at = now() - interval '8 days'`,
	); err != nil {
		t.Fatal(err)
	}

	p, err := run.Prune(t.Context(), db)
	if err != nil || p.Keys != 1 {
		t.Fatalf("Prune = %+v, %v, want 1 key", p, err)
	}
}
