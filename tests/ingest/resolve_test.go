//go:build integration

package ingest_test

import (
	"testing"

	"github.com/AlexGalhardo/vukem/database"
)

const (
	otherItem = "00000000-0000-0000-0000-0000000000c1"
	matrixQID = "Q83495"
)

func mustExec(t *testing.T, db *database.DB, q string, args ...any) {
	t.Helper()

	if _, err := db.ExecContext(t.Context(), q, args...); err != nil {
		t.Fatalf("%v\n%s", err, q)
	}
}

// owner stores an item of another QID that already owns the given external ids (RFC-001 section 5 C).
func owner(t *testing.T, db *database.DB, extIDs string) {
	t.Helper()
	mustExec(
		t,
		db,
		`INSERT INTO items (id, type, slug, wikidata_qid, external_ids) VALUES ($1, 'movie', 'owner', 'Q1', $2::jsonb)`,
		otherItem,
		extIDs,
	)
}

// RFC-001 CW-06: one external id maps to at most one item. A new item whose id another item already owns is held for
// review, not created, and the rest of the batch goes on.
func TestResolveHoldsANewItemWhoseExternalIDIsTaken(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	owner(t, db, `{"imdb":"tt0133093"}`)

	drain(t, db, (&normSpy{}).job(0))

	if n := count(t, db, `SELECT count(*) FROM items WHERE wikidata_qid = '`+matrixQID+`'`); n != 0 {
		t.Fatalf("the contested item was created: %d", n)
	}

	if n := count(t, db, `SELECT count(*) FROM items WHERE wikidata_qid IN ('Q1079', 'Q4267401')`); n != 2 {
		t.Fatalf("the other items of the batch were not created: %d", n)
	}

	const open = `SELECT count(*) FROM item_conflicts WHERE wikidata_qid = '` + matrixQID + `' AND kind = 'external_id'
AND key = 'imdb' AND value = 'tt0133093' AND other_item_id = '` + otherItem + `' AND status = 'open'`
	if n := count(t, db, open); n != 1 {
		t.Fatalf("open conflicts = %d, want 1", n)
	}

	// running again neither creates the item nor a second review
	repend(t, db, "1 minute")
	drain(t, db, (&normSpy{}).job(0))

	if n := count(t, db, `SELECT count(*) FROM item_conflicts`); n != 1 {
		t.Fatalf("conflict rows after a rerun = %d, want 1", n)
	}

	if n := count(t, db, `SELECT count(*) FROM items WHERE wikidata_qid = '`+matrixQID+`'`); n != 0 {
		t.Fatal("a rerun created the held item")
	}

	// an admin who confirms the id is wrong on this record lets it in, without the contested id
	mustExec(t, db, `UPDATE item_conflicts SET status = 'allowed'`)
	repend(t, db, "2 minutes")
	drain(t, db, (&normSpy{}).job(0))

	if n := count(
		t,
		db,
		`SELECT count(*) FROM items WHERE wikidata_qid = '`+matrixQID+`' AND NOT (external_ids ? 'imdb')
AND external_ids->>'tmdb' = '603'`,
	); n != 1 {
		t.Fatal("the allowed item must exist without the contested imdb id and keep its other ids")
	}
}

func TestResolveDismissedStaysOut(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	owner(t, db, `{"imdb":"tt0133093"}`)

	drain(t, db, (&normSpy{}).job(0))
	mustExec(t, db, `UPDATE item_conflicts SET status = 'dismissed'`)
	repend(t, db, "1 minute")
	drain(t, db, (&normSpy{}).job(0))

	if n := count(t, db, `SELECT count(*) FROM items WHERE wikidata_qid = '`+matrixQID+`'`); n != 0 {
		t.Fatal("a dismissed duplicate was created")
	}
}

// The TMDB id is per media type: a film and a series may share the number.
func TestResolveTMDBIDIsScopedToTheType(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	mustExec(
		t,
		db,
		`INSERT INTO items (id, type, slug, wikidata_qid, external_ids) VALUES ($1, 'series', 'owner', 'Q1', '{"tmdb":"603"}'::jsonb)`,
		otherItem,
	)

	drain(t, db, (&normSpy{}).job(0))

	if n := count(t, db, `SELECT count(*) FROM items WHERE wikidata_qid = '`+matrixQID+`'`); n != 1 {
		t.Fatal("a series owning TMDB 603 must not block the film with TMDB 603")
	}
}

