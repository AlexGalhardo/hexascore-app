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

// TestOffersMigrationCopiesAffiliateLinks: the admin links become offers (RFC-003 section 2, expand step), the store
// follows the Amazon host, and running the data migration again copies nothing twice.
func TestOffersMigrationCopiesAffiliateLinks(t *testing.T) {
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

	exec := func(q string, args ...any) {
		t.Helper()

		if _, err := db.ExecContext(ctx, q, args...); err != nil {
			t.Fatalf("%v\n%s", err, q)
		}
	}

	// back out everything after the data migration and the migration itself, add links as an admin would have, and
	// run them again
	rollTo := func(id string) {
		t.Helper()

		for range 20 { // upper bound only: every newer migration is rolled back first
			step, err := m.Rollback(ctx, 1)
			if err != nil || len(step) != 1 {
				t.Fatalf("rollback: %v %v", step, err)
			}

			if step[0] == id {
				return
			}
		}

		t.Fatalf("%s was never rolled back", id)
	}

	rollTo("20261103010100_migrate_affiliate_links_to_offers")

	const item = "0197f3a0-0000-7000-8000-0000000000a1"

	exec(`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'game', 'alpha', 'Q1')`, item)
	exec(`INSERT INTO affiliate_links (item_id, store, url, active) VALUES
($1, 'steam', 'https://store.steampowered.com/app/1', true),
($1, 'gog', 'https://www.gog.com/en/game/x', false),
($1, 'mercadolivre', 'https://produto.mercadolivre.com.br/MLB-1', true)`, item)
	exec(
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ('0197f3a0-0000-7000-8000-0000000000a2', 'book', 'beta', 'Q2')`,
	)
	exec(`INSERT INTO affiliate_links (item_id, store, url) VALUES
('0197f3a0-0000-7000-8000-0000000000a2', 'amazon', 'https://www.amazon.com.br/dp/B000?tag=x-20')`)
	exec(
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ('0197f3a0-0000-7000-8000-0000000000a3', 'book', 'gamma', 'Q3')`,
	)
	exec(`INSERT INTO affiliate_links (item_id, store, url) VALUES
('0197f3a0-0000-7000-8000-0000000000a3', 'amazon', 'https://www.amazon.com/dp/B001?tag=y-20')`)
	exec(
		`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ('0197f3a0-0000-7000-8000-0000000000a4', 'book', 'delta', 'Q4')`,
	)
	exec(`INSERT INTO affiliate_links (item_id, store, url) VALUES
('0197f3a0-0000-7000-8000-0000000000a4', 'amazon', 'https://www.amazon.com.evil.example/dp/B002')`)

	if _, err := m.Up(ctx); err != nil {
		t.Fatal(err)
	}

	type row struct {
		store, kind, country string
		active               bool
	}

	want := map[string]row{
		"https://store.steampowered.com/app/1":        {"steam", "buy_digital", "ZZ", true},
		"https://www.gog.com/en/game/x":               {"gog", "buy_digital", "ZZ", false},
		"https://produto.mercadolivre.com.br/MLB-1":   {"mercadolivre", "buy", "BR", true},
		"https://www.amazon.com.br/dp/B000?tag=x-20":  {"amazon_br", "buy", "BR", true},
		"https://www.amazon.com/dp/B001?tag=y-20":     {"amazon_us", "buy", "US", true},
		"https://www.amazon.com.evil.example/dp/B002": {"amazon", "buy", "ZZ", true}, // a look-alike host is not Amazon
	}

	rows, err := db.QueryContext(ctx, `SELECT url, store, kind, country, active FROM offers WHERE source = 'admin'`)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = rows.Close() }()

	got := 0

	for rows.Next() {
		var (
			url string
			r   row
		)

		if err := rows.Scan(&url, &r.store, &r.kind, &r.country, &r.active); err != nil {
			t.Fatal(err)
		}

		got++

		if w, ok := want[url]; !ok || w != r {
			t.Fatalf("offer %s = %+v, want %+v", url, r, w)
		}
	}

	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	if got != len(want) {
		t.Fatalf("copied %d offers, want %d", got, len(want))
	}

	// the key is unique even without an edition: the same offer twice is refused
	if _, err := db.ExecContext(ctx, `INSERT INTO offers (item_id, country, store, kind, url, source, licence)
VALUES ($1, 'ZZ', 'steam', 'buy_digital', 'https://store.steampowered.com/app/2', 'admin', 'proprietary')`, item); err == nil {
		t.Fatal("a second steam offer with the same key was accepted")
	}

	// round trip: roll the offers tables back and forward
	rollTo("20261103010000_create_offers")

	if _, err := m.Up(ctx); err != nil {
		t.Fatal(err)
	}
}
