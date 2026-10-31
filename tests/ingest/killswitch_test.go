//go:build integration

package ingest_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

func TestKillSwitchSkips(t *testing.T) {
	cases := map[string]map[string]string{
		"code default (fail closed)": {},
		"global off":                 {"FLAG_INGEST_WIKIDATA": "true"},
		"source off":                 {"FLAG_INGEST_ENABLED": "true"},
		"explicit false":             {"FLAG_INGEST_ENABLED": "true", "FLAG_INGEST_WIKIDATA": "false"},
	}

	for name, env := range cases {
		t.Run(name, func(t *testing.T) {
			var hits atomic.Int32

			srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits.Add(1) }))
			defer srv.Close()

			db := newDB(t)
			rec := &recorder{}
			step := func(_ context.Context, _ *database.Tx, _ run.Run) (run.Result, error) {
				resp, err := http.Get(srv.URL)
				if err == nil {
					resp.Body.Close()
				}

				return run.Result{Done: true}, err
			}
			r := newRunner(db, rec, newFlags(t, env), run.Limits{}, map[string]run.Step{"fetch": step})

			cr, err := run.Create(
				t.Context(),
				db,
				rec,
				run.Spec{Job: "fetch", Source: "wikidata", Mode: run.ModeFull, TriggeredBy: "test"},
			)
			if err != nil {
				t.Fatal(err)
			}

			out, err := r.Execute(t.Context(), cr.ID)
			if err != nil || out != run.Skipped {
				t.Fatalf("Execute = %v, %v; want Skipped", out, err)
			}

			if hits.Load() != 0 {
				t.Errorf("source hit %d times, want 0", hits.Load())
			}

			if n := count(
				t,
				db,
				`SELECT count(*) FROM job_runs WHERE id = $1 AND status = 'skipped' AND finished_at IS NOT NULL`,
				cr.ID,
			); n != 1 {
				t.Errorf("run not marked skipped")
			}
		})
	}
}

func TestOneActiveRunPerJob(t *testing.T) {
	db := newDB(t)
	rec := &recorder{}
	spec := run.Spec{Job: "fetch", Source: "wikidata", Mode: run.ModeFull, TriggeredBy: "test"}

	first, err := run.Create(t.Context(), db, rec, spec)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := run.Create(t.Context(), db, rec, spec); !errors.Is(err, run.ErrRunActive) {
		t.Fatalf("second Create err = %v, want ErrRunActive", err)
	}

	if len(rec.jobs) != 1 {
		t.Errorf("dispatched %d jobs, want 1 (the refused run must not enqueue)", len(rec.jobs))
	}

	other := spec
	other.Source = "wikipedia"
	if _, err := run.Create(t.Context(), db, rec, other); err != nil {
		t.Errorf("a different source must not conflict: %v", err)
	}

	if err := run.Finish(t.Context(), db, first.ID, run.StatusSucceeded, nil); err != nil {
		t.Fatal(err)
	}

	if _, err := run.Create(t.Context(), db, rec, spec); err != nil {
		t.Errorf("after the first run ended: %v", err)
	}
}
