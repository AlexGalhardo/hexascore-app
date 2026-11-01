//go:build integration

package ingest_test

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

func TestPruneRetention(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()

	runAt := func(age string) int64 {
		var id int64
		if err := db.QueryRowContext(ctx, `INSERT INTO job_runs (job, mode, status, triggered_by, created_at)
VALUES ('dq.asserts', 'full', 'succeeded', 'schedule', now() - $1::interval) RETURNING id`, age).Scan(&id); err != nil {
			t.Fatal(err)
		}

		return id
	}

	old, mid, fresh, ancient := runAt("200 days"), runAt("100 days"), runAt("1 day"), runAt("400 days")
	_ = fresh

	for _, id := range []int64{old, ancient} {
		if _, err := db.ExecContext(
			ctx,
			`INSERT INTO dq_results (run_id, code, severity, passed) VALUES ($1, 'DQ-01', 'error', true)`,
			id,
		); err != nil {
			t.Fatal(err)
		}

		if _, err := db.ExecContext(
			ctx,
			`INSERT INTO field_coverage (run_id, type, field, sampled, missing) VALUES ($1, 'movie', 'director', 10, 1)`,
			id,
		); err != nil {
			t.Fatal(err)
		}
	}

	for _, age := range []int{31, 5} {
		if _, err := db.ExecContext(
			ctx,
			`INSERT INTO source_record_history (source, external_id, hash, payload, fetched_at, superseded_at, superseded_by_run)
VALUES ('wikidata', 'Q' || $3::text, repeat('a', 64), '{}', now() - interval '40 days', now() - make_interval(days => $1), $2)`,
			age,
			mid,
			itoa(age),
		); err != nil {
			t.Fatal(err)
		}
	}

	got, err := run.Prune(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	if got.History != 1 || got.Results != 1 || got.Coverage != 1 || got.Runs != 1 {
		t.Errorf("pruned = %+v, want 1 history, 1 result, 1 coverage, 1 run", got)
	}

	// The 200-day run still holds its asserts (kept 365 days), so it outlives the 180-day cutoff.
	if n := count(
		t,
		db,
		`SELECT count(*) FROM job_runs WHERE id = ANY($1)`,
		"{"+itoa(int(old))+","+itoa(int(mid))+"}",
	); n != 2 {
		t.Errorf("runs left of old and mid = %d, want 2", n)
	}

	if n := count(t, db, `SELECT count(*) FROM job_runs WHERE id = $1`, ancient); n != 0 {
		t.Error("a 400-day run must be gone")
	}

	if n := count(t, db, `SELECT count(*) FROM source_record_history`); n != 1 {
		t.Errorf("history left = %d, want the 5-day row", n)
	}
}

func TestRetryFailuresQueuesRunsPerSource(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()
	rec := &recorder{}
	ff := newFlags(t, map[string]string{
		"FLAG_INGEST_ENABLED": "true", "FLAG_INGEST_TMDB": "true", "FLAG_INGEST_WIKIDATA": "true",
	})

	var failing int64
	if err := db.QueryRowContext(ctx, `INSERT INTO job_runs (job, mode, status, triggered_by) VALUES ('fetch.tmdb','full','partial','schedule')
RETURNING id`).
		Scan(&failing); err != nil {
		t.Fatal(err)
	}

	for _, f := range []struct {
		source, id string
		parked     bool
	}{
		{"tmdb", "Q1", false},
		{"tmdb", "Q2", true},
		{"tmdb", "Q6", false},
		{"normalize", "Q3", false},
		{"wikipedia:en", "Q4", false}, // kill switch off
		{"editions", "9780000000002", false},
	} {
		if _, err := db.ExecContext(
			ctx,
			`INSERT INTO ingest_failures (source, external_id, last_run_id, error_class, error, parked)
VALUES ($1, $2, $3, 'fetch', 'boom', $4)`,
			f.source,
			f.id,
			failing,
			f.parked,
		); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO source_records (source, external_id, payload, hash, normalized_at)
VALUES ('wikidata', 'Q3', '{}', repeat('a', 64), now())`); err != nil {
		t.Fatal(err)
	}

	n, err := run.RetryFailures(ctx, db, rec, ff)
	if err != nil {
		t.Fatal(err)
	}

	if n != 2 {
		t.Errorf("runs queued = %d, want 2 (tmdb and normalize)", n)
	}

	var params json.RawMessage
	if err := db.QueryRowContext(ctx, `SELECT params FROM job_runs WHERE job = 'fetch.tmdb' AND triggered_by = 'schedule:retry'`).
		Scan(&params); err != nil {
		t.Fatalf("no retry run for tmdb: %v", err)
	}

	var p struct{ QIDs []string }
	if err := json.Unmarshal(params, &p); err != nil || !slices.Equal(p.QIDs, []string{"Q1", "Q6"}) {
		t.Errorf("qids = %v, %v; want [Q1 Q6] (Q2 is parked)", p.QIDs, err)
	}

	if c := count(t, db, `SELECT count(*) FROM job_runs WHERE job = 'fetch.wikipedia:en'`); c != 0 {
		t.Error("a source whose kill switch is off must not be retried")
	}

	if c := count(
		t,
		db,
		`SELECT count(*) FROM source_records WHERE external_id = 'Q3' AND normalized_at IS NULL`,
	); c != 1 {
		t.Error("a normalise failure must put its record back in the pending index")
	}

	if c := count(
		t,
		db,
		`SELECT count(*) FROM job_runs WHERE job = 'normalize' AND triggered_by = 'schedule:retry'`,
	); c != 1 {
		t.Error("a normalize run must be queued for the normalise failures")
	}

	again, err := run.RetryFailures(ctx, db, rec, ff)
	if err != nil || again != 0 {
		t.Errorf("second pass = %d, %v; want 0 (runs still active)", again, err)
	}
}
