//go:build integration

package ingest_test

import (
	"context"
	"errors"
	"os"
	"slices"
	"testing"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/dq"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

var errRollback = errors.New("rollback")

// dqRun applies the fixtures (none for the clean seed) inside a transaction, runs the asserts, returns the
// dq_results rows by code and rolls everything back.
func dqRun(t *testing.T, db *database.DB, a dq.Asserts, params string, fixtures ...string) map[string]dqRow {
	t.Helper()

	out := map[string]dqRow{}

	err := db.Tx(t.Context(), func(tx *database.Tx) error {
		for _, f := range append([]string{"clean"}, fixtures...) {
			sql, err := os.ReadFile("testdata/dq/" + f + ".sql")
			if err != nil {
				return err
			}

			if _, err := tx.ExecContext(t.Context(), string(sql)); err != nil {
				return errors.Join(err, errors.New(f))
			}
		}

		var id int64
		if err := tx.QueryRowContext(t.Context(), `INSERT INTO job_runs (job, mode, status, triggered_by, params)
VALUES ('dq.asserts', 'manual', 'running', 'test', $1::jsonb) RETURNING id`, params).Scan(&id); err != nil {
			return err
		}

		if _, err := a.Step(t.Context(), tx, run.Run{ID: id, Mode: run.ModeManual, Params: []byte(params)}); err != nil {
			return err
		}

		rows, err := tx.QueryContext(t.Context(), `SELECT code, severity, passed, coalesce(observed, 0), jsonb_array_length(sample)
FROM dq_results WHERE run_id = $1`, id)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var (
				code string
				r    dqRow
			)

			if err := rows.Scan(&code, &r.severity, &r.passed, &r.observed, &r.sample); err != nil {
				return err
			}

			out[code] = r
		}

		if err := rows.Err(); err != nil {
			return err
		}

		return errRollback
	})
	if !errors.Is(err, errRollback) {
		t.Fatal(err)
	}

	return out
}

type dqRow struct {
	severity string
	passed   bool
	observed float64
	sample   int
}

func dqDB(t *testing.T) *database.DB {
	t.Helper()

	db := newDB(t)
	if err := normalize.SeedDataSources(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	return db
}

func gone(context.Context, string) (bool, error) { return false, nil }

func leaky(context.Context, int) ([]string, error) { return []string{"item-with-tmdb"}, nil }

func TestDQCleanSeedPassesEverything(t *testing.T) {
	db := dqDB(t)
	got := dqRun(t, db, dq.Asserts{
		Stat:      func(context.Context, string) (bool, error) { return true, nil },
		Serialise: func(context.Context, int) ([]string, error) { return nil, nil },
	}, `{}`)

	want := append(dq.Codes(), "DQ-I-04", "DQ-I-12")
	slices.Sort(want)

	var have []string
	for code, r := range got {
		have = append(have, code)

		if !r.passed || r.observed != 0 || r.sample != 0 {
			t.Errorf("%s on the clean seed: %+v", code, r)
		}
	}

	slices.Sort(have)

	if !slices.Equal(have, want) {
		t.Errorf("codes written = %v, want %v", have, want)
	}
}

func TestDQEachCodeFailsOnItsFaultyFixture(t *testing.T) {
	db := dqDB(t)

	// codes restated by SDD 06 share the fixture of the SDD 01 code
	also := map[string][]string{
		"DQ-01": {"DQ-I-02"}, "DQ-02": {"DQ-I-03"}, "DQ-03": {"DQ-I-07"}, "DQ-04": {"DQ-I-08"}, "DQ-06": {"DQ-I-05"},
	}
	// listed, never failing: the fixture must still show up in observed and sample
	reportOnly := []string{"DQ-03", "DQ-09", "DQ-I-10"}

	codes := []string{
		"DQ-01", "DQ-02", "DQ-03", "DQ-04", "DQ-05", "DQ-06", "DQ-07", "DQ-08", "DQ-09", "DQ-10", "DQ-11", "DQ-12",
		"DQ-13", "DQ-14", "DQ-I-01", "DQ-I-04", "DQ-I-06", "DQ-I-09", "DQ-I-10", "DQ-I-11", "DQ-I-12", "DQ-I-13",
	}

	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			got := dqRun(t, db, dq.Asserts{Stat: gone, Serialise: leaky}, `{}`, code)

			for _, c := range append([]string{code}, also[code]...) {
				r, ok := got[c]
				if !ok {
					t.Fatalf("%s not written", c)
				}

				if r.observed == 0 || r.sample == 0 {
					t.Errorf("%s did not see its fixture: %+v", c, r)
				}

				if slices.Contains(reportOnly, code) != r.passed {
					t.Errorf("%s passed = %v on its faulty fixture", c, r.passed)
				}
			}

			for c, r := range got {
				if c != code && !slices.Contains(also[code], c) && !r.passed && r.severity == "error" &&
					c != "DQ-I-04" && c != "DQ-I-12" {
					t.Errorf("fixture of %s also broke the error assert %s", code, c)
				}
			}
		})
	}
}

func TestDQCountDropOverrideAndSnapshot(t *testing.T) {
	db := dqDB(t)

	if r := dqRun(t, db, dq.Asserts{}, `{}`, "DQ-I-01")["DQ-I-01"]; r.passed || r.observed < 98 {
		t.Fatalf("drop without override: %+v", r)
	}

	if r := dqRun(t, db, dq.Asserts{}, `{"allow_count_drop": true}`, "DQ-I-01")["DQ-I-01"]; r.passed {
		t.Errorf("override without a reason passed: %+v", r)
	}

	if r := dqRun(t, db, dq.Asserts{}, `{"allow_count_drop": true, "reason": "legacy purge"}`, "DQ-I-01")["DQ-I-01"]; !r.passed {
		t.Errorf("override with a reason failed: %+v", r)
	}
}

func TestDQSnapshotMovesOnlyWhenTheAssertPasses(t *testing.T) {
	db := dqDB(t)

	err := db.Tx(t.Context(), func(tx *database.Tx) error {
		if _, err := tx.ExecContext(t.Context(), `INSERT INTO catalog_counts (day, type, items) VALUES (current_date - 1, 'game', 100)`); err != nil {
			return err
		}

		if _, err := tx.ExecContext(t.Context(), `INSERT INTO items (id, type, slug, wikidata_qid) SELECT gen_random_uuid(), 'game', 'g' || n, 'Q' || (n + 100) FROM generate_series(1, 100) n`); err != nil {
			return err
		}

		var id int64
		if err := tx.QueryRowContext(t.Context(), `INSERT INTO job_runs (job, mode, status, triggered_by)
VALUES ('dq.asserts', 'manual', 'running', 'test') RETURNING id`).Scan(&id); err != nil {
			return err
		}

		if _, err := (dq.Asserts{}).Step(t.Context(), tx, run.Run{ID: id, Params: []byte(`{}`)}); err != nil {
			return err
		}

		var games int
		if err := tx.QueryRowContext(t.Context(), `SELECT items FROM catalog_counts WHERE day = current_date AND type = 'game'`).Scan(&games); err != nil {
			return err
		}

		if games != 100 {
			t.Errorf("snapshot of today = %d, want 100", games)
		}

		return errRollback
	})
	if !errors.Is(err, errRollback) {
		t.Fatal(err)
	}
}
