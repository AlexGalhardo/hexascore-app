//go:build integration

package search_test

import (
	"context"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/schema"

	"github.com/AlexGalhardo/hexascore-app/database/migrations"
)

const searchMigrations = 7

func scalar(t *testing.T, db *database.DB, query string) string {
	t.Helper()

	var got string
	if err := db.QueryRowContext(context.Background(), query).Scan(&got); err != nil {
		t.Fatalf("%s: %v", query, err)
	}

	return got
}

func TestMigrationsUpDown(t *testing.T) {
	ctx := context.Background()

	db, err := database.Open(ctx, pgtest.Database(t, ""), database.WithDriver(postgres.Driver()))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() })

	m, err := schema.NewMigrator(db, migrations.All())
	if err != nil {
		t.Fatal(err)
	}

	for cycle := range 2 {
		if _, err := m.Up(ctx); err != nil {
			t.Fatalf("up (cycle %d): %v", cycle, err)
		}

		if got := scalar(t, db, `SELECT to_tsvector('hs_pt','Ação') = to_tsvector('hs_pt','acao')`); got != "true" {
			t.Fatalf("hs_pt unaccent: %s", got)
		}

		if got := scalar(t, db, `SELECT hs_norm('Pokémon: Edição Ouro!')`); got != "pokemon edicao ouro" {
			t.Fatalf("hs_norm = %q", got)
		}

		if cycle == 1 {
			break
		}

		rolled, err := m.Rollback(ctx, searchMigrations)
		if err != nil || len(rolled) != searchMigrations {
			t.Fatalf("rollback: %v, %v", rolled, err)
		}

		if got := scalar(
			t,
			db,
			`SELECT count(*) FROM pg_proc WHERE proname IN ('hs_norm','hs_unaccent','search_dirty_enqueue','item_translations_tsv')`,
		); got != "0" {
			t.Fatalf("functions left after rollback: %s", got)
		}
	}
}

func TestSearchTriggers(t *testing.T) {
	ctx := context.Background()

	db, err := database.Open(ctx, pgtest.Database(t, ""), database.WithDriver(postgres.Driver()))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() })

	m, err := schema.NewMigrator(db, migrations.All())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := m.Up(ctx); err != nil {
		t.Fatal(err)
	}

	for _, q := range []string{
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikipedia','Wikipedia','CC-BY-SA-4.0','https://wikipedia.org'), ('igdb','IGDB','custom','https://igdb.com')`,
	} {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatal(err)
		}
	}

	itemID := scalar(t, db, itemInsert)

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO item_translations (item_id, locale, title, title_source, synopsis, synopsis_source, synopsis_licence)
VALUES ($1, 'pt-BR', 'Ação Épica', 'igdb', 'Uma aventura', 'igdb', 'custom')`,
		itemID,
	); err != nil {
		t.Fatal(err)
	}

	if got := scalar(t, db, `SELECT title_norm FROM item_translations`); got != "acao epica" {
		t.Fatalf("title_norm = %q", got)
	}

	if got := scalar(
		t,
		db,
		`SELECT (search_tsv @@ websearch_to_tsquery('hs_pt','acao') AND search_tsv @@ websearch_to_tsquery('hs_pt','aventura'))::text FROM item_translations`,
	); got != "true" {
		t.Fatalf("site vector misses title or synopsis: %s", got)
	}

	if got := scalar(
		t,
		db,
		`SELECT (search_tsv_api @@ websearch_to_tsquery('hs_pt','aventura'))::text FROM item_translations`,
	); got != "false" {
		t.Fatalf("api vector holds a non-allowlisted synopsis: %s", got)
	}

	if got := scalar(
		t,
		db,
		`SELECT count(*) FROM search_dirty WHERE ref_id = '`+itemID+`' AND kind = 'item'`,
	); got != "1" {
		t.Fatalf("dirty rows for item: %s", got)
	}
}

const itemInsert = `INSERT INTO items (id, type, slug, wikidata_qid) VALUES (gen_random_uuid(), 'game', 'acao-epica', 'Q1') RETURNING id::text`
