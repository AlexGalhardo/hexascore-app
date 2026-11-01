//go:build integration

package ingest_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/rollback"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

const (
	itemOwned   = "0192f1c6-0000-7000-8000-0000000000b1"
	itemCreated = "0192f1c6-0000-7000-8000-0000000000b2"
	itemOld     = "0192f1c6-0000-7000-8000-0000000000b3"
)

// seedBadRun builds a finished fetch.wikidata run (id returned) that: changed Q1 (history kept), inserted Q2 and Q3
// (with their items), only touched Q4, and changed Q5 which a later run changed again.
func seedBadRun(t *testing.T, db *database.DB) int64 {
	t.Helper()

	ctx := t.Context()

	var id int64
	if err := db.QueryRowContext(ctx, `INSERT INTO job_runs (job, source, mode, status, triggered_by, started_at, finished_at, created_at)
VALUES ('fetch.wikidata', 'wikidata', 'full', 'succeeded', 'schedule', now() - interval '1 hour', now() - interval '10 minutes',
  now() - interval '1 hour') RETURNING id`).
		Scan(&id); err != nil {
		t.Fatal(err)
	}

	var later int64
	if err := db.QueryRowContext(ctx, `INSERT INTO job_runs (job, source, mode, status, triggered_by, started_at, finished_at)
VALUES ('fetch.wikidata', 'wikidata', 'full', 'succeeded', 'schedule', now() - interval '5 minutes', now())
RETURNING id`).
		Scan(&later); err != nil {
		t.Fatal(err)
	}

	stmts := []string{
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata','Wikidata','CC0-1.0','https://www.wikidata.org')
  ON CONFLICT DO NOTHING`,
		// Q1: existed, changed by the bad run.
		`INSERT INTO source_records (source, external_id, payload, hash, fetched_at, last_run_id, normalized_at)
VALUES ('wikidata','Q1','{"v":"new"}', repeat('b',64), now() - interval '30 minutes', $1, now())`,
		`INSERT INTO source_record_history (source, external_id, hash, payload, fetched_at, superseded_at, superseded_by_run)
VALUES ('wikidata','Q1', repeat('a',64), '{"v":"old"}', now() - interval '2 days', now() - interval '30 minutes', $1)`,
		// Q2 and Q3: inserted by the bad run, each with an item the run created.
		`INSERT INTO source_records (source, external_id, payload, hash, fetched_at, last_run_id, normalized_at)
VALUES ('wikidata','Q2','{}', repeat('c',64), now() - interval '30 minutes', $1, now()),
       ('wikidata','Q3','{}', repeat('d',64), now() - interval '30 minutes', $1, now())`,
		`INSERT INTO items (id, type, slug, wikidata_qid, original_title, created_at) VALUES
('` + itemOwned + `', 'movie', 'no-user-data', 'Q2', 'No user data', now() - interval '20 minutes'),
('` + itemCreated + `', 'movie', 'has-rating', 'Q3', 'Has rating', now() - interval '20 minutes')`,
		// Q4: older record the run only touched; Q5: changed by the bad run, then again by the later run.
		`INSERT INTO source_records (source, external_id, payload, hash, fetched_at, last_run_id, normalized_at)
VALUES ('wikidata','Q4','{"v":4}', repeat('e',64), now() - interval '5 days', $1, now())`,
		`INSERT INTO source_records (source, external_id, payload, hash, fetched_at, last_run_id, normalized_at)
VALUES ('wikidata','Q5','{"v":"third"}', repeat('f',64), now() - interval '2 minutes', $2, now())`,
		`INSERT INTO source_record_history (source, external_id, hash, payload, fetched_at, superseded_at, superseded_by_run)
VALUES ('wikidata','Q5', repeat('1',64), '{"v":"first"}', now() - interval '2 days', now() - interval '30 minutes', $1),
       ('wikidata','Q5', repeat('2',64), '{"v":"second"}', now() - interval '30 minutes', now() - interval '2 minutes', $2)`,
		// An older item whose record is not part of the bad run stays.
		`INSERT INTO items (id, type, slug, wikidata_qid, original_title, created_at)
VALUES ('` + itemOld + `', 'movie', 'old', 'Q1', 'Old', now() - interval '10 days')`,
	}

	rep := strings.NewReplacer("$1", itoa(int(id)), "$2", itoa(int(later)))

	for _, s := range stmts {
		if _, err := db.ExecContext(ctx, rep.Replace(s)); err != nil {
			t.Fatalf("%v: %s", err, s)
		}
	}

	var uid int64
	if err := db.QueryRowContext(ctx, `INSERT INTO users (name, email, password, username)
VALUES ('r', 'rb@example.test', 'x', 'rollbacker') RETURNING id`).Scan(&uid); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6)
VALUES ($1, $2, 50, 50, 50, 50, 50, 50)`, uid, itemCreated); err != nil {
		t.Fatal(err)
	}

	return id
}

