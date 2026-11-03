//go:build integration

package offers_test

import (
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/offers"
)

func (r *rig) save(item string, now time.Time, ps ...offers.Provider) {
	r.t.Helper()

	err := r.db.Tx(r.ctx, func(tx *database.Tx) error {
		return offers.SaveProviders(r.ctx, tx, item, now, ps)
	})
	if err != nil {
		r.t.Fatal(err)
	}
}

func TestSaveProviders(t *testing.T) {
	r := newRig(t)
	film := r.item("movie", "public", `{}`)
	now := time.Date(2026, 11, 3, 1, 0, 0, 0, time.UTC)

	const link = "https://www.themoviedb.org/movie/603-the-matrix/watch?locale=BR"

	netflix := offers.Provider{Country: "BR", Kind: "stream", ID: 8, Name: "Netflix", Link: link}
	apple := offers.Provider{Country: "BR", Kind: "rent", ID: 2, Name: "Apple TV", Link: link}

	r.save(film, now, netflix, apple)

	var n int
	if err := r.db.QueryRowContext(r.ctx, `SELECT count(*) FROM offers WHERE item_id = $1 AND active
AND source = 'tmdb_providers' AND expires_at = $2 AND last_verified_at = $3`, film, now.AddDate(0, 0, 180), now).Scan(&n); err != nil || n != 2 {
		t.Fatalf("active tmdb offers with a 180-day expiry = %d, %v", n, err)
	}

	var name, attribution, kind string
	if err := r.db.QueryRowContext(r.ctx, `SELECT name, attribution, kind FROM stores WHERE code = 'tmdb_8'`).
		Scan(&name, &attribution, &kind); err != nil {
		t.Fatal(err)
	}

	if name != "Netflix" || kind != "streaming" || attribution == "" {
		t.Fatalf("store = %q %q %q", name, kind, attribution)
	}

	// a later refresh: Netflix left Brazil, Apple stays, a rename is picked up, the expiry moves
	later := now.AddDate(0, 0, 30)
	apple.Name = "Apple TV+"

	r.save(film, later, apple)

	states := map[string]bool{}

	rows, err := r.db.QueryContext(r.ctx, `SELECT store, active FROM offers WHERE item_id = $1`, film)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var s string

		var a bool
		if err := rows.Scan(&s, &a); err != nil {
			t.Fatal(err)
		}

		states[s] = a
	}

	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	if states["tmdb_8"] || !states["tmdb_2"] {
		t.Fatalf("states = %v, want Netflix deactivated and Apple active", states)
	}

	if err := r.db.QueryRowContext(r.ctx, `SELECT name FROM stores WHERE code = 'tmdb_2'`).
		Scan(&name); err != nil ||
		name != "Apple TV+" {
		t.Fatalf("store name = %q, %v", name, err)
	}

	// it is back: the offer comes up again
	r.save(film, later.AddDate(0, 0, 30), netflix, apple)

	if err := r.db.QueryRowContext(r.ctx, `SELECT count(*) FROM offers WHERE item_id = $1 AND active`, film).
		Scan(&n); err != nil ||
		n != 2 {
		t.Fatalf("active after Netflix returned = %d, %v", n, err)
	}

	// the item is marked checked, even when it has no provider anywhere
	r.save(film, later.AddDate(0, 0, 60))

	var checked time.Time
	if err := r.db.QueryRowContext(r.ctx, `SELECT checked_at FROM offer_source_checks WHERE item_id = $1 AND source = 'tmdb_providers'`, film).
		Scan(&checked); err != nil ||
		!checked.Equal(later.AddDate(0, 0, 60)) {
		t.Fatalf("checked_at = %v, %v", checked, err)
	}

	if err := r.db.QueryRowContext(r.ctx, `SELECT count(*) FROM offers WHERE item_id = $1 AND active`, film).
		Scan(&n); err != nil ||
		n != 0 {
		t.Fatalf("an empty answer must deactivate what was listed before: %d, %v", n, err)
	}
}

func TestSaveProvidersLeavesAdminOffersAlone(t *testing.T) {
	r := newRig(t)
	film := r.item("movie", "public", `{}`)
	now := time.Date(2026, 11, 3, 1, 0, 0, 0, time.UTC)

	r.exec(`INSERT INTO stores (code, name, kind) VALUES ('tmdb_8', 'Netflix', 'streaming')`)
	r.exec(`INSERT INTO offers (item_id, country, store, kind, url, source, licence)
VALUES ($1, 'BR', 'tmdb_8', 'stream', 'https://www.netflix.com/title/1', 'admin', 'proprietary')`, film)

	r.save(film, now, offers.Provider{
		Country: "BR", Kind: "stream", ID: 8, Name: "Netflix",
		Link: "https://www.themoviedb.org/movie/1/watch?locale=BR",
	})

	var url, source string
	if err := r.db.QueryRowContext(r.ctx, `SELECT url, source FROM offers WHERE item_id = $1`, film).
		Scan(&url, &source); err != nil {
		t.Fatal(err)
	}

	if url != "https://www.netflix.com/title/1" || source != "admin" {
		t.Fatalf("the admin offer was overwritten: %s %s", url, source)
	}
}
