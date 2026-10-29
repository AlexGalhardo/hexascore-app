//go:build integration

package tests

import (
	"fmt"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestInertiaPages_Entity_Show(t *testing.T) {
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

	person, genre, country := uuid.NewV7().String(), uuid.NewV7().String(), uuid.NewV7().String()
	exec(`INSERT INTO entities (id, kind, slug, code, default_name) VALUES
  ($1, 'person', 'ridley-scott', NULL, 'Ridley Scott'), ($2, 'genre', 'rpg', NULL, 'RPG'),
  ($3, 'country', 'brazil', 'BR', 'Brazil')`, person, genre, country)

	qid := 9000
	item := func(typ, vis string, n, count int, score any, entity, role string) {
		t.Helper()

		id := uuid.NewV7().String()
		slug := fmt.Sprintf("%s-%d-%s", typ, n, entity[:4])
		qid++
		exec(
			`INSERT INTO items (id, type, slug, wikidata_qid, visibility, rating_count, score) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			id,
			typ,
			slug,
			fmt.Sprintf("Q%d", qid),
			vis,
			count,
			score,
		)
		exec(`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', $2, 'hexascore')`,
			id, slug)
		exec(`INSERT INTO item_entities (item_id, role, entity_id, source) VALUES ($1, $2, $3, 'hexascore')`,
			id, role, entity)
	}

	// Two public movies and a hidden one: below the threshold of 3 linked public items.
	item("movie", "public", 1, 0, nil, person, "director")
	item("movie", "public", 2, 0, nil, person, "director")
	item("movie", "hidden", 3, 0, nil, person, "director")

	// Five scored games and one movie: indexable, with a visible average for games only.
	for n := range 5 {
		item("game", "public", n, 5, 80.0, genre, "genre")
	}
	item("movie", "public", 9, 0, nil, genre, "genre")
	if err := catalog.RefreshEntityStats(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	client := app.Client(t)

	client.Inertia("").Get("/people/ridley-scott").AssertOK().AssertHeader("X-Robots-Tag", "noindex, follow").
		AssertInertia("Entity/Show").
		PropEquals("entity.name", "Ridley Scott").
		PropEquals("entity.kind", "person").
		PropEquals("counts.movie", 2).
		PropEquals("indexable", false).
		PropEquals("roles.0.role", "director").
		PropEquals("roles.0.count", 2).
		HasProp("items.data", 1).
		HasProp("items.data.0.items", 2).
		PropEquals("items.data.0.type", "movie")

	client.Inertia("").Get("/pt-br/genres/rpg").AssertOK().AssertHeaderMissing("X-Robots-Tag").
		AssertInertia("Entity/Show").
		PropEquals("indexable", true).
		PropEquals("counts.game", 5).
		PropEquals("avgScore.game", 8.0).
		HasProp("items.data", 2).
		MissingProp("avgScore.movie")

	client.Inertia("").Get("/genres/rpg?type=game&cursor=3").AssertOK().AssertInertia("Entity/Show").
		HasProp("items.data", 1).
		HasProp("items.data.0.items", 2)

	client.Inertia("").Get("/countries/BR").AssertOK().AssertInertia("Entity/Show").
		PropEquals("entity.code", "BR")
	client.Inertia("").Get("/people/missing").AssertStatus(404)
	client.Inertia("").Get("/genres/ridley-scott").AssertStatus(404)
}
