//go:build integration

package ingest_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

var normNow = time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)

type normSpy struct {
	changed [][]string
	scored  []string
}

func (s *normSpy) job(batch int) normalize.Normalizer {
	return normalize.Normalizer{
		Batch: batch,
		Now:   func() time.Time { return normNow },
		ItemsChanged: func(_ context.Context, _ database.Executor, ids []string) error {
			s.changed = append(s.changed, ids)

			return nil
		},
		ScoreChanged: func(_ context.Context, _ database.Executor, id string) error {
			s.scored = append(s.scored, id)

			return nil
		},
		Community: noCommunity{},
	}
}

type noCommunity struct{}

func (noCommunity) CommunityFields(
	context.Context,
	*database.Tx,
	string,
) (map[string]normalize.CommunityField, error) {
	return nil, nil
}

func (noCommunity) OpenSystemSuggestion(context.Context, *database.Tx, normalize.SystemSuggestion) error {
	return nil
}

// lockedCommunity reports fixed community fields; hook runs on every lookup (a concurrent writer, in tests).
type lockedCommunity struct {
	noCommunity
	fields map[string]normalize.CommunityField
	hook   func()
}

func (l lockedCommunity) CommunityFields(
	context.Context,
	*database.Tx,
	string,
) (map[string]normalize.CommunityField, error) {
	if l.hook != nil {
		l.hook()
	}

	return l.fields, nil
}

func repend(t *testing.T, db *database.DB, extra string) {
	t.Helper()

	if _, err := db.ExecContext(
		t.Context(),
		`UPDATE source_records SET fetched_at = now() + interval '`+extra+`'`,
	); err != nil {
		t.Fatal(err)
	}
}

func loadWikidata(t *testing.T, db *database.DB) {
	t.Helper()

	raw, err := os.ReadFile("../../internal/ingest/testdata/normalize_wikidata_payloads.json")
	if err != nil {
		t.Fatal(err)
	}

	var all map[string]json.RawMessage
	if err := json.Unmarshal(raw, &all); err != nil {
		t.Fatal(err)
	}

	for qid, p := range all {
		if _, err := db.ExecContext(t.Context(), `INSERT INTO source_records (source, external_id, payload, hash)
VALUES ('wikidata', $1::text, $2::jsonb, encode(sha256(($1::text)::bytea), 'hex'))`, qid, string(p)); err != nil {
			t.Fatal(err)
		}
	}
}

// drain runs Step until it reports Done, one transaction per call.
func drain(t *testing.T, db *database.DB, j normalize.Normalizer) {
	t.Helper()

	for range 20 {
		var res run.Result

		err := db.Tx(t.Context(), func(tx *database.Tx) error {
			var err error
			res, err = j.Step(t.Context(), tx, run.Run{})

			return err
		})
		if err != nil {
			t.Fatal(err)
		}

		if res.Done {
			return
		}
	}

	t.Fatal("normalise did not finish")
}

func TestNormalizeWritesItemTranslations(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	drain(t, db, (&normSpy{}).job(0))

	if n := count(t, db, `SELECT count(*) FROM items`); n != 4 {
		t.Errorf("items = %d, want 4", n)
	}

	if n := count(t, db, `SELECT count(*) FROM data_sources`); n < len(normalize.DataSources) {
		t.Errorf("data_sources = %d, want the seed", n)
	}

	if n := count(t, db, `SELECT count(*) FROM item_translations t JOIN items i ON i.id = t.item_id
WHERE i.wikidata_qid = 'Q83495' AND t.locale = 'en' AND t.title = 'The Matrix' AND t.title_source = 'wikidata'`); n != 1 {
		t.Error("Matrix en translation missing")
	}

	if n := count(t, db, `SELECT count(*) FROM items WHERE provenance ->> 'translations.en.title' = 'wikidata'
AND cover_source IS NULL`); n != 4 {
		t.Errorf("items with provenance and no cover source = %d, want 4", n)
	}

	if n := count(t, db, `SELECT count(*) FROM source_records WHERE normalized_at IS NULL OR item_id IS NULL`); n != 0 {
		t.Errorf("records still pending or unlinked = %d", n)
	}
}

