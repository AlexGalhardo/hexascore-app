//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestRedirect_OldSlugAnswers301(t *testing.T) {
	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	id := uuid.NewV7().String()
	merged := uuid.NewV7().String()

	for _, q := range []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')`, nil},
		{`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'game', 'half-life', 'Q1')`, []any{id}},
		{`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'game', 'half-life-old', 'Q2')`, []any{merged}},
	} {
		if _, err := db.ExecContext(ctx, q.sql, q.args...); err != nil {
			t.Fatal(err)
		}
	}

	if err := catalog.RenameSlug(ctx, db, catalog.ScopeItem, "game", id, "half-life-1"); err != nil {
		t.Fatal(err)
	}

	if err := catalog.RecordMerge(
		ctx,
		db,
		catalog.ScopeItem,
		"game",
		catalog.Target{ID: merged, Slug: "half-life-old"},
		id,
	); err != nil {
		t.Fatal(err)
	}

	c := func() *vukemtest.Client { return app.Client(t) }

	c().Get("/pt-br/games/half-life").AssertStatus(301).AssertRedirect("/pt-br/games/half-life-1")
	c().Get("/games/half-life?ref=x").AssertStatus(301).AssertRedirect("/games/half-life-1?ref=x")
	c().Get("/es/games/half-life-old").AssertStatus(301).AssertRedirect("/es/games/half-life-1")
	c().Get("/games/" + merged).AssertStatus(301).AssertRedirect("/games/half-life-1")
	// A missing item is a 301 home (owner decision 2026-11-01, item 14), never the redirect of another type.
	c().Get("/movies/half-life").AssertStatus(301).AssertHeader("Location", "/")
	c().Get("/es/games/never-existed").AssertStatus(301).AssertHeader("Location", "/es/")
}
