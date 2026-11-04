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

// Owner 2026-11-04: the landing page and the category listings hide items without a cover. An item without a score
// stays; the page leaves out only its score badge. The API listing is not affected.
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
		t.Errorf("without the flag every item shows: %v", got)
	}

	// The item without a visible score stays: only its badge is hidden, by the page.
	if got := slugs(catalog.Showcase{Cover: true}); len(got) != 2 {
		t.Errorf("with the flag only the item without a cover is hidden: %v", got)
	}

	hl, _ := catalog.Highlights(ctx, db, "game", "en", catalog.Showcase{Cover: true})
	if hl.Count != 2 {
		t.Errorf("the section count follows the showcase: %d", hl.Count)
	}

	page, err := catalog.List(ctx, db, nil, catalog.ListParams{
		Type: "game", Limit: 20, Page: 1, Showcase: catalog.Showcase{Cover: true},
	})
	if err != nil || len(page.Items) != 2 {
		t.Errorf("listing with the flag: %v, %d items, want 2", err, len(page.Items))
	}
}
