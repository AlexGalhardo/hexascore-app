//go:build integration

package ingest_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

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

	if err := s.SetPrimary(t.Context(), workA, "pt"); err != nil {
		t.Fatal(err)
	}

	if n := count(t, db, `SELECT count(*) FROM items i JOIN editions e ON e.id = i.primary_edition_id
WHERE i.id = $1 AND e.isbn13 = '9780306406157'`, workA); n != 1 {
		t.Error("primary edition not set")
	}

	ol := editions.Candidate{OpenLibraryID: "OL1M", Title: "OL Book", Language: "en", Source: "openlibrary"}
	if out, err := editions.Save(t.Context(), s, workA, ol); err != nil || out != editions.Saved {
		t.Fatalf("openlibrary save: %v, %v", out, err)
	}

	if _, err := editions.Save(t.Context(), s, workB, ol); err != nil {
		t.Fatalf("taken openlibrary id on another work must not fail the save: %v", err)
	}

	if n := count(t, db, `SELECT count(*) FROM editions WHERE work_id = $1`, workB); n != 0 {
		t.Errorf("editions of B = %d, want 0", n)
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

// editionEntity stores the wikidata:entity record of a P747 edition.
func editionEntity(t *testing.T, db *database.DB, qid, lang, year string) {
	t.Helper()

	payload := fmt.Sprintf(`{"id":%q,"labels":{"en":"Edition %s"},"claims":{
"P407":[{"value":%q,"rank":"normal"}],
"P577":[{"value":{"time":"+%s-00-00T00:00:00Z","precision":9},"rank":"normal"}]}}`, qid, qid, lang, year)
	if _, err := db.ExecContext(t.Context(), `INSERT INTO source_records (source, external_id, payload, hash)
VALUES ('wikidata:entity', $1::text, $2::jsonb, encode(sha256(($1::text)::bytea), 'hex'))`, qid, payload); err != nil {
		t.Fatal(err)
	}
}

// ADR 0012: normalize and backfill.editions read the stored P747 edition records, so the editions carry language
// and year (P407, P577) instead of `und` and no year.
func TestEditionsTakeLanguageAndYearFromEditionRecords(t *testing.T) {
	db := newDB(t)
	bookRecord(t, db, "Q700", nil, "Q701")
	editionEntity(t, db, "Q701", "Q5146", "1999")
	drain(t, db, (&normSpy{}).job(0))

	var lang string
	var year int
	if err := db.QueryRowContext(t.Context(), `SELECT language, coalesce(publication_year, 0) FROM editions
WHERE wikidata_qid = 'Q701'`).Scan(&lang, &year); err != nil {
		t.Fatal(err)
	}
	if lang != "pt" || year != 1999 {
		t.Errorf("normalize: edition = %s %d, want pt 1999", lang, year)
	}

	if _, err := db.ExecContext(t.Context(), `DELETE FROM editions`); err != nil {
		t.Fatal(err)
	}

	err := db.Tx(t.Context(), func(tx *database.Tx) error {
		_, err := backfill.Editions{}.Step(t.Context(), tx, run.Run{ID: 9})

		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.QueryRowContext(t.Context(), `SELECT language, coalesce(publication_year, 0) FROM editions
WHERE wikidata_qid = 'Q701'`).Scan(&lang, &year); err != nil {
		t.Fatal(err)
	}
	if lang != "pt" || year != 1999 {
		t.Errorf("backfill: edition = %s %d, want pt 1999", lang, year)
	}
}

// ADR 0012: an edition's publisher (P123) resolves through the credit matcher (QID first), and the primary edition
// is the earliest one in the work's original language (P407) while no edition has a cover. Open Library fills the
// language and year Wikidata lacks; both normalize and backfill.editions do this.
func TestEditionsResolvePublisherPrimaryAndOpenLibrary(t *testing.T) {
	db := newDB(t)

	const org = "0192f1c6-0000-7000-8000-0000000000e1"
	if _, err := db.ExecContext(t.Context(), `INSERT INTO entities (id, kind, slug, wikidata_qid, default_name)
VALUES ($1, 'organization', 'penguin', 'Q55', 'Penguin')`, org); err != nil {
		t.Fatal(err)
	}

	payload := `{"type":"book","id":"Q800","labels":{"en":"Book Q800"},"claims":{
"P407":[{"value":"Q1860","rank":"normal"}],
"P212":[{"value":"9780306406157","rank":"normal"}],
"P747":[{"value":"Q801","rank":"normal"},{"value":"Q802","rank":"normal"}]}}`
	if _, err := db.ExecContext(t.Context(), `INSERT INTO source_records (source, external_id, payload, hash)
VALUES ('wikidata', 'Q800', $1::jsonb, encode(sha256('h800'), 'hex')),
('openlibrary', 'Q800', $2::jsonb, encode(sha256('o800'), 'hex'))`, payload,
		`{"qid":"Q800","key":"/books/OL9M","isbn":"0306406152","language":"eng","year":1970,"covers":[]}`); err != nil {
		t.Fatal(err)
	}

	editionEntity(t, db, "Q802", "Q1860", "1990")
	if _, err := db.ExecContext(t.Context(), `UPDATE source_records SET payload = jsonb_set(payload, '{claims,P123}',
'[{"value":"Q55","rank":"normal"}]') WHERE external_id = 'Q802'`); err != nil {
		t.Fatal(err)
	}

	check := func(stage string) {
		t.Helper()

		if n := count(
			t,
			db,
			`SELECT count(*) FROM editions WHERE wikidata_qid = 'Q802' AND publisher_id = $1`,
			org,
		); n != 1 {
			t.Errorf("%s: Q802 publisher not resolved to the Q55 entity", stage)
		}

		if n := count(t, db, `SELECT count(*) FROM editions WHERE isbn13 = '9780306406157' AND language = 'en'
AND publication_year = 1970 AND openlibrary_id = 'OL9M'`); n != 1 {
			t.Errorf("%s: the ISBN edition did not take the Open Library language, year and id", stage)
		}

		if n := count(t, db, `SELECT count(*) FROM items i JOIN editions e ON e.id = i.primary_edition_id
WHERE i.wikidata_qid = 'Q800' AND e.isbn13 = '9780306406157'`); n != 1 {
			t.Errorf("%s: primary edition is not the 1970 English edition", stage)
		}
	}

	drain(t, db, (&normSpy{}).job(0))
	check("normalize")

	if _, err := db.ExecContext(
		t.Context(),
		`UPDATE items SET primary_edition_id = NULL; DELETE FROM editions`,
	); err != nil {
		t.Fatal(err)
	}

	err := db.Tx(t.Context(), func(tx *database.Tx) error {
		_, err := backfill.Editions{}.Step(t.Context(), tx, run.Run{ID: 9})

		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	check("backfill")
}

// SDD 06 §7.8: a backfill pauses 100 ms between batches, through the runner's clock; other modes do not.
func TestBackfillPausesBetweenBatches(t *testing.T) {
	for mode, want := range map[string]int{run.ModeBackfill: 2, run.ModeFull: 0} {
		t.Run(mode, func(t *testing.T) {
			db := newDB(t)
			rec := &recorder{}

			var pauses []time.Duration

			r := newRunner(db, rec, newFlags(t, map[string]string{"FLAG_INGEST_ENABLED": "true"}), run.Limits{},
				map[string]run.Step{"b": upsertStep(3, 1, nil)})
			r.Sleep = func(_ context.Context, d time.Duration) error {
				pauses = append(pauses, d)

				return nil
			}

			cr, err := run.Create(t.Context(), db, rec, run.Spec{Job: "b", Mode: mode, TriggeredBy: "test"})
			if err != nil {
				t.Fatal(err)
			}

			if out, err := r.Execute(t.Context(), cr.ID); err != nil || out != run.Finished {
				t.Fatalf("Execute = %v, %v", out, err)
			}

			if len(pauses) != want || want > 0 && pauses[0] != run.BackfillPause {
				t.Errorf("pauses = %v, want %d of %v", pauses, want, run.BackfillPause)
			}
		})
	}
}
