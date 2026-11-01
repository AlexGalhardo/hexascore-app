//go:build integration

package tests

import (
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/covers"
)

// TestCoversRekey: covers:rekey moves item covers to their slug keys on a fake disk, dry run first, idempotent,
// keeping the old objects for covers.gc (owner decision 2026-11-01, item 18).
func TestCoversRekey(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	disk := storage.NewMemory("")
	put := func(key string) {
		t.Helper()

		if err := storage.PutBytes(t.Context(), disk, key, []byte(key)); err != nil {
			t.Fatal(err)
		}
	}

	exec := func(q string, args ...any) {
		t.Helper()

		if _, err := db.ExecContext(t.Context(), q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	old := func(id, suffix string) string { return "covers/" + id + "-0123456789abcdef" + suffix + ".webp" }
	ids := map[string]string{}

	// zelda: both renditions; x and x-t: x's thumbnail key is x-t's full key; dune: object missing; bare: no cover.
	for i, it := range []struct {
		slug, typ    string
		cover, store bool
	}{
		{"zelda", "game", true, true},
		{"x", "game", true, true},
		{"x-t", "game", true, true},
		{"dune", "book", true, false},
		{"bare", "movie", false, false},
	} {
		id := uuid.NewV7().String()
		ids[it.slug] = id

		var full, thumb any
		if it.cover {
			full, thumb = old(id, ""), old(id, "-t")
		}

		if it.store {
			put(old(id, ""))
			put(old(id, "-t"))
		}

		exec(
			`INSERT INTO items (id, type, slug, wikidata_qid, cover_key, cover_thumb_key) VALUES ($1, $2, $3, $4, $5, $6)`,
			id,
			it.typ,
			it.slug,
			"Q88"+string(rune('0'+i)),
			full,
			thumb,
		)
	}

	// fan: a community upload, never moved (its object is the user's, outside covers/).
	ids["fan"] = uuid.NewV7().String()
	put("community/covers/fan.webp")
	exec(`INSERT INTO items (id, type, slug, wikidata_qid, cover_key) VALUES ($1, 'series', 'fan', 'Q8899', $2)`,
		ids["fan"], "community/covers/fan.webp")

	keys := func(slug string) (string, string) {
		t.Helper()

		var full, thumb string
		if err := db.QueryRowContext(t.Context(),
			`SELECT coalesce(cover_key, ''), coalesce(cover_thumb_key, '') FROM items WHERE id = $1`, ids[slug]).
			Scan(&full, &thumb); err != nil {
			t.Fatal(err)
		}

		return full, thumb
	}

	dry, err := covers.Rekey(t.Context(), db, disk, true, 2)
	if err != nil {
		t.Fatal(err)
	}

	if dry != (covers.RekeyResult{Items: 5, Moved: 3, Current: 1, Missing: 1}) {
		t.Fatalf("dry run = %+v", dry)
	}

	if full, _ := keys("zelda"); full != old(ids["zelda"], "") {
		t.Fatalf("the dry run moved zelda to %q", full)
	}

	if ok, _ := disk.Exists(t.Context(), "covers/games/zelda.webp"); ok {
		t.Fatal("the dry run copied an object")
	}

	res, err := covers.Rekey(t.Context(), db, disk, false, 2)
	if err != nil {
		t.Fatal(err)
	}

	// x's thumbnail stays where it is: 5 objects for 3 moved rows.
	if res != (covers.RekeyResult{Items: 5, Moved: 3, Copied: 5, Current: 1, Missing: 1}) {
		t.Fatalf("run = %+v", res)
	}

	for slug, want := range map[string][2]string{
		"zelda": {"covers/games/zelda.webp", "covers/games/zelda-t.webp"},
		"x":     {"covers/games/x.webp", old(ids["x"], "-t")},
		"x-t":   {"covers/games/x-t.webp", "covers/games/x-t-t.webp"},
		"dune":  {old(ids["dune"], ""), old(ids["dune"], "-t")},
		"fan":   {"community/covers/fan.webp", ""},
	} {
		if full, thumb := keys(slug); full != want[0] || thumb != want[1] {
			t.Errorf("%s keys = %q, %q, want %q, %q", slug, full, thumb, want[0], want[1])
		}
	}

	if body, _ := storage.Get(t.Context(), disk, "covers/games/zelda-t.webp"); string(body) != old(ids["zelda"], "-t") {
		t.Errorf("zelda thumbnail copy = %q", body)
	}

	// The old objects stay for covers.gc, which the trigger has already scheduled.
	if ok, _ := disk.Exists(t.Context(), old(ids["zelda"], "")); !ok {
		t.Error("the old zelda object was deleted")
	}

	var released int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM cover_gc WHERE key LIKE '%-0123456789abcdef%'`).
		Scan(&released); err != nil || released != 5 {
		t.Errorf("released old keys = %d (%v), want 5", released, err)
	}

	again, err := covers.Rekey(t.Context(), db, disk, false, 500)
	if err != nil {
		t.Fatal(err)
	}

	if again != (covers.RekeyResult{Items: 5, Current: 4, Missing: 1}) {
		t.Fatalf("second run = %+v, want nothing to move", again)
	}

	if _, err := covers.Rekey(t.Context(), db, nil, true, 0); err == nil {
		t.Fatal("no disk must fail")
	}
}
