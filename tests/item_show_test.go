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

func TestInertiaPages_Catalog_Show(t *testing.T) {
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

	game, book := uuid.NewV7().String(), uuid.NewV7().String()
	exec(
		`INSERT INTO items (id, type, slug, wikidata_qid, rating_count, score, cover_key, cover_thumb_key)
		 VALUES ($1, 'game', 'zelda', 'Q9001', 9, 87.25, 'covers/zelda-0123456789abcdef.webp', 'covers/zelda-0123456789abcdef-t.webp')`,
		game,
	)
	exec(`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'book', 'dune', 'Q9002')`, book)
	for id, title := range map[string]string{game: "Zelda", book: "Dune"} {
		exec(
			`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', $2, 'hexascore')`,
			id,
			title,
		)
	}
	for i := range 12 {
		exec(
			`INSERT INTO editions (id, work_id, openlibrary_id, title, language, source) VALUES ($1, $2, $3, $4, 'en', 'hexascore')`,
			uuid.NewV7().String(),
			book,
			fmt.Sprintf("OL%dM", i),
			fmt.Sprintf("Dune %d", i),
		)
	}

	client := app.Client(t)

	client.Inertia("").Get("/games/zelda").AssertOK().AssertInertia("Catalog/Show").
		PropEquals("item.title", "Zelda").
		PropEquals("ratingSummary.score", 8.7).
		PropEquals("ratingSummary.ratingCount", 9).
		HasProp("timeToBeat.reports").
		PropEquals("cover.full", "/covers/zelda-0123456789abcdef.webp").
		PropEquals("cover.thumb", "/covers/zelda-0123456789abcdef-t.webp").
		MissingProp("me")

	client.Inertia("").Get("/pt-br/books/dune").AssertOK().AssertInertia("Catalog/Show").
		PropEquals("editionCount", 12).
		HasProp("editions", 10)

	client.Inertia("").Get("/games/missing").AssertStatus(404)
}
