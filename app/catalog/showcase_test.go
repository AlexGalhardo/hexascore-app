//go:build integration

package catalog_test

import (
	"context"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// Owner 2026-11-04: the landing page and the category listings hide items without a cover or without a visible
// score, and show them again once they have one. The API listing is not affected.
func TestShowcaseHidesItemsWithoutCoverOrScore(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
	})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	for _, q := range []string{
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata','Wikidata','CC0-1.0','https://www.wikidata.org')`,
		// complete: cover and 5+ ratings. nocover: score but no cover. noscore: cover but 4 ratings.
		`INSERT INTO items (id, type, slug, wikidata_qid, rating_count, score, cover_key, visibility) VALUES
 ('00000000-0000-7000-8000-000000000001','game','complete','Q1',10,80,'covers/game/complete.webp','public'),
 ('00000000-0000-7000-8000-000000000002','game','nocover','Q2',10,80,NULL,'public'),
 ('00000000-0000-7000-8000-000000000003','game','noscore','Q3',4,80,'covers/game/noscore.webp','public')`,
	} {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatal(err)
		}
	}

	slugs := func(sc catalog.Showcase) []string {
		hl, err := catalog.Highlights(ctx, db, "game", "en", sc)
		if err != nil {
			t.Fatal(err)
		}

		out := make([]string, len(hl.Pool))
		for i, h := range hl.Pool {
			out[i] = h.Slug
		}

		return out
	}

	if got := slugs(catalog.Showcase{}); len(got) != 3 {
		t.Errorf("without the flags every item shows: %v", got)
	}

	if got := slugs(catalog.Showcase{Cover: true, Score: true}); len(got) != 1 || got[0] != "complete" {
		t.Errorf("with both flags only the complete item shows: %v", got)
	}

	if got := slugs(catalog.Showcase{Cover: true}); len(got) != 2 {
		t.Errorf("cover rule alone hides only the item without a cover: %v", got)
	}

	hl, _ := catalog.Highlights(ctx, db, "game", "en", catalog.Showcase{Cover: true, Score: true})
	if hl.Count != 1 {
		t.Errorf("the section count follows the showcase: %d", hl.Count)
	}

	// The item without a score comes back with its fifth rating.
	if _, err := db.ExecContext(ctx, `UPDATE items SET rating_count = 5 WHERE slug = 'noscore'`); err != nil {
		t.Fatal(err)
	}

	if got := slugs(catalog.Showcase{Cover: true, Score: true}); len(got) != 2 {
		t.Errorf("a scored item shows again: %v", got)
	}

	list := func(sc catalog.Showcase) int {
		page, err := catalog.List(ctx, db, nil, catalog.ListParams{Type: "game", Limit: 20, Page: 1, Showcase: sc})
		if err != nil {
			t.Fatal(err)
		}

		return len(page.Items)
	}

	if all, shown := list(
		catalog.Showcase{},
	), list(
		catalog.Showcase{Cover: true, Score: true},
	); all != 3 ||
		shown != 2 {
		t.Errorf("listing: all %d, showcase %d, want 3 and 2", all, shown)
	}
}
