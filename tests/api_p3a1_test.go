//go:build integration

package tests

import (
	"net/http"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

func TestAPIEntityItems(t *testing.T) {
	r := newAPICatalog(t)
	var studio string
	if err := r.acc.DB.QueryRowContext(t.Context(),
		`SELECT id::text FROM entities WHERE slug = 'team-cherry'`).Scan(&studio); err != nil {
		t.Fatal(err)
	}

	base := "/api/v1/entities/" + studio + "/items"
	r.client.GetJSON(base).AssertOK().AssertMatchesOpenAPI().AssertJSONPath("data.0.id", r.game)
	r.client.GetJSON(base+"?type=games&role=developer").AssertOK().AssertMatchesOpenAPI().
		AssertJSONPath("data.0.id", r.game)
	r.client.GetJSON(base+"?type=movies").AssertOK().AssertJSONPath("data", []string{})
	r.client.GetJSON("/api/v1/entities/" + uuid.NewV7().String() + "/items").AssertNotFound()
}

func TestAPICursorTamper(t *testing.T) {
	r := newAPICatalog(t)
	var studio string
	if err := r.acc.DB.QueryRowContext(t.Context(),
		`SELECT id::text FROM entities WHERE slug = 'team-cherry'`).Scan(&studio); err != nil {
		t.Fatal(err)
	}

	base := "/api/v1/entities/" + studio + "/items"
	r.client.GetJSON(base+"?cursor=garbage").AssertStatus(http.StatusBadRequest).
		AssertJSONPath("code", "invalid-cursor")
	r.client.GetJSON(base+"?limit=11").AssertStatus(http.StatusBadRequest).
		AssertJSONPath("detail", "invalid-query: limit exceeds the free plan page size 10")
}

func TestAPIEditions(t *testing.T) {
	r := newAPICatalog(t)

	r.client.GetJSON("/api/v1/books/"+r.book+"/editions").AssertOK().AssertMatchesOpenAPI().
		AssertJSONPath("data.0.isbn13", "9780441013593")
	r.client.GetJSON("/api/v1/books/" + r.game + "/editions").AssertNotFound()

	gone := uuid.NewV7().String()
	if err := catalog.RecordMerge(t.Context(), r.acc.DB, catalog.ScopeItem, "book",
		catalog.Target{ID: gone, Slug: "dune-dup"}, r.book); err != nil {
		t.Fatal(err)
	}
	r.client.GetJSON("/api/v1/books/"+gone+"/editions").AssertStatus(http.StatusMovedPermanently).
		AssertHeader("Location", "/api/v1/books/"+r.book+"/editions")
}

func TestAPILookupNoRestrictedEcho(t *testing.T) {
	r := newAPICatalog(t)

	for _, q := range []string{"source=tmdb&id=999", "source=igdb&id=1234"} {
		res := r.client.GetJSON("/api/v1/lookup?"+q).AssertOK().AssertJSONPath("id", r.game)
		body := res.Body()
		for _, leak := range []string{`"tmdb"`, `"igdb"`} {
			if strings.Contains(body, leak) {
				t.Fatalf("lookup %s echoes %s: %s", q, leak, body)
			}
		}
	}
}
