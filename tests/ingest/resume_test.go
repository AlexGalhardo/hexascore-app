//go:build integration

package ingest_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

// snapshot is every observable table state that must not depend on interruptions.
func snapshot(t *testing.T, db *database.DB) string {
	t.Helper()

	rows, err := db.QueryContext(context.Background(), `SELECT 'rec', source||external_id||hash FROM source_records
UNION ALL SELECT 'run', status||seen||'/'||changed||'/'||failed FROM job_runs ORDER BY 1, 2`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var out strings.Builder
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			t.Fatal(err)
		}

		out.WriteString(k + ":" + v + "\n")
	}

	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	return out.String()
}

// drive executes the run and its successors until it stops continuing, as the queue would.
func drive(t *testing.T, r *run.Runner, id int64) {
	t.Helper()

	for {
		out, err := r.Execute(t.Context(), id)
		if err != nil {
			t.Fatal(err)
		}

		if out != run.Continued {
			return
		}
	}
}

func TestResumeAfterCancel(t *testing.T) {
	const total = 95

	lim := run.Limits{Targets: 20}
	spec := run.Spec{Job: "fetch", Source: "wikidata", Mode: run.ModeFull, TriggeredBy: "test"}

	// Uninterrupted reference.
	refDB := newDB(t)
	refRec := &recorder{}
	ref := newRunner(refDB, refRec, newFlags(t, allOn), lim, map[string]run.Step{"fetch": upsertStep(total, 10, nil)})

	rr, err := run.Create(t.Context(), refDB, refRec, spec)
	if err != nil {
		t.Fatal(err)
	}

	drive(t, ref, rr.ID)

	want := snapshot(t, refDB)
	if n := count(t, refDB, `SELECT count(*) FROM source_records`); n != total {
		t.Fatalf("reference stored %d records, want %d", n, total)
	}

	// Cancelled after a random chunk, then resumed.
	db := newDB(t)
	rec := &recorder{}
	cancelAt := int(time.Now().UnixNano()%8) + 1
	calls := 0
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	step := upsertStep(total, 10, func() {
		calls++
		if calls == cancelAt {
			cancel()
		}
	})
	r := newRunner(db, rec, newFlags(t, allOn), lim, map[string]run.Step{"fetch": step})

	cr, err := run.Create(t.Context(), db, rec, spec)
	if err != nil {
		t.Fatal(err)
	}

	for {
		out, err := r.Execute(ctx, cr.ID)
		if err != nil {
			break // interrupted: the chunk in flight rolled back with its cursor
		}

		if out != run.Continued {
			break
		}
	}

	cancel()
	drive(t, r, cr.ID)

	if got := snapshot(t, db); got != want {
		t.Fatalf("resumed run differs from uninterrupted (cancel at chunk %d):\n%s\nwant:\n%s", cancelAt, got, want)
	}
}

func TestContinuationDispatchesSuccessorAtTargetBudget(t *testing.T) {
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

	rec.jobs = nil

	out, err := r.Execute(t.Context(), cr.ID)
	if err != nil || out != run.Continued {
		t.Fatalf("Execute = %v, %v; want Continued", out, err)
	}

	if len(rec.jobs) != 1 || rec.jobs[0].(run.Job).RunID != cr.ID {
		t.Fatalf("successors = %v, want one for run %d", rec.jobs, cr.ID)
	}

	if n := count(t, db, `SELECT count(*) FROM source_records`); n != 20 {
		t.Errorf("records after first chunked execute = %d, want 20", n)
	}
}
