//go:build integration

package tests

import (
	"fmt"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// TestCoversBaseJSONAbsolute: with COVERS_BASE_URL set, the /app/v1 JSON (list items, search hits, library,
// public profile ratings) and the
// suggest JSON carry absolute cover URLs, while the Inertia props keep app paths and share the trimmed base.
func TestCoversBaseJSONAbsolute(t *testing.T) {
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL": pgtest.Database(t, ""), "COVERS_BASE_URL": " https://bucket.example/ ",
	}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	exec := func(q string, args ...any) {
		t.Helper()

		if _, err := db.ExecContext(t.Context(), q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	exec(`INSERT INTO data_sources (code, name, licence, url) VALUES ('hexascore', 'x', 'proprietary', 'https://x')`)

	itemID := uuid.NewV7().String()
	exec(`INSERT INTO items (id, type, slug, wikidata_qid, popularity_prior, cover_key, cover_thumb_key)
VALUES ($1, 'game', 'game-0', 'Q5000', 10, 'covers/a-1.webp', 'covers/a-1-t.webp')`, itemID)
	exec(`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', 'Zelda', 'hexascore')`, itemID)

	if err := search.DrainSuggest(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	r := &listRig{t: t, app: app, acc: acc}
	u, owner := r.user("covers-base@example.com", "lists:write", "catalog:read")
	r.setNickname(u, "CoverFan", false)
	r.status(u.ID, itemID, "want")
	exec(`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6) VALUES ($1, $2, 50, 50, 50, 50, 50, 50)`,
		u.ID, itemID)
	id, _ := owner.PostJSON("/app/v1/lists", listBody("game", "Covers")).AssertStatus(201).JSONPath("id")
	path := fmt.Sprintf("/app/v1/lists/%v", id)
	owner.PutJSON(path+"/items/"+itemID, map[string]any{}).AssertOK()

	const thumb, full = "https://bucket.example/covers/a-1-t.webp", "https://bucket.example/covers/a-1.webp"

	owner.GetJSON(path).AssertOK().
		AssertJSONPath("items.0.item.cover.thumb", thumb).
		AssertJSONPath("items.0.item.cover.full", full)
	owner.GetJSON("/app/v1/search?q=zelda").AssertOK().
		AssertJSONPath("data.0.cover.thumb", thumb).
		AssertJSONPath("data.0.cover.full", full)
	owner.GetJSON("/app/v1/me/library").AssertOK().
		AssertJSONPath("groups.0.items.0.cover.thumb", thumb)
	owner.GetJSON("/app/v1/users/coverfan").AssertOK().
		AssertJSONPath("ratings.0.cover.thumb", thumb)
	app.Client(t).Get("/search/suggest?q=zel").AssertOK().AssertJSONPath("data.0.coverThumbUrl", thumb)
	app.Client(t).Get("/app/v1/search/suggest?q=zel").AssertOK().AssertJSONPath("data.0.coverThumbUrl", thumb)

	app.Client(t).Get("/search?q=zelda").AssertOK().AssertInertia("Search/Index").
		PropEquals("coversBase", "https://bucket.example").
		PropEquals("groups.0.items.0.cover.thumb", "/covers/a-1-t.webp")
}
