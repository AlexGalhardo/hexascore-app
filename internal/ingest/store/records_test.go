//go:build integration

package store_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

func newDB(t *testing.T) *database.DB {
	t.Helper()

	env := map[string]string{"DATABASE_URL": pgtest.Database(t, "")}
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func count(t *testing.T, db *database.DB, q string, args ...any) int {
	t.Helper()

	var n int
	if err := db.QueryRowContext(context.Background(), q, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

func rec(id, payload string) store.Record {
	return store.Record{Source: "wikidata", ExternalID: id, Payload: json.RawMessage(payload)}
}

func TestUpsertIsNoopWhenPayloadUnchanged(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()

	if ok, err := store.Upsert(ctx, db, 1, rec("Q1", `{"a":1}`)); err != nil || !ok {
		t.Fatalf("first upsert = %v, %v; want inserted", ok, err)
	}

	if _, err := db.ExecContext(ctx, `UPDATE source_records SET fetched_at = now() - interval '1 day',
checked_at = now() - interval '1 day'`); err != nil {
		t.Fatal(err)
	}

	// Same payload, different key order and whitespace: still unchanged.
	if ok, err := store.Upsert(ctx, db, 2, rec("Q1", `{ "a": 1 }`)); err != nil || ok {
		t.Fatalf("second upsert = %v, %v; want no change", ok, err)
	}

	if n := count(t, db, `SELECT count(*) FROM source_record_history`); n != 0 {
		t.Errorf("history rows = %d, want 0", n)
	}

	if n := count(t, db, `SELECT count(*) FROM source_records WHERE last_run_id = 2
AND checked_at > now() - interval '1 hour' AND fetched_at < now() - interval '1 hour'`); n != 1 {
		t.Errorf("row not touched checked_at-only: %d", n)
	}
}

func TestStoreRecordsCountsInsertsAndChanges(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()

	seen, changed, err := store.Store(ctx, db, 1, []store.Record{rec("Q1", `{"v":1}`), rec("Q2", `{"v":1}`)})
	if err != nil || seen != 2 || changed != 2 {
		t.Fatalf("first = %d seen, %d changed, %v; want 2, 2", seen, changed, err)
	}

	second := []store.Record{rec("Q1", `{"v":1}`), rec("Q2", `{"v":2}`), rec("Q3", `{"v":1}`)}
	seen, changed, err = store.Store(ctx, db, 2, second)
	if err != nil || seen != 3 || changed != 2 {
		t.Fatalf("second = %d seen, %d changed, %v; want 3, 2 (Q2 changed, Q3 new)", seen, changed, err)
	}
}

func TestStoreSpansChunks(t *testing.T) {
	db := newDB(t)

	recs := make([]store.Record, store.ChunkSize+5)
	for i := range recs {
		recs[i] = rec(fmt.Sprintf("Q%d", i+1), `{"v":1}`)
	}

	seen, changed, err := store.Store(t.Context(), db, 1, recs)
	if err != nil || seen != len(recs) || changed != len(recs) {
		t.Fatalf("= %d, %d, %v; want %d", seen, changed, err, len(recs))
	}
}

func TestHistoryWrittenOnHashChange(t *testing.T) {
	db := newDB(t)
	ctx := t.Context()

	for run, payload := range []string{`{"v":1}`, `{"v":2}`, `{"v":3}`} {
		if _, err := store.Upsert(ctx, db, int64(run+1), rec("Q1", payload)); err != nil {
			t.Fatal(err)
		}
	}

	if n := count(t, db, `SELECT count(*) FROM source_record_history WHERE payload->>'v' IN ('1','2')
AND superseded_by_run = (CASE payload->>'v' WHEN '1' THEN 2 ELSE 3 END)`); n != 2 {
		t.Errorf("history rows for v1 and v2 = %d, want 2", n)
	}

	if n := count(t, db, `SELECT count(*) FROM source_records WHERE payload->>'v' = '3'`); n != 1 {
		t.Errorf("current row not v3")
	}
}

func TestHistoryTwoChangesInOneTransaction(t *testing.T) {
	db := newDB(t)

	_, changed, err := store.Store(t.Context(), db, 1,
		[]store.Record{rec("Q1", `{"v":1}`), rec("Q1", `{"v":2}`), rec("Q1", `{"v":3}`)})
	if err != nil {
		t.Fatal(err)
	}

	if changed != 3 {
		t.Errorf("changed = %d, want 3", changed)
	}
}
