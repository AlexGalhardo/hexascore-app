//go:build integration

package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

func searchApp(t *testing.T) (*vukemtest.App, func(q string, args ...any)) {
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

	exec := func(q string, args ...any) {
		t.Helper()

		if _, err := db.ExecContext(t.Context(), q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	exec(`INSERT INTO data_sources (code, name, licence, url) VALUES ('hexascore', 'x', 'proprietary', 'https://x')`)

	for i, title := range []string{"Zelda", "Zelda II", "Zero Escape"} {
		id := uuid.NewV7().String()
		exec(`INSERT INTO items (id, type, slug, wikidata_qid, popularity_prior) VALUES ($1, 'game', $2, $3, $4)`,
			id, fmt.Sprintf("game-%d", i), fmt.Sprintf("Q%d", 5000+i), 10-i)
		exec(
			`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', $2, 'hexascore')`,
			id,
			title,
		)
	}

	if err := search.DrainSuggest(t.Context(), db); err != nil {
		t.Fatal(err)
	}

	return app, exec
}

func TestSearchSuggest(t *testing.T) {
	app, _ := searchApp(t)

	res := app.Client(t).Get("/search/suggest?q=zel").AssertOK().AssertJSONCount("data", 2)
	res.AssertJSONPath("data.0.title", "Zelda").AssertJSONPath("data.0.path", "/games/game-0")
	etag := res.Header("ETag")

	if etag == "" {
		t.Fatal("suggest has no ETag")
	}

	app.Client(t).WithHeader("If-None-Match", etag).Get("/search/suggest?q=zel").AssertStatus(304)

	app.Client(t).
		Get("/pt-br/search/suggest?q=zel&locale=").
		AssertOK().
		AssertJSONPath("data.0.path", "/pt-br/games/game-0")
	app.Client(t).Get("/search/suggest?q=z").AssertOK().AssertJSONCount("data", 0)
	app.Client(t).
		Get("/api/search/suggest?q=zel&type=game").
		AssertStatus(308).
		AssertRedirect("/search/suggest?q=zel&type=game")
}

func TestSearchPage(t *testing.T) {
	app, _ := searchApp(t)

	// HTML never carries an ETag (09-2); anonymous Inertia visits do.
	app.Client(t).Get("/search?q=zelda").AssertOK().AssertInertia("Search/Index").
		PropEquals("query.q", "zelda").HasProp("searchId")
	app.Client(t).Get("/search?q=zelda").AssertHeaderMissing("ETag")

	visit := app.Client(t).Inertia("")
	etag := visit.Get("/search?q=zelda").AssertOK().Header("ETag")

	if etag == "" {
		t.Fatal("Inertia search visit has no ETag")
	}

	app.Client(t).Inertia("").WithHeader("If-None-Match", etag).Get("/search?q=zelda").AssertStatus(304)

	app.Client(t).Get("/search?q=a").AssertOK().AssertInertia("Search/Index").PropEquals("tooShort", true)
	app.Client(t).Get("/search?q=zelda&type=bogus").AssertStatus(400)
}

func TestSearchEventsClickAndPurge(t *testing.T) {
	app, exec := searchApp(t)
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	count := func(where string, args ...any) (n int) {
		t.Helper()

		if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM search_events WHERE `+where, args...).
			Scan(&n); err != nil {
			t.Fatal(err)
		}

		return n
	}

	app.Client(t).Get("/search?q=zelda").AssertOK()

	if n := count(`surface = 'web' AND query_norm = 'zelda' AND clicked_ref IS NULL`); n != 1 {
		t.Fatalf("performed events = %d", n)
	}

	var searchID, ref string
	if err := db.QueryRowContext(t.Context(), `SELECT search_id::text, result_refs[1]::text FROM search_events`).
		Scan(&searchID, &ref); err != nil {
		t.Fatal(err)
	}

	click := map[string]any{"searchId": searchID, "refId": ref, "position": 1, "group": "game"}
	app.Client(t).PostJSON("/search/events", click).AssertNoContent()

	if n := count(`clicked_ref = $1::uuid AND click_position = 1`, ref); n != 1 {
		t.Fatalf("click rows = %d", n)
	}

	// TestSecurity_A04: a repeated click and a ref the search never returned
	// store nothing, so click analytics cannot be inflated.
	app.Client(t).PostJSON("/search/events", click).AssertNoContent()
	forged := map[string]any{"searchId": searchID, "refId": uuid.NewV7().String(), "position": 1, "group": "game"}
	app.Client(t).PostJSON("/search/events", forged).AssertNoContent()

	if n := count(`clicked_ref IS NOT NULL`); n != 1 {
		t.Fatalf("click rows after repeat and forged ref = %d", n)
	}

	// Unknown search ids store nothing; malformed bodies are rejected.
	click["searchId"] = uuid.NewV7().String()
	app.Client(t).PostJSON("/search/events", click).AssertNoContent()
	app.Client(t).PostJSON("/search/events", map[string]any{"searchId": "x"}).AssertUnprocessable()

	if n := count(`clicked_ref IS NOT NULL`); n != 1 {
		t.Fatalf("click rows after unknown id = %d", n)
	}

	app.Client(t).Get("/search?q=mail+ana@example.com").AssertOK()

	if n := count(`query_norm LIKE '%example%' OR query_norm LIKE '%ana@%'`); n != 0 {
		t.Fatalf("unredacted query stored: %d rows", n)
	}

	exec(`INSERT INTO search_events (at, search_id, surface, query_norm) VALUES ($1, $2::uuid, 'web', 'old')`,
		time.Now().Add(-31*24*time.Hour), uuid.NewV7().String())

	for _, cmd := range app.Commands() {
		if cmd.Name() == "search:purge" {
			if err := cmd.Run(t.Context(), app.App, nil); err != nil {
				t.Fatal(err)
			}
		}
	}

	if n := count(`query_norm = 'old'`); n != 0 {
		t.Fatalf("old event survived search:purge: %d", n)
	}

	if n := count(`query_norm = 'zelda'`); n == 0 {
		t.Fatal("recent events were purged")
	}
}
