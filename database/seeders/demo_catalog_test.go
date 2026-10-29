//go:build integration

package seeders_test

import (
	"context"
	"slices"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/database/seeders"
)

func seededDB(t *testing.T) *database.DB {
	t.Helper()

	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if err := (seeders.DemoCatalog{}).Seed(context.Background(), db); err != nil {
		t.Fatal(err)
	}

	return db
}

func count(t *testing.T, db *database.DB, query string) int {
	t.Helper()

	var n int
	if err := db.QueryRowContext(context.Background(), query).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

func slugs(t *testing.T, db *database.DB) []string {
	t.Helper()

	rows, err := db.QueryContext(context.Background(), `SELECT type || '/' || slug FROM items ORDER BY type, slug`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var out []string

	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}

		out = append(out, s)
	}

	return out
}

func TestDemoCatalog(t *testing.T) {
	db := seededDB(t)

	for query, want := range map[string]int{
		`SELECT count(*) FROM items`:                                                    300,
		`SELECT count(DISTINCT type) FROM items`:                                        4,
		`SELECT count(*) FROM item_translations`:                                        900,
		`SELECT count(*) FROM users`:                                                    51,
		`SELECT count(*) FROM users WHERE role = 'admin'`:                               1,
		`SELECT count(*) FROM editions`:                                                 75 + 25,
		`SELECT count(*) FROM items WHERE type = 'book' AND primary_edition_id IS NULL`: 0,
		`SELECT count(*) FROM entity_translations`:                                      3 * (6 + 10 + 3 + 40 + 15 + 10),
	} {
		if got := count(t, db, query); got != want {
			t.Errorf("%s = %d, want %d", query, got, want)
		}
	}

	if got := count(t, db, `SELECT count(*) FROM item_entities`); got < 300 {
		t.Errorf("item_entities = %d, want at least one relation per item", got)
	}

	// Same seed on a fresh database: identical slugs.
	if a, b := slugs(t, db), slugs(t, seededDB(t)); !slices.Equal(a, b) {
		t.Fatal("two runs with Seed(42) produced different slugs")
	}

	// A second run on the same database adds nothing.
	if err := (seeders.DemoCatalog{}).Seed(context.Background(), db); err != nil {
		t.Fatal(err)
	}

	if got := count(t, db, `SELECT count(*) FROM items`); got != 300 {
		t.Errorf("items after second run = %d, want 300", got)
	}
}