func TestSlugCollisionAppendsYear(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)

	if _, err := db.ExecContext(t.Context(), `INSERT INTO items (id, type, slug, wikidata_qid)
VALUES ('00000000-0000-0000-0000-0000000000bb', 'movie', 'the-matrix', 'Q1')`); err != nil {
		t.Fatal(err)
	}

	drain(t, db, (&normSpy{}).job(0))

	if n := count(
		t,
		db,
		`SELECT count(*) FROM items WHERE wikidata_qid = 'Q83495' AND slug = 'the-matrix-1999'`,
	); n != 1 {
		t.Error("collision did not append the year")
	}

	// A stored slug never changes, even after the collision is gone.
	if _, err := db.ExecContext(t.Context(), `DELETE FROM items WHERE wikidata_qid = 'Q1';
UPDATE source_records SET fetched_at = now() + interval '1 minute'`); err != nil {
		t.Fatal(err)
	}

	drain(t, db, (&normSpy{}).job(0))

	if n := count(
		t,
		db,
		`SELECT count(*) FROM items WHERE wikidata_qid = 'Q83495' AND slug = 'the-matrix-1999'`,
	); n != 1 {
		t.Error("slug changed on re-normalise")
	}
}

func TestItemsChangedPerBatch(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)

	spy := &normSpy{}
	drain(t, db, spy.job(3))

	if len(spy.changed) != 2 || len(spy.changed[0]) != 3 || len(spy.changed[1]) != 1 {
		t.Errorf("ItemsChanged batches = %v, want sizes 3 and 1", spy.changed)
	}
}

func TestRecomputeDispatchedOnDateChange(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	drain(t, db, (&normSpy{}).job(0))

	var id string
	if err := db.QueryRowContext(t.Context(), `SELECT id::text FROM items WHERE wikidata_qid = 'Q83495'`).
		Scan(&id); err != nil {
		t.Fatal(err)
	}

	var uid int64
	if err := db.QueryRowContext(t.Context(), `INSERT INTO users (name, email, password, username)
VALUES ('r', 'r@example.test', 'x', 'rater') RETURNING id`).Scan(&uid); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), `INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6)
VALUES ($1, $2, 5, 5, 5, 5, 5, 5)`, uid, id); err != nil {
		t.Fatal(err)
	}

	// Unchanged data: re-normalising dispatches nothing.
	if _, err := db.ExecContext(
		t.Context(),
		`UPDATE source_records SET fetched_at = now() + interval '1 minute'`,
	); err != nil {
		t.Fatal(err)
	}

	spy := &normSpy{}
	drain(t, db, spy.job(0))

	if len(spy.scored) != 0 {
		t.Fatalf("recompute dispatched without a change: %v", spy.scored)
	}

	if _, err := db.ExecContext(
		t.Context(),
		`UPDATE items SET release_precision = 'year' WHERE id = $1`,
		id,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(
		t.Context(),
		`UPDATE source_records SET fetched_at = now() + interval '2 minutes'`,
	); err != nil {
		t.Fatal(err)
	}

	drain(t, db, spy.job(0))

	if len(spy.scored) != 1 || spy.scored[0] != id {
		t.Errorf("recompute dispatched for %v, want only the rated item %s", spy.scored, id)
	}
}

// FR-06-4: a community-locked field the source does not carry keeps its stored value.
func TestLockedFieldsSurviveASourceWithoutThem(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	drain(t, db, (&normSpy{}).job(0))

	for _, q := range []string{
		`UPDATE items SET release_date = '2000-01-01', external_ids = external_ids || '{"custom":"c-1"}'
 WHERE wikidata_qid = 'Q83495'`,
		`UPDATE item_translations SET synopsis = 'Community synopsis', synopsis_source = 'wikidata', synopsis_licence = 'proprietary'
 WHERE locale = 'en' AND item_id = (SELECT id FROM items WHERE wikidata_qid = 'Q83495')`,
		`UPDATE source_records SET payload = payload #- '{claims,P577}' WHERE external_id = 'Q83495'`,
	} {
		if _, err := db.ExecContext(t.Context(), q); err != nil {
			t.Fatal(err)
		}
	}

	repend(t, db, "1 minute")

	long := normNow.AddDate(-1, 0, 0)

	for _, days := range []int{0, 400} {
		j := (&normSpy{}).job(0)
		j.Now = func() time.Time { return normNow.AddDate(0, 0, days) }
		j.Community = lockedCommunity{fields: map[string]normalize.CommunityField{
			"release_date":        {Since: long, Current: json.RawMessage(`{"value":"2000-01-01","precision":"day"}`)},
			"external_ids.custom": {Since: long, Current: json.RawMessage(`"c-1"`)},
			"synopsis.en":         {Since: long, Current: json.RawMessage(`"Community synopsis"`)},
		}}
		drain(t, db, j)

		if n := count(t, db, `SELECT count(*) FROM items WHERE wikidata_qid = 'Q83495'
AND release_date = '2000-01-01' AND external_ids ->> 'custom' = 'c-1'`); n != 1 {
			t.Errorf("day %d: release_date or external_ids were wiped", days)
		}

		if n := count(t, db, `SELECT count(*) FROM item_translations t JOIN items i ON i.id = t.item_id
WHERE i.wikidata_qid = 'Q83495' AND t.locale = 'en' AND t.synopsis = 'Community synopsis'`); n != 1 {
			t.Errorf("day %d: synopsis was wiped", days)
		}

		repend(t, db, "2 minutes")
	}
}

