//go:build integration

package catalog_test

import (
	"context"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

func TestCanonicalizeGenres(t *testing.T) {
	ctx := context.Background()
	db := entityDB(t)

	movie := newItem(t, db, "movie", "the-godfather", "Q47703", "public")
	gangster := newEntity(t, db, "genre", "gangster-film")
	drama := newEntity(t, db, "genre", "drama-film")
	flashback := newEntity(t, db, "genre", "flashback-film")
	for id, name := range map[string]string{gangster: "gangster film", drama: "drama film", flashback: "flashback film"} {
		if _, err := db.ExecContext(ctx, `UPDATE entities SET default_name = $2 WHERE id = $1`, id, name); err != nil {
			t.Fatal(err)
		}
	}
	for i, e := range []string{gangster, drama, flashback} {
		if _, err := db.ExecContext(ctx, `INSERT INTO item_entities (item_id, entity_id, role, position, source)
VALUES ($1, $2, 'genre', $3, 'wikidata')`, movie, e, i); err != nil {
			t.Fatal(err)
		}
	}

	genres := func() []string {
		rows, err := db.QueryContext(ctx, `SELECT e.slug FROM item_entities ie JOIN entities e ON e.id = ie.entity_id
WHERE ie.item_id = $1 AND ie.role = 'genre' ORDER BY e.slug`, movie)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var out []string
		for rows.Next() {
			var s string
			if err := rows.Scan(&s); err != nil {
				t.Fatal(err)
			}
			out = append(out, s)
		}
		return out
	}

	dry, err := catalog.CanonicalizeGenres(ctx, db, true)
	if err != nil {
		t.Fatal(err)
	}
	if dry.Links != 3 || dry.Mapped != 2 || dry.Hidden != 1 || dry.Unmapped["movie: flashback film"] != 1 {
		t.Fatalf("dry run report = %+v", dry)
	}
	if got := genres(); len(got) != 3 || got[0] != "drama-film" {
		t.Fatalf("dry run changed the links: %v", got)
	}

	rep, err := catalog.CanonicalizeGenres(ctx, db, false)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Created != len(catalog.Genres) || rep.Mapped != 2 || rep.Hidden != 1 {
		t.Fatalf("report = %+v", rep)
	}
	if got := genres(); len(got) != 2 || got[0] != "crime" || got[1] != "drama" {
		t.Fatalf("genres = %v, want [crime drama]", got)
	}

	var kept int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM item_genre_sources WHERE item_id = $1`, movie).Scan(&kept); err != nil || kept != 3 {
		t.Fatalf("item_genre_sources = %d, %v; want the 3 source links kept", kept, err)
	}

	var pt string
	if err := db.QueryRowContext(ctx, `SELECT t.name FROM entity_translations t JOIN entities e ON e.id = t.entity_id
WHERE e.kind = 'genre' AND e.code = 'crime' AND t.locale = 'es'`).Scan(&pt); err != nil || pt != "Crimen" {
		t.Fatalf("es name of crime = %q, %v", pt, err)
	}

	facets, err := catalog.FacetsOf(ctx, db, "movie", "genre", "pt-BR", nil)
	if err != nil || len(facets) != 2 || !facets[0].Canonical {
		t.Fatalf("facets = %+v, %v", facets, err)
	}

	again, err := catalog.CanonicalizeGenres(ctx, db, false)
	if err != nil || again.Links != 0 || again.Created != 0 || again.Adopted != 0 {
		t.Fatalf("second run = %+v, %v; want nothing left to do", again, err)
	}
}
