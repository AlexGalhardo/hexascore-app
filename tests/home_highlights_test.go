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
		PropEquals("sections.0.newest.0.score", nil)
	page.HasProp("sections.1.topRated", 0)
}
