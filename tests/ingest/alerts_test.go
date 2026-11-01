//go:build integration

package ingest_test

import (
	"bytes"
	"context"
	"database/sql"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
)

// failFirst makes the failed_run check error, as a broken query would.
type failFirst struct{ database.Executor }

func (f failFirst) QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row {
	if strings.Contains(q, "status = 'failed'") {
		return f.Executor.QueryRowContext(ctx, `SELECT 1/0`)
	}

	return f.Executor.QueryRowContext(ctx, q, args...)
}

func TestAlertA7ContinuesPastAFailingCheck(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()

	if _, err := db.ExecContext(ctx, `INSERT INTO source_records (source, external_id, payload, hash, checked_at)
VALUES ('tmdb', 'Q1', '{}', repeat('a', 64), now() - interval '171 days')`); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	n, err := ingest.CheckA7(ctx, failFirst{db}, slog.New(slog.NewTextHandler(&buf, nil)), time.Now(), true)
	if err == nil {
		t.Fatal("want the failing check's error")
	}

	out := buf.String()
	if !strings.Contains(out, "check=failed_run") || !strings.Contains(out, "state=unknown") {
		t.Errorf("the failing check must log state=unknown: %s", out)
	}

	if !strings.Contains(out, "check=tmdb_age") || n < 1 {
		t.Errorf("the checks after the failing one must still run (raised %d): %s", n, out)
	}
}

func TestAlertA7Lines(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()

	check := func(enabled bool) (int, string) {
		var buf bytes.Buffer

		n, err := ingest.CheckA7(ctx, db, slog.New(slog.NewTextHandler(&buf, nil)), time.Now(), enabled)
		if err != nil {
			t.Fatal(err)
		}

		return n, buf.String()
	}

	// A healthy day: an incremental run finished an hour ago.
	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO job_runs (job, source, mode, status, triggered_by, started_at, finished_at)
VALUES ('fetch.wikidata', 'wikidata', 'incremental', 'succeeded', 'schedule', now() - interval '2 hours', now() - interval '1 hour')`,
	); err != nil {
		t.Fatal(err)
	}

	if n, out := check(true); n != 0 || out != "" {
		t.Fatalf("healthy state raised %d alerts: %s", n, out)
	}

	var run int64
	if err := db.QueryRowContext(ctx, `INSERT INTO job_runs (job, mode, status, triggered_by, started_at, finished_at)
VALUES ('dq.asserts', 'full', 'failed', 'schedule', now(), now()) RETURNING id`).Scan(&run); err != nil {
		t.Fatal(err)
	}

	for _, s := range []string{
		`INSERT INTO dq_results (run_id, code, severity, passed, observed) VALUES (` + itoa(int(run)) + `, 'DQ-01', 'error', false, 3)`,
		`INSERT INTO source_records (source, external_id, payload, hash, checked_at)
VALUES ('tmdb', 'Q1', '{}', repeat('a', 64), now() - interval '171 days')`,
	} {
		if _, err := db.ExecContext(ctx, s); err != nil {
			t.Fatal(err)
		}
	}

	if n, out := check(false); n != 0 || out != "" {
		t.Errorf("a disabled ingestion must stay silent, got %d: %s", n, out)
	}

	n, out := check(true)
	if n != 3 {
		t.Errorf("alerts = %d, want 3 (failed run, error assert, old TMDB): %s", n, out)
	}

	for _, want := range []string{"alert=A7", "check=failed_run", "check=error_assert", "check=tmdb_age"} {
		if !strings.Contains(out, want) {
			t.Errorf("log misses %q: %s", want, out)
		}
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM job_runs WHERE job = 'fetch.wikidata'`); err != nil {
		t.Fatal(err)
	}

	if _, out := check(true); !strings.Contains(out, "check=no_incremental") {
		t.Errorf("no incremental in 26 h must alert: %s", out)
	}
}
