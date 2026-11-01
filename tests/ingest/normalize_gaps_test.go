//go:build integration

package ingest_test

import "testing"

func TestNormalizeCreditsOnlyChangeWritesEntitiesRow(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	drain(t, db, (&normSpy{}).job(0))

	var id string
	if err := db.QueryRowContext(t.Context(), `SELECT item_id::text FROM item_entities WHERE role = 'director'
LIMIT 1`).Scan(&id); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), `DELETE FROM item_entities WHERE item_id = $1 AND role = 'director'`,
		id); err != nil {
		t.Fatal(err)
	}

	before := count(t, db, `SELECT count(*) FROM api_changes`)

	repend(t, db, "1 minute")
	drain(t, db, (&normSpy{}).job(0))

	var fields string
	if err := db.QueryRowContext(t.Context(), `SELECT array_to_string(fields, ',') FROM api_changes
WHERE item_id = $1 AND action = 'updated'`, id).Scan(&fields); err != nil || fields != "entities" {
		t.Errorf("updated fields = %q, %v; want entities", fields, err)
	}

	if n := count(t, db, `SELECT count(*) FROM api_changes`); n != before+1 {
		t.Errorf("api_changes rows = %d, want %d (only the credits-only item)", n, before+1)
	}

	// An unchanged re-run adds nothing.
	repend(t, db, "2 minutes")
	drain(t, db, (&normSpy{}).job(0))

	if n := count(t, db, `SELECT count(*) FROM api_changes`); n != before+1 {
		t.Errorf("api_changes rows after an unchanged re-run = %d, want %d", n, before+1)
	}
}

// A hidden or takedown item answers 404 on the API (FR-01-14), so the feed must not name it: ingestion keeps
// updating it silently, and making it public again announces it as created.
func TestNormalizeNonPublicItemWritesNoRowAndReleaseAnnounces(t *testing.T) {
	db := newDB(t)
	loadWikidata(t, db)
	drain(t, db, (&normSpy{}).job(0))

	var id string
	if err := db.QueryRowContext(t.Context(), `SELECT id::text FROM items WHERE wikidata_qid = 'Q83495'`).
		Scan(&id); err != nil {
		t.Fatal(err)
	}

	for _, vis := range []string{"hidden", "takedown"} {
		if _, err := db.ExecContext(t.Context(), `UPDATE items SET visibility = $2, runtime_minutes = 1 WHERE id = $1`,
			id, vis); err != nil {
			t.Fatal(err)
		}

		before := count(t, db, `SELECT count(*) FROM api_changes`)

		repend(t, db, "1 minute")
		drain(t, db, (&normSpy{}).job(0))

		if n := count(t, db, `SELECT count(*) FROM api_changes`); n != before {
			t.Errorf("%s item: api_changes rows %d -> %d, want none written", vis, before, n)
		}
	}

	if n := count(t, db, `SELECT count(*) FROM api_changes WHERE item_id = $1 AND action = 'deleted'`, id); n != 1 {
		t.Errorf("deleted rows after leaving public = %d, want 1 (hidden to takedown adds none)", n)
	}

	if _, err := db.ExecContext(t.Context(), `UPDATE items SET visibility = 'public' WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}

	if n := count(t, db, `SELECT count(*) FROM api_changes WHERE item_id = $1 AND action = 'created'`, id); n != 2 {
		t.Errorf("created rows after going public again = %d, want 2", n)
	}
}