func runRollback(t *testing.T, db *database.DB, target int64) run.Result {
	t.Helper()

	ctx := t.Context()

	rec := &recorder{}

	r, err := ingest.StartRollback(ctx, db, rec, target, 0, "bad payloads")
	if err != nil {
		t.Fatal(err)
	}

	var res run.Result

	err = db.Tx(ctx, func(tx *database.Tx) error {
		var serr error

		res, serr = rollback.Step(ctx, tx, r)

		return serr
	})
	if err != nil {
		t.Fatal(err)
	}

	return res
}

func TestRollbackRestoresPreviousState(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()
	target := seedBadRun(t, db)

	res := runRollback(t, db, target)
	if !res.Done {
		t.Error("rollback is a single step")
	}

	var payload string
	if err := db.QueryRowContext(ctx, `SELECT payload->>'v' FROM source_records WHERE external_id = 'Q1'`).
		Scan(&payload); err != nil ||
		payload != "old" {
		t.Errorf("Q1 payload = %q, %v; want old", payload, err)
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM source_records WHERE external_id = 'Q1' AND normalized_at IS NULL`,
	); n != 1 {
		t.Error("a restored record must be pending for the normaliser")
	}

	if n := count(t, db, `SELECT count(*) FROM source_record_history WHERE superseded_by_run = $1`, target); n != 1 {
		// only the Q5 row stays: a later run changed it, so it is not restored
		t.Errorf("history rows of the run = %d, want 1 (Q5)", n)
	}

	if n := count(t, db, `SELECT count(*) FROM source_records WHERE external_id IN ('Q2','Q3')`); n != 0 {
		t.Errorf("inserted records left = %d", n)
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM source_records WHERE external_id = 'Q4' AND normalized_at IS NOT NULL`,
	); n != 1 {
		t.Error("a record the run only touched must stay as it was")
	}

	if err := db.QueryRowContext(ctx, `SELECT payload->>'v' FROM source_records WHERE external_id = 'Q5'`).
		Scan(&payload); err != nil ||
		payload != "third" {
		t.Errorf("Q5 payload = %q, %v; want third (a later run owns it)", payload, err)
	}

	if res.Skipped != 1 {
		t.Errorf("skipped = %d, want 1 (Q5)", res.Skipped)
	}

	if n := count(t, db, `SELECT count(*) FROM items WHERE id = $1`, itemOwned); n != 0 {
		t.Error("an item the run created, without user data, must be deleted")
	}

	if n := count(t, db, `SELECT count(*) FROM items WHERE id = $1`, itemOld); n != 1 {
		t.Error("an item older than the run must stay")
	}

	if n := count(t, db, `SELECT count(*) FROM search_dirty WHERE ref_id = $1`, itemOwned); n != 1 {
		t.Error("a deleted item must be queued for the search resync")
	}
}

