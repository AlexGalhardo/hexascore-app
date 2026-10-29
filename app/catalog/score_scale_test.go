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

// A stored 87 must read 8.7 on landing highlights, listing summaries and the item page alike.
func TestScoreScaleConsistent(t *testing.T) {
	url := pgtest.Database(t, "")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": url}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	const id = "00000000-0000-7000-8000-000000000001"
	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata','Wikidata','CC0-1.0','https://www.wikidata.org')`,
	); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO items (id, type, slug, wikidata_qid, popularity_prior, rating_count, score, visibility)
VALUES ($1,'game','scaled','Q1',1,10,87,'public')`, id); err != nil {
		t.Fatal(err)
	}

	hl, err := catalog.Highlights(ctx, db, "game", "en")
	if err != nil || len(hl.TopRated) != 1 || hl.TopRated[0].Score == nil {
		t.Fatalf("highlights: %v %+v", err, hl)
	}
	sums, err := catalog.Summaries(ctx, db, []string{id}, "en")
	if err != nil || len(sums) != 1 || sums[0].Score == nil {
		t.Fatalf("summaries: %v %+v", err, sums)
	}
	page, err := catalog.LoadItemPage(ctx, db, "game", "scaled", "en")
	if err != nil || page.RatingSummary.Score == nil {
		t.Fatalf("item page: %v", err)
	}
	for name, got := range map[string]float64{"highlights": *hl.TopRated[0].Score, "summaries": *sums[0].Score, "item": *page.RatingSummary.Score} {
		if got != 8.7 {
			t.Errorf("%s score = %v, want 8.7", name, got)
		}
	}
}
