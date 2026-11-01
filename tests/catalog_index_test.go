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
		HasProp("items.data", 25).PropEquals("total", 60).PropEquals("filters.per", 25)
	slug, ok := first.Prop("facets.genres.0.slug")
	if !ok || slug != "rpg" {
		t.Fatalf("first paint facets slug = %#v", slug)
	}
	count, ok := first.Prop("facets.genres.0.count")
	if !ok || fmt.Sprint(count) != "3" {
		t.Fatalf("first paint facet count = %#v", count)
	}
	// Paged with buttons, not infinite scroll (owner round 2026-11-01 items 12 and 20): signed cursors both
	// ways, page size from ?per= (25, 50, 75 or 100; anything else is 25).
	first.PropEquals("pagination.prev", "")
	cursorOf := func(r *vukemtest.InertiaPage, name string) string {
		t.Helper()

		v, _ := r.Prop("pagination." + name)
		s, _ := v.(string)
		if s == "" {
			t.Fatalf("no %s cursor", name)
		}

		return s
	}

	second := c.Get("/movies?per=50&cursor="+cursorOf(first, "next")).AssertOK().AssertInertia("Catalog/Index").
		HasProp("items.data", 35).PropEquals("pagination.next", "").PropEquals("filters.per", 50)
	c.Get("/movies?cursor="+cursorOf(second, "prev")).AssertOK().AssertInertia("Catalog/Index").
		HasProp("items.data", 25).PropEquals("pagination.prev", "")

	for _, tc := range []struct {
		per          string
		items, given int
	}{
		{"50", 50, 50},
		{"75", 60, 75},
		{"100", 60, 100},
		{"25", 25, 25},
		{"20", 25, 25},
		{"7", 25, 25},
		{"abc", 25, 25},
		{"-1", 25, 25},
		{"1000", 25, 25},
	} {
		c.Get("/movies?per="+tc.per).AssertOK().AssertInertia("Catalog/Index").
			HasProp("items.data", tc.items).PropEquals("filters.per", tc.given)
	}

	c.Get("/movies?cursor=forged").AssertStatus(400)

	c.Get("/movies?genre=rpg").AssertOK().AssertInertia("Catalog/Index").HasProp("items.data", 3).PropEquals("total", 3)
	c.Get("/movies?sort=bogus").AssertStatus(400)
}