func TestRollbackKeepsItemsWithUserData(t *testing.T) {
	db := newDB(t)
	target := seedBadRun(t, db)

	runRollback(t, db, target)

	var vis string
	if err := db.QueryRowContext(t.Context(), `SELECT visibility FROM items WHERE id = $1`, itemCreated).
		Scan(&vis); err != nil {
		t.Fatalf("item with a rating must survive: %v", err)
	}

	if vis != "hidden" {
		t.Errorf("visibility = %q, want hidden (flagged)", vis)
	}
}

func TestRollbackIsLoggedAsRunAndAudit(t *testing.T) {
	db := newDB(t)
	target := seedBadRun(t, db)

	runRollback(t, db, target)

	if n := count(t, db, `SELECT count(*) FROM job_runs WHERE job = 'rollback' AND mode = 'rollback'
AND params->>'target_run' = $1::text AND params->>'reason' = 'bad payloads'`, itoa(int(target))); n != 1 {
		t.Errorf("rollback runs = %d, want 1", n)
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM catalog_audit WHERE field = '*ingest-rollback*' AND target_id = ANY($1::uuid[])
AND reason = 'bad payloads'`,
		"{"+itemOwned+","+itemCreated+"}",
	); n != 2 {
		t.Errorf("audit rows = %d, want 2 (deleted and hidden items)", n)
	}

	// A run is rolled back once.
	if _, err := ingest.StartRollback(
		t.Context(),
		db,
		&recorder{},
		target,
		0,
		"again",
	); !errors.Is(
		err,
		rollback.ErrAlreadyRolledBack,
	) {
		t.Errorf("second rollback err = %v, want ErrAlreadyRolledBack", err)
	}
}

func TestRollbackChainRenormalises(t *testing.T) {
	db := newDB(t)
	target := seedBadRun(t, db)
	rec := &recorder{}

	r, err := ingest.StartRollback(t.Context(), db, rec, target, 7, "x")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(
		t.Context(),
		`UPDATE job_runs SET status = 'succeeded' WHERE id = $1`,
		r.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := ingest.Advance(t.Context(), db, rec, r.ID); err != nil {
		t.Fatal(err)
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM job_runs WHERE job = 'normalize' AND triggered_by = $1`,
		"chain:"+itoa(int(r.ID)),
	); n != 1 {
		t.Errorf("normalize successor = %d, want 1", n)
	}

	if n := count(t, db, `SELECT count(*) FROM job_runs WHERE id = $1 AND triggered_by = 'admin:7'`, r.ID); n != 1 {
		t.Error("the rollback run must record the admin")
	}
}

func TestRollbackHistoryExpired(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()

	var id int64
	if err := db.QueryRowContext(ctx, `INSERT INTO job_runs (job, source, mode, status, triggered_by, started_at, finished_at, created_at)
VALUES ('fetch.wikidata', 'wikidata', 'full', 'succeeded', 'schedule', now() - interval '31 days', now() - interval '31 days',
  now() - interval '31 days') RETURNING id`).
		Scan(&id); err != nil {
		t.Fatal(err)
	}

	if _, err := ingest.StartRollback(
		ctx,
		db,
		&recorder{},
		id,
		0,
		"late",
	); !errors.Is(
		err,
		rollback.ErrHistoryExpired,
	) {
		t.Errorf("err = %v, want ErrHistoryExpired", err)
	}

	var norm int64
	if err := db.QueryRowContext(ctx, `INSERT INTO job_runs (job, mode, status, triggered_by, started_at)
VALUES ('normalize', 'full', 'succeeded', 'schedule', now()) RETURNING id`).Scan(&norm); err != nil {
		t.Fatal(err)
	}

	if _, err := ingest.StartRollback(
		ctx,
		db,
		&recorder{},
		norm,
		0,
		"x",
	); !errors.Is(
		err,
		rollback.ErrNotRollbackable,
	) {
		t.Errorf("normalize err = %v, want ErrNotRollbackable", err)
	}

	if _, err := ingest.StartRollback(ctx, db, &recorder{}, 999999, 0, "x"); !errors.Is(err, rollback.ErrRunNotFound) {
		t.Errorf("unknown run err = %v, want ErrRunNotFound", err)
	}
}
