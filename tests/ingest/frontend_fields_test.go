//go:build integration

package ingest_test

import (
	"errors"
	"testing"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/dq"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

const completeMovie = `
INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata','Wikidata','CC0-1.0','https://www.wikidata.org')
  ON CONFLICT DO NOTHING;
INSERT INTO items (id, type, slug, wikidata_qid, original_title, release_date, release_precision, runtime_minutes, cover_key)
VALUES ('0192f1c6-0000-7000-8000-0000000000a1', 'movie', 'complete', 'Q9001', 'Complete', '2020-01-01', 'day', 100, 'covers/a1.webp');
INSERT INTO item_translations (item_id, locale, title, title_source, synopsis, synopsis_source, synopsis_licence)
VALUES ('0192f1c6-0000-7000-8000-0000000000a1', 'en', 'Complete', 'wikidata', 'A story.', 'wikidata', 'CC0-1.0');
INSERT INTO entities (id, kind, slug, default_name) VALUES ('0192f1c6-0000-7000-8000-0000000000a2', 'genre', 'drama', 'Drama');
INSERT INTO item_entities (item_id, entity_id, role, source)
VALUES ('0192f1c6-0000-7000-8000-0000000000a1', '0192f1c6-0000-7000-8000-0000000000a2', 'genre', 'wikidata');`

func frontendRun(t *testing.T, tx *database.Tx, db *database.DB) run.Result {
	t.Helper()

	var id int64
	if err := tx.QueryRowContext(t.Context(), `INSERT INTO job_runs (job, mode, status, triggered_by)
VALUES ('dq.frontend_fields', 'manual', 'running', 'test') RETURNING id`).Scan(&id); err != nil {
		t.Fatal(err)
	}

	res, err := dq.FrontendFields{Suggest: community.IngestGate{DB: db}.OpenMissingField}.
		Step(t.Context(), tx, run.Run{ID: id, Mode: run.ModeManual})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tx.ExecContext(
		t.Context(),
		`UPDATE job_runs SET status = 'succeeded', finished_at = now() WHERE id = $1`,
		id,
	); err != nil {
		t.Fatal(err)
	}

	return res
}

func inTx(t *testing.T, db *database.DB, f func(tx *database.Tx)) {
	t.Helper()

	err := db.Tx(t.Context(), func(tx *database.Tx) error {
		f(tx)

		return errRollback
	})
	if !errors.Is(err, errRollback) {
		t.Fatal(err)
	}
}

func txCount(t *testing.T, tx *database.Tx, q string, args ...any) int {
	t.Helper()

	var n int
	if err := tx.QueryRowContext(t.Context(), q, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

const systemSuggestions = `SELECT count(*) FROM suggestions WHERE author_id IS NULL AND meta->>'source' = 'dq.frontend_fields'`

// A movie with every required field but a director yields one coverage miss and one system suggestion; a rerun the
// same day adds none.
func TestFrontendFields(t *testing.T) {
	db := dqDB(t)

	inTx(t, db, func(tx *database.Tx) {
		if _, err := tx.ExecContext(t.Context(), completeMovie); err != nil {
			t.Fatal(err)
		}

		first := frontendRun(t, tx, db)
		if first.Seen != 1 || first.Changed != 1 {
			t.Fatalf("first run = %+v, want 1 seen, 1 suggestion", first)
		}

		if n := txCount(t, tx, `SELECT count(*) FROM field_coverage WHERE type = 'movie' AND missing > 0`); n != 1 {
			t.Fatalf("fields with a miss = %d, want 1", n)
		}

		if n := txCount(t, tx, `SELECT count(*) FROM field_coverage WHERE type = 'movie' AND field = 'director'
AND sampled = 1 AND missing = 1`); n != 1 {
			t.Fatal("director miss not recorded")
		}

		if n := txCount(
			t,
			tx,
			`SELECT count(*) FROM field_coverage WHERE type = 'movie' AND field = 'synopsis' AND locale = 'en'`,
		); n != 1 {
			t.Fatal("synopsis is not measured for en")
		}

		if n := txCount(
			t,
			tx,
			systemSuggestions+` AND jsonb_exists(changes, 'director') AND status = 'submitted'`,
		); n != 1 {
			t.Fatalf("suggestions = %d, want 1", n)
		}

		if again := frontendRun(t, tx, db); again.Changed != 0 {
			t.Fatalf("second run opened %d suggestions", again.Changed)
		}

		if n := txCount(t, tx, systemSuggestions); n != 1 {
			t.Fatalf("suggestions after rerun = %d, want 1", n)
		}
	})
}

// At most 200 system suggestions a day, across runs.
func TestFrontendFieldsCap(t *testing.T) {
	db := dqDB(t)

	inTx(t, db, func(tx *database.Tx) {
		if _, err := tx.ExecContext(t.Context(), completeMovie); err != nil {
			t.Fatal(err)
		}

		// 250 bare movies, each missing six required fields.
		if _, err := tx.ExecContext(t.Context(), `INSERT INTO items (type, id, slug, wikidata_qid, original_title)
SELECT 'movie', gen_random_uuid(), 'bare-' || g, 'Q' || (20000 + g), 'Bare ' || g FROM generate_series(1, 250) g`); err != nil {
			t.Fatal(err)
		}

		if got := frontendRun(t, tx, db); got.Changed != dq.DailySuggestionCap {
			t.Fatalf("opened %d, want the cap %d", got.Changed, dq.DailySuggestionCap)
		}

		if got := frontendRun(t, tx, db); got.Changed != 0 {
			t.Fatalf("second run opened %d past the cap", got.Changed)
		}

		if n := txCount(t, tx, systemSuggestions); n != dq.DailySuggestionCap {
			t.Fatalf("suggestions = %d, want %d", n, dq.DailySuggestionCap)
		}
	})
}
