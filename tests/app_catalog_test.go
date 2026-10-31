//go:build integration

package tests

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// appCatalogClient is an hx_app_ client for the licence fixture's reader.
func appCatalogClient(t *testing.T, r *apiCatalog, abilities ...string) *vukemtest.Client {
	t.Helper()

	u, err := r.acc.Users.ByCredentials(t.Context(), map[string]string{"email": "reader@example.com"})
	if err != nil {
		t.Fatal(err)
	}

	plain, _, err := r.acc.Tokens.Create(t.Context(), u.AuthID(), "phone", abilities, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	return r.app.Client(t).WithToken(plain)
}

// TestAppCatalog covers the site-audience reads on /app/v1. Trending is not a route.
func TestAppCatalog(t *testing.T) {
	r := newAPICatalog(t)
	client := appCatalogClient(t, r, "catalog:read")
	item := "/app/v1/games/" + r.game

	body := client.GetJSON(item).AssertOK().
		AssertJSONPath("title", "Hollow Knight").
		AssertJSONPath("synopsis.text", "TMDB overview text").
		AssertJSONPath("synopsis.licence", "proprietary").
		AssertJSONPath("runtimeMinutes", 40).
		AssertJSONPath("externalIds.tmdb", "999").
		AssertJSONPath("externalIds.igdb", "1234").
		Body()
	for _, want := range []string{"Metroidvania", "IGDB Published"} {
		if !strings.Contains(body, want) {
			t.Errorf("/app/v1 item missing %q: %s", want, body)
		}
	}

	client.GetJSON("/app/v1/games").AssertOK().AssertJSONPath("data.0.id", r.game)
	client.GetJSON("/app/v1/games?limit=11").AssertOK()
	client.GetJSON("/app/v1/games?limit=101").AssertStatus(http.StatusBadRequest)
	client.GetJSON("/app/v1/games/genres").AssertOK().AssertJSONPath("data.0.slug", "metroidvania")
	client.GetJSON("/app/v1/books/"+r.book+"/editions").AssertOK().
		AssertJSONPath("data.0.title", "Dune").
		AssertJSONPath("data.0.pageCount", 412)

	var genre string
	if err := r.acc.DB.QueryRowContext(
		t.Context(), `SELECT id::text FROM entities WHERE slug = $1`, "metroidvania",
	).Scan(&genre); err != nil {
		t.Fatal(err)
	}

	client.GetJSON("/app/v1/entities/"+genre).AssertOK().
		AssertJSONPath("name", "Metroidvania").
		AssertJSONPath("kind", "genre")
	client.GetJSON("/app/v1/entities/"+genre+"/items?type=games&role=genre").AssertOK().
		AssertJSONPath("data.0.id", r.game)

	// Schemathesis sends no Accept. An unknown collection is still problem+json.
	r.app.Client(t).Get("/app/v1/0").AssertStatus(http.StatusNotFound).
		AssertHeaderContains("Content-Type", "application/problem+json").
		AssertJSONPath("code", "not-found")
	r.app.Client(t).Get("/app/v1/0/0").AssertStatus(http.StatusNotFound).
		AssertHeaderContains("Content-Type", "application/problem+json")

	r.app.Client(t).GetJSON(item).AssertUnauthorized()
	r.client.GetJSON(item).AssertUnauthorized()
	appCatalogClient(t, r, "ratings:write").GetJSON(item).AssertStatus(http.StatusForbidden)
	client.GetJSON("/app/v1/games/trending").AssertOK().AssertJSONCount("data", 0)

	gone := uuid.NewV7().String()
	if err := catalog.RecordMerge(t.Context(), r.acc.DB, catalog.ScopeItem, "game",
		catalog.Target{ID: gone, Slug: "hollow-knight-dup"}, r.game); err != nil {
		t.Fatal(err)
	}
	client.GetJSON("/app/v1/games/"+gone).AssertStatus(http.StatusMovedPermanently).
		AssertHeader("Location", item)
}
