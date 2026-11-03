//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/schema"

	"github.com/AlexGalhardo/hexascore-app/database/migrations"
)

// RFC-001 CW-06: a second item with the same TMDB id (of the same media type) cannot be inserted; the same number on
// another media type can.
func TestExternalIDsAreUnique(t *testing.T) {
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

	insert := func(slug, qid, typ, ext string) error {
		_, err := db.ExecContext(
			ctx,
			`INSERT INTO items (id, type, slug, wikidata_qid, external_ids) VALUES (gen_random_uuid(), $1, $2, $3, $4::jsonb)`,
			typ,
			slug,
			qid,
			ext,
		)

		return err
	}

	if err := insert("a", "Q1", "movie", `{"tmdb":"603","imdb":"tt1"}`); err != nil {
		t.Fatal(err)
	}

	if err := insert("b", "Q2", "movie", `{"tmdb":"603"}`); err == nil {
		t.Fatal("a second film with TMDB 603 was accepted")
	}

	if err := insert("c", "Q3", "series", `{"tmdb":"603"}`); err != nil {
		t.Fatalf("a series may share the TMDB number of a film: %v", err)
	}

	if err := insert("d", "Q4", "game", `{"imdb":"tt1"}`); err == nil {
		t.Fatal("a second item with the same imdb id was accepted")
	}

	if err := insert("e", "Q5", "book", `{"imdb":""}`); err != nil {
		t.Fatalf("an empty id is no id: %v", err)
	}

	if err := insert("f", "Q6", "book", `{"imdb":""}`); err != nil {
		t.Fatalf("two empty ids must not collide: %v", err)
	}
}

// Data that already breaks the rule must not break the deploy: the migration notes it and leaves that index out.
func TestExternalIDIndexIsSkippedWhenDuplicatesExist(t *testing.T) {
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

	for range 20 { // back out the conflicts migration
		step, err := m.Rollback(ctx, 1)
		if err != nil || len(step) != 1 {
			t.Fatalf("rollback: %v %v", step, err)
		}

		if step[0] == "20261103020000_create_item_conflicts" {
			break
		}
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO items (id, type, slug, wikidata_qid, external_ids) VALUES
(gen_random_uuid(), 'movie', 'a', 'Q1', '{"imdb":"tt9","steam":"1"}'), (gen_random_uuid(), 'movie', 'b', 'Q2', '{"imdb":"tt9"}')`); err != nil {
		t.Fatal(err)
	}

	if _, err := m.Up(ctx); err != nil {
		t.Fatalf("the migration failed on existing duplicates: %v", err)
	}

	exists := func(index string) bool {
		var n int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM pg_indexes WHERE indexname = $1`, index).
			Scan(&n); err != nil {
			t.Fatal(err)
		}

		return n == 1
	}

	if exists("items_ext_imdb_uidx") {
		t.Fatal("the imdb index was created over duplicates")
	}

	if !exists("items_ext_steam_uidx") || !exists("items_ext_tmdb_uidx") {
		t.Fatal("the indexes without duplicates must still be created")
	}
}
