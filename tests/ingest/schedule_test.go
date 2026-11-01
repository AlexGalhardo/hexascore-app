//go:build integration

package ingest_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

func TestScheduleTable(t *testing.T) {
	env := map[string]string{"DATABASE_URL": pgtest.Database(t, "")}
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})

	sched, err := app.Schedule()
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"ingest:wikidata-incremental": "30 0 * * *",
		"ingest:wikidata-weekly":      "0 1 * * 0",
		"ingest:wikipedia-weekly":     "0 1 * * 1",
		"ingest:openlibrary-weekly":   "0 1 * * 2",
		"ingest:tmdb-weekly":          "0 1 * * 5",
		"ingest:trailers-weekly":      "0 4 * * 0",
		"ingest:tmdb-expire":          "0 3 * * *",
		"ingest:dq-daily":             "0 5 * * *",
	}

	for _, task := range sched.Tasks() {
		if exp, ok := want[task.Name()]; ok {
			if task.Expression() != exp {
				t.Errorf("%s = %q, want %q", task.Name(), task.Expression(), exp)
			}

			delete(want, task.Name())
		}
	}

	for name := range want {
		t.Errorf("schedule missing %s", name)
	}
}

func TestChainDispatchesNext(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	rec := &recorder{}

	first, err := ingest.WikidataIncremental.Start(ctx, db, rec, "schedule", nil)
	if err != nil {
		t.Fatal(err)
	}

	// A stage still running dispatches nothing.
	if err := ingest.Advance(ctx, db, rec, first.ID); err != nil {
		t.Fatal(err)
	}

	if n := count(t, db, `SELECT count(*) FROM job_runs`); n != 1 {
		t.Fatalf("running stage queued a successor: %d runs", n)
	}

	if err := run.Finish(ctx, db, first.ID, run.StatusSucceeded, nil); err != nil {
		t.Fatal(err)
	}

	// Twice: the redelivered job must not duplicate the successor.
	for range 2 {
		if err := ingest.Advance(ctx, db, rec, first.ID); err != nil {
			t.Fatal(err)
		}
	}

	if n := count(t, db, `SELECT count(*) FROM job_runs WHERE triggered_by = $1 AND job = 'fetch.entities'`,
		fmt.Sprintf("chain:%d", first.ID)); n != 1 {
		t.Fatalf("successor runs = %d, want 1", n)
	}

	// A failed stage stops the chain.
	var second int64
	err = db.QueryRowContext(ctx, `SELECT id FROM job_runs WHERE job = 'fetch.entities'`).Scan(&second)
	if err != nil {
		t.Fatal(err)
	}

	if err := run.Finish(ctx, db, second, run.StatusFailed, context.DeadlineExceeded); err != nil {
		t.Fatal(err)
	}

	if err := ingest.Advance(ctx, db, rec, second); err != nil {
		t.Fatal(err)
	}

	if n := count(t, db, `SELECT count(*) FROM job_runs`); n != 2 {
		t.Fatalf("failed stage queued a successor: %d runs", n)
	}
}
