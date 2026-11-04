//go:build integration

package normalize_test

import (
	"context"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/normalize"
)

// Production 2026-11-04: titles that differ only in case or accents collide on the (item_id, normalised title) index; the
// old insert named (item_id, title) and failed the item's whole relations step with SQLSTATE 23505.
func TestAltTitlesThatNormaliseAlikeAreStoredOnce(t *testing.T) {
	ctx := context.Background()
	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org')`,
	); err != nil {
		t.Fatal(err)
	}

	id := uuid.NewV7().String()
	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'movie', 'm-1', 'Q1')`,
		id,
	); err != nil {
		t.Fatal(err)
	}

	rel := &normalize.Relations{AltTitles: []normalize.AltTitle{
		{Title: "Amélie", Lang: "fr", Kind: "alias", Source: "wikidata"},
		{Title: "AMELIE", Lang: "en", Kind: "alias", Source: "wikidata"},
		{Title: "Amélie", Lang: "fr", Kind: "alias", Source: "wikidata"},
	}}

	if err := normalize.SaveRelations(ctx, db, id, rel, map[string]string{}); err != nil {
		t.Fatalf("relations must not fail on a normalised duplicate: %v", err)
	}

	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM item_alt_titles WHERE item_id = $1`, id).
		Scan(&n); err != nil {
		t.Fatal(err)
	}

	if n != 1 {
		t.Fatalf("alt titles stored = %d, want 1", n)
	}
}
