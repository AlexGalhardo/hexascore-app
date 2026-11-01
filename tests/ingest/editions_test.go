//go:build integration

package ingest_test

import (
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/editions"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
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
