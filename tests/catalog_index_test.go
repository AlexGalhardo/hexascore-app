//go:build integration

package tests

import (
	"fmt"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestInertiaPages_Catalog_Index(t *testing.T) {
	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	exec := func(q string, args ...any) {
		t.Helper()

		if _, err := db.ExecContext(t.Context(), q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	exec(`INSERT INTO data_sources (code, name, licence, url) VALUES ('hexascore', 'x', 'proprietary', 'https://x')`)

	genre := uuid.NewV7().String()
	exec(`INSERT INTO entities (id, kind, slug, default_name) VALUES ($1, 'genre', 'rpg', 'RPG')`, genre)

	for i := range 60 {
		id := uuid.NewV7().String()
		exec(`INSERT INTO items (id, type, slug, wikidata_qid, rating_count, score)
VALUES ($1, 'movie', $2, $3, 6, 80)`, id, fmt.Sprintf("movie-%02d", i), fmt.Sprintf("Q%d", 8000+i))
		exec(`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', $2, 'hexascore')`,
			id, fmt.Sprintf("Movie %02d", i))

		if i < 3 {
			exec(
				`INSERT INTO item_entities (item_id, role, entity_id, source) VALUES ($1, 'genre', $2, 'hexascore')`,
				id,
				genre,
			)
		}
	}

	c := app.Client(t).Inertia("")

	first := c.Get("/movies").AssertOK().AssertInertia("Catalog/Index").
		HasProp("items.data", 40).PropEquals("total", 60).PropEquals("filters.per", 40)
	slug, ok := first.Prop("facets.genres.0.slug")
	if !ok || slug != "rpg" {
		t.Fatalf("first paint facets slug = %#v", slug)
	}
	count, ok := first.Prop("facets.genres.0.count")
	if !ok || fmt.Sprint(count) != "3" {
		t.Fatalf("first paint facet count = %#v", count)
	}
	// Paged with buttons, not infinite scroll (owner round 2026-11-01 items 12 and 20), by readable page numbers
	// with the total on every page (owner 2026-11-02); page size from ?per= (40, 80 or 120; anything else is 40).
	first.PropEquals("pagination.prev", "").PropEquals("pagination.next", "2")

	c.Get("/movies?per=80&page=2").AssertOK().AssertInertia("Catalog/Index").
		HasProp("items.data", 0).PropEquals("pagination.prev", "1").PropEquals("pagination.next", "").
		PropEquals("total", 60).PropEquals("filters.per", 80).PropEquals("lastPage", 1)
	c.Get("/movies?page=2").AssertOK().AssertInertia("Catalog/Index").
		HasProp("items.data", 20).PropEquals("pagination.next", "").PropEquals("total", 60)
	// Past the end: no items, and lastPage lets the page link back (W-06).
	c.Get("/movies?page=9").AssertOK().AssertInertia("Catalog/Index").HasProp("items.data", 0).
		PropEquals("total", 60).PropEquals("lastPage", 2)

	for _, bad := range []string{"0", "-3", "abc"} {
		c.Get("/movies?page="+bad).AssertOK().AssertInertia("Catalog/Index").PropEquals("pagination.prev", "")
	}

	c.Get("/movies?page=10001").AssertOK().AssertInertia("Catalog/Index").HasProp("items.data", 0)
	// An old cursor link no longer pages; it lands on page 1.
	c.Get("/movies?cursor=forged").AssertOK().AssertInertia("Catalog/Index").PropEquals("pagination.prev", "")

	for _, tc := range []struct {
		per          string
		items, given int
	}{
		{"80", 60, 80},
		{"120", 60, 120},
		{"40", 40, 40},
		{"25", 40, 40},
		{"100", 40, 40},
		{"abc", 40, 40},
		{"-1", 40, 40},
		{"1000", 40, 40},
	} {
		c.Get("/movies?per="+tc.per).AssertOK().AssertInertia("Catalog/Index").
			HasProp("items.data", tc.items).PropEquals("filters.per", tc.given)
	}

	c.Get("/movies?genre=rpg").AssertOK().AssertInertia("Catalog/Index").HasProp("items.data", 3).PropEquals("total", 3)
	c.Get("/movies?sort=bogus").AssertStatus(400)
}

// TestServerError_PageLinksHome: a full-page 5xx keeps its status and error page, now with a link to the home page
// of the locale (owner round 2026-11-01 item 14); an Inertia visit gets the Error component with the status.
func TestServerError_PageLinksHome(t *testing.T) {
	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	// The listing query fails once its table is gone: a real 500 without a test-only route.
	if _, err := db.ExecContext(t.Context(), `ALTER TABLE items RENAME TO items_gone`); err != nil {
		t.Fatal(err)
	}

	app.Client(t).Get("/movies").AssertStatus(500).AssertSee(`<a href="/">Back to home</a>`)
	app.Client(t).Get("/pt-br/movies").AssertStatus(500).AssertSee(`<a href="/pt-br/">`)
	app.Client(t).Inertia("").Get("/movies").AssertStatus(500).AssertInertia("Error").PropEquals("status", 500)
}