// A different QID with the same normalised English title and a release year within one is held as a near-duplicate.
func TestResolveHoldsANearDuplicateTitle(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	mustExec(
		t,
		db,
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')
ON CONFLICT DO NOTHING`,
	)
	mustExec(
		t,
		db,
		`INSERT INTO items (id, type, slug, wikidata_qid, release_date) VALUES ($1, 'movie', 'the-matrix', 'Q1', '2000-03-31')`,
		otherItem,
	)
	mustExec(
		t,
		db,
		`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', 'The Matrix', 'wikidata')`,
		otherItem,
	)

	drain(t, db, (&normSpy{}).job(0))

	if n := count(t, db, `SELECT count(*) FROM items WHERE wikidata_qid = '`+matrixQID+`'`); n != 0 {
		t.Fatal("the near-duplicate was created")
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM item_conflicts WHERE wikidata_qid = '`+matrixQID+`' AND kind = 'title_year'
AND other_item_id = '`+otherItem+`' AND similarity >= 0.92`,
	); n != 1 {
		t.Fatal("no title_year review was recorded")
	}
}

func TestResolveIgnoresTheSameTitleThreeYearsApart(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	mustExec(
		t,
		db,
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')
ON CONFLICT DO NOTHING`,
	)
	mustExec(
		t,
		db,
		`INSERT INTO items (id, type, slug, wikidata_qid, release_date) VALUES ($1, 'movie', 'the-matrix-2002', 'Q1', '2002-03-31')`,
		otherItem,
	)
	mustExec(
		t,
		db,
		`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', 'The Matrix', 'wikidata')`,
		otherItem,
	)

	drain(t, db, (&normSpy{}).job(0))

	if n := count(t, db, `SELECT count(*) FROM items WHERE wikidata_qid = '`+matrixQID+`'`); n != 1 {
		t.Fatal("a work of the same name three years apart is another work")
	}
}

// An item that already exists is never held; an id another item owns is dropped from the update.
func TestResolveDropsAContestedIDFromAnExistingItem(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)

	drain(t, db, (&normSpy{}).job(0)) // the film exists with imdb tt0133093

	mustExec(t, db, `UPDATE items SET external_ids = external_ids - 'imdb' WHERE wikidata_qid = '`+matrixQID+`'`)
	mustExec(
		t,
		db,
		`INSERT INTO items (id, type, slug, wikidata_qid, external_ids) VALUES ($1, 'movie', 'owner', 'Q1', '{"imdb":"tt0133093"}'::jsonb)`,
		otherItem,
	)

	repend(t, db, "1 minute")
	drain(t, db, (&normSpy{}).job(0))

	if n := count(
		t,
		db,
		`SELECT count(*) FROM items WHERE wikidata_qid = '`+matrixQID+`' AND NOT (external_ids ? 'imdb')`,
	); n != 1 {
		t.Fatal("the contested id was written to the existing item")
	}

	if n := count(
		t,
		db,
		`SELECT count(*) FROM item_conflicts WHERE wikidata_qid = '`+matrixQID+`' AND key = 'imdb'`,
	); n != 1 {
		t.Fatal("the contested id was not recorded for review")
	}
}

// A hidden item does not own its ids: a public item with the same id is created (FR-01-14).
func TestResolveAHiddenItemDoesNotOwnItsIDs(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	mustExec(t, db, `INSERT INTO items (id, type, slug, wikidata_qid, external_ids, visibility)
VALUES ($1, 'movie', 'owner', 'Q1', '{"imdb":"tt0133093"}'::jsonb, 'hidden')`, otherItem)

	drain(t, db, (&normSpy{}).job(0))

	if n := count(
		t,
		db,
		`SELECT count(*) FROM items WHERE wikidata_qid = '`+matrixQID+`' AND external_ids->>'imdb' = 'tt0133093'`,
	); n != 1 {
		t.Fatal("a hidden item blocked the public one")
	}

	if n := count(t, db, `SELECT count(*) FROM item_conflicts`); n != 0 {
		t.Fatalf("conflicts = %d, want none", n)
	}
}
