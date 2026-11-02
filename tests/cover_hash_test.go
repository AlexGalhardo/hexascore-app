//go:build integration

package tests

import (
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/covers"
)

// cover_hashes round trip: a hash with the top bit set survives the bigint column, a near copy matches and a far one
// does not; data_sources licences decide display rights, and an unknown source has none.
func TestCoverHashStoreAndRights(t *testing.T) {
	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), `INSERT INTO data_sources (code, name, licence, url) VALUES
 ('steam','Steam','STEAM-WEB-API-TERMS','https://s'), ('olpending','OL','pending','https://o')`); err != nil {
		t.Fatal(err)
	}

	const dhash = uint64(0xF0F0F0F0F0F0F0F1)

	h := covers.PGHashes{DB: db}
	if err := h.Put(t.Context(), covers.HashRow{
		Key: "covers/games/zelda.webp", DHash: dhash, Width: 600, Height: 600,
		Bytes: 10, Source: "steam", Licence: "STEAM-WEB-API-TERMS",
	}); err != nil {
		t.Fatal(err)
	}

	for hash, want := range map[uint64]bool{dhash: true, dhash ^ 0b1011: true, ^dhash: false} {
		if got, err := h.Same(t.Context(), "covers/games/zelda.webp", hash); err != nil || got != want {
			t.Errorf("Same(%x) = %v, %v, want %v", hash, got, err, want)
		}
	}

	if got, err := h.Same(t.Context(), "covers/games/other.webp", dhash); err != nil || got {
		t.Errorf("unknown key matched: %v, %v", got, err)
	}

	r := covers.PGRights{DB: db}
	for source, want := range map[string]bool{"steam": true, "olpending": false, "nosuch": false} {
		if _, ok, err := r.Allowed(t.Context(), source); err != nil || ok != want {
			t.Errorf("Allowed(%s) = %v, %v, want %v", source, ok, err, want)
		}
	}
}
