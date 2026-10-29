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

	for i := range 30 {
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
		HasProp("items.data", 24).PropEquals("total", 30)
	scroll, _ := first.Page()["scrollProps"].(map[string]any)["items"].(map[string]any)
	cursor, _ := scroll["nextPage"].(string)
	if cursor == "" {
		t.Fatal("first page has no opaque next cursor")
	}

	c.Get("/movies?cursor="+cursor).AssertOK().AssertInertia("Catalog/Index").HasProp("items.data", 6)
	c.Get("/movies?cursor=forged").AssertStatus(400)

	c.Get("/movies?genre=rpg").AssertOK().AssertInertia("Catalog/Index").HasProp("items.data", 3).PropEquals("total", 3)
	c.Get("/movies?sort=bogus").AssertStatus(400)
}
