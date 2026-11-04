//go:build integration

package tests

import (
	"net/http"
	"testing"

	"github.com/AlexGalhardo/vukem/uuid"
)

// Owner 2026-11-04: a watch provider links to the service itself, never to TMDB's page about the title, and a service
// with no known site is left out of the page and of the redirect.
func TestWatchProvidersLinkToTheServiceOrAreHidden(t *testing.T) {
	r := newListRig(t)
	id := r.item("game", "public")

	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	exec := func(q string, args ...any) {
		t.Helper()

		if _, err := db.ExecContext(t.Context(), q, args...); err != nil {
			t.Fatalf("%v\n%s", err, q)
		}
	}

	exec(`INSERT INTO stores (code, name, kind) VALUES ('tmdb_8', 'Netflix', 'streaming'),
  ('tmdb_999999', 'Some Local Cinema Club', 'streaming')`)

	netflix, unknown := uuid.NewV7().String(), uuid.NewV7().String()

	for offerID, store := range map[string]string{netflix: "tmdb_8", unknown: "tmdb_999999"} {
		exec(`INSERT INTO offers (id, item_id, country, store, kind, url, source, licence)
VALUES ($1, $2, 'BR', $3, 'stream', 'https://www.themoviedb.org/movie/603/watch?locale=BR', 'tmdb_providers', 'tmdb-justwatch')`,
			offerID, id, store)
	}

	r.app.Client(t).Inertia("").Get(r.itemPath(id)).AssertOK().AssertInertia("Catalog/Show").
		HasProp("offers.byCountry.BR", 1).
		PropEquals("offers.byCountry.BR.0.name", "Netflix")

	r.app.Client(t).Get("/go/" + netflix).AssertStatus(http.StatusFound).AssertRedirect("https://www.netflix.com")
	r.app.Client(t).Get("/go/" + unknown).AssertStatus(http.StatusNotFound)
}
