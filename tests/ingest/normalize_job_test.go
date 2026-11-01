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

func (noCommunity) CommunityFields(context.Context, string) (map[string]normalize.CommunityField, error) {
	return nil, nil
}

func (noCommunity) OpenSystemSuggestion(context.Context, normalize.SystemSuggestion) error {
	return nil
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
