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
		vukemtest.Options{Env: showEverythingEnv(pgtest.Database(t, ""))},
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
  ($1, 'person', 'ridley-scott', NULL, 'Ridley Scott'), ($2, 'franchise', 'rpg', NULL, 'RPG'),
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
		item("game", "public", n, 5, 80.0, genre, "franchise")
	}
	item("movie", "public", 9, 0, nil, genre, "franchise")
	if err := catalog.RefreshEntityStats(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	client := app.Client(t)

	// No entity has a page of its own (owner 2026-11-02): every entity path answers one 301 to the section listing
	// filtered by the entity, whatever the kind and the locale.
	for path, target := range map[string]string{
		"/movies/people/ridley-scott":       "/movies?per=40&person=ridley-scott&sort=-popularity",
		"/pt-br/games/franchises/rpg":       "/pt-br/games?franchise=rpg&per=40&sort=-popularity",
		"/es/games/countries/br":            "/es/games?country=br&per=40&sort=-popularity",
		"/games/game-modes/multiplayer":     "/games?mode=multiplayer&per=40&sort=-popularity",
		"/games/companies/wright-systems":   "/games?company=wright-systems&per=40&sort=-popularity",
		"/games/genres/action-role-playing": "/games?genre=action-role-playing&per=40&sort=-popularity",
	} {
		client.Get(path).AssertStatus(301).AssertHeader("Location", target)
	}

	// The old top-level paths go straight to the listing of the category with the most public items of the entity,
	// or to ?type= when it names one, with the canonical code for countries and languages.
	for _, prefix := range []string{"", "/pt-br", "/es"} {
		for old, target := range map[string]string{
			"/people/ridley-scott":       "/movies?per=40&person=ridley-scott&sort=-popularity",
			"/franchises/rpg":            "/games?franchise=rpg&per=40&sort=-popularity",
			"/franchises/rpg?type=movie": "/movies?franchise=rpg&per=40&sort=-popularity",
			"/countries/br?type=series":  "/series?country=BR&per=40&sort=-popularity",
			"/countries/br":              "/games?country=BR&per=40&sort=-popularity",
		} {
			client.Get(prefix+old).AssertStatus(301).AssertHeader("Location", prefix+target)
		}
	}

	// The listing filters by any role of the entity and names it in a chip; hidden items stay out.
	client.Inertia("").Get("/movies?person=ridley-scott").AssertOK().AssertInertia("Catalog/Index").
		PropEquals("total", 2).PropEquals("active.0.param", "person").PropEquals("active.0.name", "Ridley Scott")
	client.Inertia("").Get("/games?franchise=rpg").AssertOK().AssertInertia("Catalog/Index").PropEquals("total", 5)
	item("movie", "public", 10, 0, nil, country, "country")
	client.Inertia("").Get("/movies?country=br").AssertOK().AssertInertia("Catalog/Index").PropEquals("total", 1).
		PropEquals("active.0.name", "Brazil")

	// The entity's recent changes keep their page under the category.
	client.Inertia("").Get("/movies/people/ridley-scott/changes").AssertOK().AssertInertia("Catalog/RecentChanges")
	client.Get("/pt-br/people/ridley-scott/changes").AssertStatus(301).
		AssertHeader("Location", "/pt-br/movies/people/ridley-scott/changes")

	// A missing entity on an old top-level path is a missing page.
	client.Inertia("").Get("/people/missing").AssertStatus(404).AssertInertia("Error").PropEquals("status", 404)
}
