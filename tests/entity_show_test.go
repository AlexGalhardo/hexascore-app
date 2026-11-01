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

	// Entity pages live under their category (owner round 2026-11-01 item 19): /{section}/{entity path}/{key}.
	client.Inertia("").Get("/movies/people/ridley-scott").AssertOK().
		AssertHeader("X-Robots-Tag", "noindex, follow").
		AssertInertia("Entity/Show").
		PropEquals("entity.name", "Ridley Scott").
		PropEquals("entity.kind", "person").
		MissingProp("entity.wikidataQid").
		PropEquals("counts.movie", 2).
		PropEquals("indexable", false).
		PropEquals("roles.0.role", "director").
		PropEquals("roles.0.count", 2).
		PropEquals("filters.type", "movie").
		HasProp("items.data", 1).
		HasProp("items.data.0.items", 2).
		PropEquals("items.data.0.type", "movie")

	client.Inertia("").Get("/pt-br/games/genres/rpg").AssertOK().AssertHeaderMissing("X-Robots-Tag").
		AssertInertia("Entity/Show").
		PropEquals("indexable", true).
		PropEquals("counts.game", 5).
		PropEquals("counts.movie", 1).
		PropEquals("avgScore.game", 8.0).
		HasProp("items.data", 1).
		PropEquals("items.data.0.type", "game").
		MissingProp("avgScore.movie")

	// Indexable counts the category's own items: one movie is below the threshold.
	client.Inertia("").Get("/movies/genres/rpg").AssertOK().AssertHeader("X-Robots-Tag", "noindex, follow").
		AssertInertia("Entity/Show").PropEquals("indexable", false).HasProp("items.data.0.items", 1)

	client.Inertia("").Get("/games/genres/rpg?cursor=3").AssertOK().AssertInertia("Entity/Show").
		HasProp("items.data", 1).
		HasProp("items.data.0.items", 2)

	// No linked items yet: the group list is an empty array, never null (the page reads its length).
	client.Inertia("").Get("/games/countries/BR").AssertOK().AssertInertia("Entity/Show").
		PropEquals("entity.code", "BR").
		PropEquals("items.data", []any{})
	client.Inertia("").Get("/games/countries/br").AssertStatus(301).AssertHeader("Location", "/games/countries/BR")

	// The old top-level paths answer one 301 to the category with the most public items of the entity, or to
	// ?type= when it names one; the canonical code and the other query parameters travel along.
	for _, prefix := range []string{"", "/pt-br", "/es"} {
		for old, target := range map[string]string{
			"/people/ridley-scott":            "/movies/people/ridley-scott",
			"/genres/rpg":                     "/games/genres/rpg",
			"/genres/rpg?type=movie&cursor=0": "/movies/genres/rpg?cursor=0",
			"/genres/rpg?type=bogus":          "/games/genres/rpg",
			"/countries/br?type=series":       "/series/countries/BR",
			"/countries/br":                   "/games/countries/BR",
		} {
			client.Get(prefix+old).AssertStatus(301).AssertHeader("Location", prefix+target)
		}
	}

	// The entity's recent changes live under the category too; the old path keeps its /changes.
	client.Inertia("").Get("/movies/people/ridley-scott/changes").AssertOK().AssertInertia("Catalog/RecentChanges")
	client.Get("/pt-br/people/ridley-scott/changes").AssertStatus(301).
		AssertHeader("Location", "/pt-br/movies/people/ridley-scott/changes")

	// The item page links countries and languages by code, the key their entity route matches.
	item("movie", "public", 10, 0, nil, country, "country")
	client.Inertia("").Get("/movies/movie-10-"+country[:4]).AssertOK().AssertInertia("Catalog/Show").
		PropEquals("entities.country.0.code", "BR")
	client.Get("/countries/BR").AssertStatus(301).AssertHeader("Location", "/movies/countries/BR")

	// Missing entities, unknown entity paths and a kind mismatch are missing pages (item 14).
	for _, p := range []string{
		"/people/missing", "/movies/people/missing", "/games/genres/ridley-scott", "/games/no-such-kind/rpg",
		"/games/no-such-kind/rpg/changes",
	} {
		client.Inertia("").Get(p).AssertStatus(301).AssertHeader("Location", "/")
	}

	// A renamed entity slug answers 301 under its category too (SDD 01 §7.1).
	if err := catalog.RenameSlug(t.Context(), db, catalog.ScopeEntity, "genre", genre, "role-playing"); err != nil {
		t.Fatal(err)
	}

	client.Get("/es/games/genres/rpg?cursor=3").
		AssertStatus(301).
		AssertHeader("Location", "/es/games/genres/role-playing?cursor=3")
}
