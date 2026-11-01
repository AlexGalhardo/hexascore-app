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

func TestHome_Highlights(t *testing.T) {
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

	// i=0 has a visible score, i=1 has too few ratings, i=2 is hidden.
	for i, row := range []struct {
		visibility string
		count      int
		score      float64
	}{{"public", 9, 87.25}, {"public", 2, 99}, {"hidden", 50, 95}} {
		id := uuid.NewV7().String()
		exec(`INSERT INTO items (id, type, slug, wikidata_qid, visibility, release_date, rating_count, score)
VALUES ($1, 'game', $2, $3, $4, '2020-05-01', $5, $6)`,
			id, fmt.Sprintf("game-%d", i), fmt.Sprintf("Q%d", 7000+i), row.visibility, row.count, row.score)
		exec(`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', $2, 'hexascore')`,
			id, fmt.Sprintf("Game %d", i))
	}

	page := app.Client(t).Inertia("").Get("/pt-br/").AssertOK().AssertInertia("Home/Index").
		PropEquals("sections.0.type", "game").
		PropEquals("sections.0.count", 2).
		HasProp("sections.0.topRated", 1).
		HasProp("sections.0.newest", 2).
		PropEquals("sections.0.topRated.0.title", "Game 0").
		PropEquals("sections.0.topRated.0.score", 8.7).
		PropEquals("sections.0.mostRated.0.ratingCount", 9).
		PropEquals("sections.0.newest.0.score", nil).
		MissingProp("recent") // owner round 2026-11-01 item 5: the strip and its query are gone
	page.HasProp("sections.1.topRated", 0)
}

// TestHome_Trending7d: the home strip ranks the last 7 days only, 3 per type; activity 20 days ago does not count.
func TestHome_Trending7d(t *testing.T) {
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

	// Games 0-3 are active today with growing ratings; game 4 only 20 days ago, with far more activity.
	for i := range 5 {
		id := uuid.NewV7().String()
		exec(`INSERT INTO items (id, type, slug, wikidata_qid, release_date)
VALUES ($1, 'game', $2, $3, '2020-05-01')`, id, fmt.Sprintf("trend-%d", i), fmt.Sprintf("Q%d", 8000+i))
		exec(`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', $2, 'hexascore')`,
			id, fmt.Sprintf("Trend %d", i))

		if i == 3 {
			// Two genres: the one at the lower position is the primary one shown on the card.
			for pos, g := range []string{"adventure", "action"} {
				gid := uuid.NewV7().String()
				exec(`INSERT INTO entities (id, kind, slug, default_name) VALUES ($1, 'genre', $2, $3)`,
					gid, g, map[string]string{"action": "Action", "adventure": "Adventure"}[g])
				exec(`INSERT INTO item_entities (item_id, role, entity_id, source, position) VALUES ($1, 'genre', $2, 'hexascore', $3)`,
					id, gid, 1-pos)
			}
		}

		if i < 4 {
			exec(`INSERT INTO item_activity_daily (item_id, day, item_type, views, ratings_new_w)
VALUES ($1, current_date, 'game', 100, $2)`, id, 6+i)
		} else {
			exec(`INSERT INTO item_activity_daily (item_id, day, item_type, views, ratings_new_w)
VALUES ($1, current_date - 20, 'game', 9000, 60)`, id)
		}
	}

	app.Client(t).Inertia("").Get("/").AssertOK().AssertInertia("Home/Index").
		HasProp("trending.game.items", 3).
		PropEquals("trending.game.items.0.title", "Trend 3").
		PropEquals("trending.game.items.0.rank", 1).
		PropEquals("trending.game.items.0.basis", "trend").
		PropEquals("trending.game.items.0.genre", "Action").
		PropEquals("trending.game.items.1.genre", nil).
		PropEquals("trending.game.items.2.title", "Trend 1").
		HasProp("trending.movie.items", 0)
}
