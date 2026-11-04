//go:build integration

package catalog_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

var repairCases = []struct{ typ, label string }{
	{"movie", "gangster film"},
	{"movie", "romantic comedy film"},
	{"movie", "flashback film"},
	{"movie", "American football film"},
	{"series", "crime thriller television series"},
	{"series", "children's television series"},
	{"game", "role-playing video game"},
	{"game", "first-person shooter"},
	{"game", "Sci-Fi game"},
	{"book", "science fiction novel"},
	{"book", "young adult fiction"},
	{"book", "poetry collection"},
	{"book", "XYZ unmapped label"},
}

// The migration SQL must do what CanonicalizeGenres does: every label lands on the canonical genre GenreFor picks, or
// is hidden when no rule maps it, with its source link kept.
func TestGenreRepairSQLMatchesTheGoRules(t *testing.T) {
	ctx := context.Background()
	db := entityDB(t)

	type fixture struct{ item, entity string }

	fx := map[string]fixture{}

	for i, c := range repairCases {
		key := c.typ + "/" + c.label
		item := newItem(t, db, c.typ, fmt.Sprintf("item-%d", i), fmt.Sprintf("Q%d", 100+i), "public")
		ent := newEntity(t, db, "genre", fmt.Sprintf("source-genre-%d", i))

		if _, err := db.ExecContext(
			ctx,
			`UPDATE entities SET default_name = $2 WHERE id = $1`,
			ent,
			c.label,
		); err != nil {
			t.Fatal(err)
		}

		if _, err := db.ExecContext(ctx, `INSERT INTO item_entities (item_id, entity_id, role, position, source)
VALUES ($1, $2, 'genre', 0, 'wikidata')`, item, ent); err != nil {
			t.Fatal(err)
		}

		fx[key] = fixture{item, ent}
	}

	for range 2 { // the second pass proves it is idempotent
		for _, st := range catalog.GenreRepairStatements() {
			if _, err := db.ExecContext(ctx, st.SQL, st.Args...); err != nil {
				t.Fatalf("%v\n%s", err, st.SQL)
			}
		}
	}

	for _, c := range repairCases {
		f := fx[c.typ+"/"+c.label]
		want, mapped := catalog.GenreFor(c.typ, c.label)

		var got []string

		rows, err := db.QueryContext(ctx, `SELECT e.code FROM item_entities ie JOIN entities e ON e.id = ie.entity_id
WHERE ie.item_id = $1 AND ie.role = 'genre'`, f.item)
		if err != nil {
			t.Fatal(err)
		}

		for rows.Next() {
			var code *string
			if err := rows.Scan(&code); err != nil {
				t.Fatal(err)
			}

			if code != nil {
				got = append(got, *code)
			}
		}

		if err := rows.Close(); err != nil {
			t.Fatal(err)
		}

		switch {
		case mapped && (len(got) != 1 || got[0] != want):
			t.Errorf("%s %q: genres %v, want [%s]", c.typ, c.label, got, want)
		case !mapped && len(got) != 0:
			t.Errorf("%s %q: genres %v, want none (hidden)", c.typ, c.label, got)
		}

		var kept int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM item_genre_sources WHERE item_id = $1 AND entity_id = $2`,
			f.item, f.entity).
			Scan(&kept); err != nil ||
			kept != 1 {
			t.Errorf("%s %q: source link kept %d times (%v), want once", c.typ, c.label, kept, err)
		}
	}

	needed := map[string]bool{}

	for _, c := range repairCases {
		if code, ok := catalog.GenreFor(c.typ, c.label); ok {
			needed[code] = true
		}
	}

	var names int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM entity_translations t JOIN entities e ON e.id = t.entity_id
WHERE e.kind = 'genre' AND t.name_source = 'hexascore'`).Scan(&names); err != nil || names != 3*len(needed) {
		t.Errorf("canonical genre names = %d (%v), want %d: only the genres some item needs are created", names, err,
			3*len(needed))
	}
}

// The normaliser canonicalises each item it saves, so a new ingest never leaves a source label behind.
func TestCanonicalizeItemGenresKeepsNewIngestsCanonical(t *testing.T) {
	ctx := context.Background()
	db := entityDB(t)

	item := newItem(t, db, "movie", "new-film", "Q900", "public")
	ent := newEntity(t, db, "genre", "gangster-film")

	if _, err := db.ExecContext(
		ctx,
		`UPDATE entities SET default_name = 'gangster film' WHERE id = $1`,
		ent,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO item_entities (item_id, entity_id, role, position, source)
VALUES ($1, $2, 'genre', 0, 'wikidata')`, item, ent); err != nil {
		t.Fatal(err)
	}

	if err := catalog.CanonicalizeItem(ctx, db, item); err != nil {
		t.Fatal(err)
	}

	var code string
	if err := db.QueryRowContext(ctx, `SELECT e.code FROM item_entities ie JOIN entities e ON e.id = ie.entity_id
WHERE ie.item_id = $1 AND ie.role = 'genre'`, item).Scan(&code); err != nil || code != "crime" {
		t.Errorf("genre = %q (%v), want crime", code, err)
	}
}
