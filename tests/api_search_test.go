//go:build integration

package tests

import (
	"net/http"
	"strings"
	"testing"
)

func TestAPISearch(t *testing.T) {
	r := newAPICatalog(t)

	r.client.GetJSON("/api/v1/search?q=hollow").AssertOK().AssertMatchesOpenAPI().
		AssertJSONPath("data.0.id", r.game).
		AssertJSONPath("meta.page", 1)
	r.client.GetJSON("/api/v1/search?q=hollow&type=books").AssertOK().AssertMatchesOpenAPI().
		AssertJSONPath("data", []string{})
	// The TMDB synopsis is not allowlisted text, so it neither ranks nor appears.
	res := r.client.GetJSON("/api/v1/search?q=overview").
		AssertOK().
		AssertMatchesOpenAPI().
		AssertJSONPath("data", []string{})
	if strings.Contains(res.Body(), "TMDB") {
		t.Fatalf("search leaks a restricted field: %s", res.Body())
	}

	if _, err := r.acc.DB.ExecContext(
		t.Context(),
		`INSERT INTO entity_translations (entity_id, locale, name, name_source)
SELECT id, 'en', 'Team Cherry', 'wikidata' FROM entities WHERE slug = 'team-cherry'`,
	); err != nil {
		t.Fatal(err)
	}
	r.client.GetJSON("/api/v1/search/entities?q=cherry").AssertOK().AssertMatchesOpenAPI().
		AssertJSONPath("data.0.kind", "organization")
	r.client.GetJSON("/api/v1/search/entities?q=cherry&kind=person").AssertOK().AssertMatchesOpenAPI().
		AssertJSONPath("data", []string{})

	for _, q := range []string{
		"", "?q=", "?q=a&bogus=1", "?q=zelda&page=11", "?q=zelda&limit=51", "?q=zelda&type=films", "?q=zelda&page=x",
	} {
		r.client.GetJSON("/api/v1/search" + q).AssertStatus(http.StatusBadRequest)
	}
	r.client.GetJSON("/api/v1/search/entities?q=cherry&kind=genre").AssertStatus(http.StatusBadRequest)
}

func TestAPILookup(t *testing.T) {
	r := newAPICatalog(t)

	// FR-05-16: the TMDB and IGDB ids are accepted as input and never echoed.
	for _, q := range []string{"source=tmdb&id=999", "source=igdb&id=1234", "source=wikidata&id=Q20984815"} {
		res := r.client.GetJSON("/api/v1/lookup?"+q).AssertOK().AssertMatchesOpenAPI().AssertJSONPath("id", r.game)
		for _, leak := range []string{"tmdb", "igdb", "999", "1234"} {
			if strings.Contains(res.Body(), leak) {
				t.Errorf("lookup %s leaks %q: %s", q, leak, res.Body())
			}
		}
		res.AssertJSONPath("externalIds.wikidata", "Q20984815")
	}
	r.client.GetJSON("/api/v1/lookup?source=isbn13&id=9780441013593").AssertOK().AssertMatchesOpenAPI().
		AssertJSONPath("id", r.book)

	r.client.GetJSON("/api/v1/lookup?source=tmdb&id=1").AssertNotFound()
	r.client.GetJSON("/api/v1/lookup?source=isbn13&id=9780441013593x").AssertNotFound()

	// A hidden item sharing the id and sorting first must not shadow the public one (FR-01-14).
	if _, err := r.acc.DB.ExecContext(t.Context(), `INSERT INTO items (id, type, slug, wikidata_qid, external_ids,
  provenance, visibility) VALUES ('00000000-0000-7000-8000-000000000001', 'game', 'hidden-game', 'Q42',
  '{"igdb":1234}', '{}', 'hidden')`); err != nil {
		t.Fatal(err)
	}
	r.client.GetJSON("/api/v1/lookup?source=igdb&id=1234").AssertOK().AssertJSONPath("id", r.game)
	r.client.GetJSON("/api/v1/lookup?source=wikidata&id=Q42").AssertNotFound()
	for _, q := range []string{"", "?source=tmdb", "?source=bogus&id=1", "?source=tmdb&id=", "?source=tmdb&id=1&x=1"} {
		r.client.GetJSON("/api/v1/lookup" + q).AssertStatus(http.StatusBadRequest)
	}
}
