//go:build integration

package offers_test

import (
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/offers"
)

func TestPruneExpiresTMDBOffersAndOldClicks(t *testing.T) {
	r := newRig(t)
	game := r.item("movie", "public", `{}`)
	now := time.Date(2026, 11, 3, 3, 5, 0, 0, time.UTC)

	// TMDB data is cached for 180 days (ADR 0006): past expires_at it is deactivated, 30 days later deleted
	r.exec(
		`INSERT INTO stores (code, name, kind) VALUES ('tmdb_8', 'Netflix', 'streaming'), ('tmdb_9', 'Prime', 'streaming'),
('tmdb_10', 'Max', 'streaming')`,
	)

	insert := func(store string, expires time.Time, active bool) {
		r.exec(`INSERT INTO offers (item_id, country, store, kind, url, source, licence, expires_at, active)
VALUES ($1, 'BR', $2, 'stream', 'https://www.justwatch.com/br/x', 'tmdb_providers', 'tmdb', $3, $4)`, game, store, expires, active)
	}

	insert("tmdb_8", now.Add(-time.Hour), true)     // just expired: deactivated, kept
	insert("tmdb_9", now.AddDate(0, 0, -40), false) // expired long ago: deleted
	insert("tmdb_10", now.AddDate(0, 0, 100), true) // still fresh

	// admin and wikidata offers double as tombstones: never deleted, whatever their age or state
	r.offer(game, "steam", "buy_digital", "https://store.steampowered.com/app/1", "wikidata")
	r.exec(`UPDATE offers SET active = false, first_seen = '2020-01-01' WHERE store = 'steam'`)

	var click string
	if err := r.db.QueryRowContext(r.ctx, `SELECT id::text FROM offers WHERE store = 'tmdb_10'`).
		Scan(&click); err != nil {
		t.Fatal(err)
	}

	r.exec(
		`INSERT INTO offer_clicks (day, offer_id, country, clicks) VALUES ('2025-09-01', $1, 'BR', 4), ('2026-10-01', $1, 'BR', 2)`,
		click,
	)

	got, err := offers.Prune(r.ctx, r.db, now)
	if err != nil {
		t.Fatal(err)
	}

	if want := (offers.Pruned{Deactivated: 1, Deleted: 1, Clicks: 1}); got != want {
		t.Fatalf("pruned = %+v, want %+v", got, want)
	}

	count := func(q string) (n int) {
		t.Helper()

		if err := r.db.QueryRowContext(r.ctx, q).Scan(&n); err != nil {
			t.Fatal(err)
		}

		return n
	}

	if n := count(`SELECT count(*) FROM offers WHERE store = 'tmdb_8' AND NOT active`); n != 1 {
		t.Fatalf("the offer that just expired must stay, deactivated: %d", n)
	}

	if n := count(`SELECT count(*) FROM offers WHERE store IN ('tmdb_9')`); n != 0 {
		t.Fatal("the offer expired 40 days ago must be deleted")
	}

	if n := count(`SELECT count(*) FROM offers WHERE store = 'steam'`); n != 1 {
		t.Fatal("a wikidata tombstone must never be deleted")
	}

	if n := count(`SELECT count(*) FROM offer_clicks`); n != 1 {
		t.Fatalf("clicks left = %d, want the recent one", n)
	}

	again, err := offers.Prune(r.ctx, r.db, now)
	if err != nil || again != (offers.Pruned{}) {
		t.Fatalf("a second prune must find nothing: %+v, %v", again, err)
	}
}
