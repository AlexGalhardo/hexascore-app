//go:build integration

package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/controllers"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestCoversRoute_BySlug: covers are addressed /covers/{section}/{slug}[-t].webp whatever key the row holds, and
// old content-hashed item URLs answer 301 there (owner decision 2026-11-01, item 18).
func TestCoversRoute_BySlug(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	disk := storage.NewMemory("")
	put := func(key, body string) {
		t.Helper()

		if err := storage.PutBytes(t.Context(), disk, key, []byte(body)); err != nil {
			t.Fatal(err)
		}
	}

	exec := func(q string, args ...any) {
		t.Helper()

		if _, err := db.ExecContext(t.Context(), q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	zelda, x, xt := uuid.NewV7().String(), uuid.NewV7().String(), uuid.NewV7().String()
	old := func(id, suffix string) string { return "covers/" + id + "-0123456789abcdef" + suffix + ".webp" }

	for i, it := range []struct{ id, slug string }{{zelda, "zelda"}, {x, "x"}, {xt, "x-t"}} {
		put(old(it.id, ""), it.slug+" full")
		put(old(it.id, "-t"), it.slug+" thumb")
		exec(
			`INSERT INTO items (id, type, slug, wikidata_qid, cover_key, cover_thumb_key) VALUES ($1, 'game', $2, $3, $4, $5)`,
			it.id,
			it.slug,
			"Q77"+string(rune('0'+i)),
			old(it.id, ""),
			old(it.id, "-t"),
		)
	}

	edition := uuid.NewV7().String()
	put(old(edition, ""), "edition")

	mux := http.NewServeMux()
	mux.Handle("GET /covers/{key...}", controllers.CoversFrom(disk, db))

	get := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		return rec
	}

	for path, want := range map[string]string{
		"/covers/games/zelda.webp":   "zelda full",
		"/covers/games/zelda-t.webp": "zelda thumb",
		"/covers/games/x.webp":       "x full",
		// x-t's full wins the key; x's thumbnail keeps its old URL.
		"/covers/games/x-t.webp":   "x-t full",
		"/covers/games/x-t-t.webp": "x-t thumb",
		"/" + old(x, "-t"):         "x thumb",
		"/" + old(edition, ""):     "edition",
	} {
		rec := get(path)
		if rec.Code != http.StatusOK || rec.Body.String() != want {
			t.Errorf("%s = %d %q, want %q", path, rec.Code, rec.Body.String(), want)
		}
	}

	// A slug URL can change content (rekey, new cover): a day of cache, revalidated by ETag, even while the row
	// still holds a hashed key; content-hashed URLs stay immutable.
	if got := get("/covers/games/zelda.webp").Header().Get("Cache-Control"); got != "public, max-age=86400" {
		t.Errorf("hashed key behind a slug URL: Cache-Control = %q", got)
	}

	if got := get("/" + old(edition, "")).Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Errorf("edition Cache-Control = %q", got)
	}

	put("covers/games/zelda.webp", "zelda rekeyed")
	exec(`UPDATE items SET cover_key = 'covers/games/zelda.webp' WHERE id = $1`, zelda)

	rec := get("/covers/games/zelda.webp")
	if rec.Body.String() != "zelda rekeyed" || rec.Header().Get("Cache-Control") != "public, max-age=86400" {
		t.Errorf("slug key = %q, Cache-Control %q", rec.Body.String(), rec.Header().Get("Cache-Control"))
	}

	for path, target := range map[string]string{
		"/" + old(zelda, ""):   "/covers/games/zelda.webp",
		"/" + old(zelda, "-t"): "/covers/games/zelda-t.webp",
		"/" + old(xt, ""):      "/covers/games/x-t.webp",
	} {
		if rec := get(path); rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != target {
			t.Errorf("%s = %d to %q, want 301 to %s", path, rec.Code, rec.Header().Get("Location"), target)
		}
	}

	for _, path := range []string{"/covers/games/missing.webp", "/covers/books/zelda.webp", "/covers/games/Zelda.webp"} {
		if rec := get(path); rec.Code != http.StatusNotFound {
			t.Errorf("%s = %d, want 404", path, rec.Code)
		}
	}
}
