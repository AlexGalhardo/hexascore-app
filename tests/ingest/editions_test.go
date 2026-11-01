//go:build integration

package ingest_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/backfill"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/editions"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
)

const (
	workA = "0192f1c6-0000-7000-8000-0000000000b1"
	workB = "0192f1c6-0000-7000-8000-0000000000b2"
)

func TestEditionsSaveIsIdempotentAndNeverMovesAnIsbn(t *testing.T) {
	db := newDB(t)

	if err := normalize.SeedDataSources(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	for id, q := range map[string]string{workA: "Q1", workB: "Q2"} {
		if _, err := db.ExecContext(
			t.Context(),
			`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'book', 'book-'||lower($2::text), $2)`,
			id,
			q,
		); err != nil {
			t.Fatal(err)
		}
	}

	s := editions.PG{Ex: db}
	c := editions.Candidate{ISBN: "9780306406157", Title: "A Book", Language: "pt", Year: 1990, Source: "wikidata"}

	for range 2 {
		if out, err := editions.Save(t.Context(), s, workA, c); err != nil || out != editions.Saved {
			t.Fatalf("save: %v, %v", out, err)
		}
	}

	if n := count(t, db, `SELECT count(*) FROM editions WHERE work_id = $1`, workA); n != 1 {
		t.Errorf("editions of A = %d, want 1", n)
	}

	if out, err := editions.Save(t.Context(), s, workB, c); err != nil || out != editions.Conflict {
		t.Fatalf("other work: %v, %v, want Conflict", out, err)
	}

	if n := count(t, db, `SELECT count(*) FROM editions WHERE work_id = $1`, workB); n != 0 {
		t.Errorf("editions of B = %d, want 0", n)
	}

	if err := s.SetPrimary(t.Context(), workA, c); err != nil {
		t.Fatal(err)
	}

	if n := count(t, db, `SELECT count(*) FROM items i JOIN editions e ON e.id = i.primary_edition_id
WHERE i.id = $1 AND e.isbn13 = '9780306406157'`, workA); n != 1 {
		t.Error("primary edition not set")
	}
}

const (
	isbnOne = "9780451524935"
	isbnTwo = "9780141187761"
)

func bookRecord(t *testing.T, db *database.DB, qid string, isbns []string, editionQids ...string) {
	t.Helper()

	claim := func(vals ...string) []map[string]any {
		out := make([]map[string]any, len(vals))
		for i, v := range vals {
			out[i] = map[string]any{"value": v, "rank": "normal"}
		}

		return out
	}

	payload, err := json.Marshal(map[string]any{
		"type": "book", "id": qid, "labels": map[string]string{"en": "Book " + qid},
		"claims": map[string]any{"P212": claim(isbns...), "P747": claim(editionQids...)},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), `INSERT INTO source_records (source, external_id, payload, hash)
VALUES ('wikidata', $1::text, $2::jsonb, encode(sha256(($1::text)::bytea), 'hex'))`, qid, string(payload)); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeWritesEditionsAndRerunIsIdempotent(t *testing.T) {
	db := newDB(t)
	bookRecord(t, db, "Q500", []string{isbnOne}, "Q501")
	drain(t, db, (&normSpy{}).job(0))

	if n := count(
		t,
		db,
		`SELECT count(*) FROM editions e JOIN items i ON i.id = e.work_id WHERE i.wikidata_qid = 'Q500'`,
	); n != 2 {
		t.Fatalf("editions = %d, want 2 (one ISBN, one QID)", n)
	}

	repend(t, db, "1 hour")
	drain(t, db, (&normSpy{}).job(0))

	if n := count(t, db, `SELECT count(*) FROM editions`); n != 2 {
		t.Errorf("editions after a rerun = %d, want 2", n)
	}
}

func TestNormalizeIsbnConflictIsAFailureAndAReview(t *testing.T) {
	db := newDB(t)

	if err := normalize.SeedDataSources(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(
		t.Context(),
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'book', 'owner', 'Q1')`,
		workA,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := editions.Save(
		t.Context(),
		editions.PG{Ex: db},
		workA,
		editions.Candidate{ISBN: isbnOne, Title: "Owner", Language: "pt", Source: "wikidata"},
	); err != nil {
		t.Fatal(err)
	}

	bookRecord(t, db, "Q500", []string{isbnOne, isbnTwo})

	err := db.Tx(t.Context(), func(tx *database.Tx) error {
		res, err := (&normSpy{}).job(0).Step(t.Context(), tx, run.Run{ID: 7})
		if err == nil && res.Failed != 1 {
			err = fmt.Errorf("failed = %d, want 1", res.Failed)
		}

		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	if n := count(t, db, `SELECT count(*) FROM items WHERE wikidata_qid = 'Q500'`); n != 1 {
		t.Errorf("the item was not saved (%d)", n)
	}

	if n := count(t, db, `SELECT count(*) FROM editions WHERE work_id = $1`, workA); n != 1 {
		t.Errorf("owner editions = %d, want 1 (never moved)", n)
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM editions e JOIN items i ON i.id = e.work_id WHERE i.wikidata_qid = 'Q500' AND e.isbn13 = $1`,
		isbnTwo,
	); n != 1 {
		t.Errorf("the other ISBN was not saved (%d)", n)
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM ingest_failures WHERE source = 'editions' AND external_id = $1 AND error_class = 'isbn_conflict'`,
		isbnOne,
	); n != 1 {
		t.Errorf("isbn_conflict failures = %d, want 1", n)
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM entity_match_reviews WHERE role = 'edition' AND raw_name = $1 AND $2::uuid = ANY(candidates)`,
		isbnOne,
		workA,
	); n != 1 {
		t.Errorf("edition reviews = %d, want 1", n)
	}

	if n := count(t, db, `SELECT count(*) FROM ingest_failures WHERE source = 'normalize'`); n != 0 {
		t.Errorf("an ISBN conflict must not park the item (%d normalize failures)", n)
	}
}

func TestBackfillEditionsWalksBooksByKeyset(t *testing.T) {
	db := newDB(t)

	if err := normalize.SeedDataSources(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	for i, isbn := range []string{isbnOne, isbnTwo, "9780306406157"} {
		qid := fmt.Sprintf("Q%d", 600+i)
		bookRecord(t, db, qid, []string{isbn})

		if _, err := db.ExecContext(
			t.Context(),
			`INSERT INTO items (id, type, slug, wikidata_qid) VALUES (gen_random_uuid(), 'book', $1, $2)`,
			"book-"+strings.ToLower(qid),
			qid,
		); err != nil {
			t.Fatal(err)
		}
	}

	step, cursor, pages := backfill.Editions{Batch: 2}.Step, json.RawMessage(nil), 0

	for {
		var res run.Result

		err := db.Tx(t.Context(), func(tx *database.Tx) error {
			var err error
			res, err = step(t.Context(), tx, run.Run{ID: 9, Cursor: cursor})

			return err
		})
		if err != nil {
			t.Fatal(err)
		}

		pages++
		cursor = res.Cursor

		if res.Done {
			break
		}

		if pages > 5 {
			t.Fatal("backfill did not finish")
		}
	}

	if pages != 2 {
		t.Errorf("pages = %d, want 2 (2 + 1 items)", pages)
	}

	if n := count(t, db, `SELECT count(*) FROM editions`); n != 3 {
		t.Errorf("editions = %d, want 3", n)
	}
}

func TestNormalizeQidConflictIsAFailureAndAReviewAndTheItemStillSaves(t *testing.T) {
	db := newDB(t)

	if err := normalize.SeedDataSources(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(
		t.Context(),
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'book', 'owner', 'Q1')`,
		workA,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := editions.Save(
		t.Context(),
		editions.PG{Ex: db},
		workA,
		editions.Candidate{QID: "Q777", Title: "Owner", Language: "pt", Source: "wikidata"},
	); err != nil {
		t.Fatal(err)
	}

	bookRecord(t, db, "Q500", []string{isbnTwo}, "Q777")

	err := db.Tx(t.Context(), func(tx *database.Tx) error {
		res, err := (&normSpy{}).job(0).Step(t.Context(), tx, run.Run{ID: 7})
		if err == nil && res.Failed != 1 {
			err = fmt.Errorf("failed = %d, want 1", res.Failed)
		}

		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	if n := count(t, db, `SELECT count(*) FROM items WHERE wikidata_qid = 'Q500'`); n != 1 {
		t.Errorf("the item was not saved (%d)", n)
	}

	if n := count(t, db, `SELECT count(*) FROM editions WHERE wikidata_qid = 'Q777' AND work_id = $1`, workA); n != 1 {
		t.Errorf("the QID moved away from its owner (%d)", n)
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM editions e JOIN items i ON i.id = e.work_id WHERE i.wikidata_qid = 'Q500' AND e.isbn13 = $1`,
		isbnTwo,
	); n != 1 {
		t.Errorf("the other edition was not saved (%d)", n)
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM ingest_failures WHERE source = 'editions' AND external_id = 'Q777' AND error_class = 'qid_conflict'`,
	); n != 1 {
		t.Errorf("qid_conflict failures = %d, want 1", n)
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM entity_match_reviews WHERE role = 'edition' AND raw_name = 'Q777' AND $1::uuid = ANY(candidates)`,
		workA,
	); n != 1 {
		t.Errorf("edition reviews = %d, want 1", n)
	}

	if n := count(t, db, `SELECT count(*) FROM ingest_failures WHERE source = 'normalize'`); n != 0 {
		t.Errorf("a QID conflict must not park the item (%d normalize failures)", n)
	}
}
