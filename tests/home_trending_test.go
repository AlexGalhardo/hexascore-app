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

func TestHome_Trending(t *testing.T) {
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

	item := func(i int, visibility string) string {
		id := uuid.NewV7().String()
		exec(`INSERT INTO items (id, type, slug, wikidata_qid, visibility, release_date, rating_count, score)
VALUES ($1, 'game', $2, $3, $4, '2020-05-01', 9, 87.25)`, id, fmt.Sprintf("game-%d", i), fmt.Sprintf("Q%d", 9000+i), visibility)
		exec(`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', $2, 'hexascore')`,
			id, fmt.Sprintf("Game %d", i))

		return id
	}
	snap := func(at, typ string, rank int, id, basis string) {
		exec(`INSERT INTO trending_snapshots (computed_at, item_type, rank, item_id, trend_score, basis)
VALUES ($1, $2, $3, $4, 1.5, $5)`, at, typ, rank, id, basis)
	}

	snap("2026-10-31T09:00:00Z", "game", 1, item(0, "public"), "trend") // older snapshot, ignored
	snap("2026-10-31T10:00:00Z", "game", 1, item(1, "public"), "trend")
	snap("2026-10-31T10:00:00Z", "game", 2, item(2, "public"), "prior")
	snap("2026-10-31T10:00:00Z", "game", 3, item(3, "hidden"), "trend") // hidden since the snapshot

	page := app.Client(t).Inertia("").Get("/").AssertOK().AssertInertia("Home/Index").
		HasProp("trending.game.items", 2).
		PropEquals("trending.game.computedAt", "2026-10-31T10:00:00Z").
		PropEquals("trending.game.items.0.rank", 1).
		PropEquals("trending.game.items.0.basis", "trend").
		PropEquals("trending.game.items.0.title", "Game 1").
		PropEquals("trending.game.items.0.slug", "game-1").
		PropEquals("trending.game.items.0.score", 8.7).
		PropEquals("trending.game.items.0.ratingCount", 9).
		PropEquals("trending.game.items.1.basis", "prior")
	// A type with no snapshot has an empty list, which the page renders as no section.
	page.HasProp("trending.movie.items", 0)
}