// One record the database rejects is recorded and skipped; the rest of the chunk commits.
func TestBadRecordDoesNotStallTheChunk(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)

	if _, err := db.ExecContext(t.Context(), `INSERT INTO source_records (source, external_id, payload, hash)
SELECT source, 'Q9999999', jsonb_set(payload, '{type}', '"podcast"'), repeat('a', 64)
FROM source_records WHERE external_id = 'Q83495'`); err != nil {
		t.Fatal(err)
	}

	var res run.Result

	err := db.Tx(t.Context(), func(tx *database.Tx) error {
		var err error
		res, err = (&normSpy{}).job(0).Step(t.Context(), tx, run.Run{})

		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	if res.Failed != 1 || res.Changed != 4 {
		t.Errorf("failed = %d, changed = %d, want 1 and 4", res.Failed, res.Changed)
	}

	if n := count(t, db, `SELECT count(*) FROM items`); n != 4 {
		t.Errorf("items = %d, want 4", n)
	}

	if n := count(t, db, `SELECT count(*) FROM ingest_failures WHERE external_id = 'Q9999999'`); n != 1 {
		t.Errorf("ingest_failures = %d, want 1", n)
	}

	if n := count(t, db, `SELECT count(*) FROM source_records WHERE normalized_at IS NULL`); n != 0 {
		t.Errorf("records still pending = %d, the bad one must not block the index", n)
	}
}

func TestCancelledStatusIsKept(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	drain(t, db, (&normSpy{}).job(0))

	if _, err := db.ExecContext(
		t.Context(),
		`UPDATE items SET release_status = 'cancelled' WHERE wikidata_qid = 'Q83495'`,
	); err != nil {
		t.Fatal(err)
	}

	repend(t, db, "1 minute")
	drain(t, db, (&normSpy{}).job(0))

	if n := count(
		t,
		db,
		`SELECT count(*) FROM items WHERE wikidata_qid = 'Q83495' AND release_status = 'cancelled'`,
	); n != 1 {
		t.Error("a cancelled item was revived")
	}
}

// A record fetched while the batch runs stays pending: markDone retires only the version it read.
func TestRecordFetchedMidBatchStaysPending(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	drain(t, db, (&normSpy{}).job(0))

	if _, err := db.ExecContext(t.Context(), `UPDATE source_records SET normalized_at = NULL`); err != nil {
		t.Fatal(err)
	}

	j := (&normSpy{}).job(0)
	j.Community = lockedCommunity{hook: func() {
		if _, err := db.ExecContext(
			context.Background(),
			`UPDATE source_records SET fetched_at = fetched_at + interval '1 hour' WHERE external_id = 'Q83495'`,
		); err != nil {
			t.Error(err)
		}
	}}
	drain(t, db, j)

	if n := count(t, db, `SELECT count(*) FROM source_records WHERE normalized_at IS NULL`); n != 1 {
		t.Errorf("pending after the batch = %d, want only the record fetched mid-batch", n)
	}
}

func TestNormalizeWritesAPIChanges(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	drain(t, db, (&normSpy{}).job(0))

	if n := count(t, db, `SELECT count(*) FROM api_changes WHERE action = 'created'`); n != 4 {
		t.Errorf("created api_changes rows = %d, want 4", n)
	}

	// Re-normalising unchanged records adds nothing.
	repend(t, db, "1 minute")
	drain(t, db, (&normSpy{}).job(0))

	if n := count(t, db, `SELECT count(*) FROM api_changes`); n != 4 {
		t.Errorf("api_changes rows after an unchanged re-run = %d, want 4", n)
	}

	var id string
	if err := db.QueryRowContext(t.Context(), `SELECT id::text FROM items WHERE wikidata_qid = 'Q83495'`).
		Scan(&id); err != nil {
		t.Fatal(err)
	}

	// A stored change to a public field is rewritten by the next run and lands in the feed.
	if _, err := db.ExecContext(t.Context(), `UPDATE items SET runtime_minutes = 1 WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}

	repend(t, db, "2 minutes")
	drain(t, db, (&normSpy{}).job(0))

	var fields string
	if err := db.QueryRowContext(t.Context(), `SELECT array_to_string(fields, ',') FROM api_changes
WHERE item_id = $1 AND action = 'updated'`, id).Scan(&fields); err != nil || fields != "runtimeMinutes" {
		t.Errorf("updated fields = %q, %v; want runtimeMinutes", fields, err)
	}
}
