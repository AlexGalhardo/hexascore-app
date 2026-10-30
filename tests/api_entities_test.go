//go:build integration

package tests

import (
	"net/http"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/uuid"
)

func TestAPIEntities(t *testing.T) {
	r := newAPICatalog(t)

	var studio, genre string
	for slug, dst := range map[string]*string{"team-cherry": &studio, "metroidvania": &genre} {
		if err := r.acc.DB.QueryRowContext(t.Context(), `SELECT id::text FROM entities WHERE slug = $1`, slug).
			Scan(dst); err != nil {
			t.Fatal(err)
		}
	}
	person := uuid.NewV7().String()
	for _, q := range [][]any{
		{`INSERT INTO entities (id, kind, slug, default_name, born_on, founded_year, external_ids, provenance)
VALUES ($1, 'person', 'ari-gibson', 'Ari Gibson', '1980-01-02', 1999, '{"tmdb":"55","wikidata":"Q1"}',
  '{"default_name":"wikidata","born_on":"tmdb","founded_year":"wikidata"}')`, person},
		{`INSERT INTO entity_translations (entity_id, locale, name, name_source, description, description_source,
  description_licence) VALUES ($1, 'en', 'Ari Gibson', 'wikidata', 'TMDB bio', 'tmdb', 'proprietary')`, person},
	} {
		if _, err := r.acc.DB.ExecContext(t.Context(), q[0].(string), q[1:]...); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("licence filter and kind", func(t *testing.T) {
		res := r.client.GetJSON("/api/v1/people/" + person).AssertOK().AssertMatchesOpenAPI()
		for _, leak := range []string{"tmdb", "TMDB bio", "1980", `"55"`} {
			if strings.Contains(res.Body(), leak) {
				t.Errorf("entity body leaks %q: %s", leak, res.Body())
			}
		}
		res.AssertJSONPath("name", "Ari Gibson").
			AssertJSONPath("foundedYear", 1999).
			AssertJSONPath("externalIds.wikidata", "Q1").
			AssertJSONPath("provenance.name.source", "wikidata").
			AssertJSONPath("meta.omittedFields", []string{"bornOn", "description", "externalIds"})

		r.client.GetJSON("/api/v1/entities/" + person).AssertOK().AssertMatchesOpenAPI()
		// default_name without a provenance source is hidden (fail closed, ADR 0006).
		r.client.GetJSON("/api/v1/organizations/"+studio).AssertOK().AssertMatchesOpenAPI().
			AssertJSONPath("meta.omittedFields", []string{"name"})
		for _, path := range []string{
			"/api/v1/organizations/" + person, "/api/v1/people/" + studio, "/api/v1/franchises/" + studio,
			"/api/v1/platforms/" + genre, "/api/v1/people/not-a-uuid", "/api/v1/entities/" + uuid.NewV7().String(),
			"/api/v1/entities/genres",
		} {
			r.client.GetJSON(path).AssertNotFound()
		}
		r.client.GetJSON("/api/v1/people/" + person + "?locale=fr").AssertStatus(http.StatusBadRequest)
	})

	t.Run("etag 304", func(t *testing.T) {
		path := "/api/v1/entities/" + person
		first := r.client.GetJSON(path).AssertOK()
		etag := first.Header("ETag")
		again := r.app.Client(t).WithToken(r.key).WithHeader("If-None-Match", etag).GetJSON(path).
			AssertStatus(http.StatusNotModified)
		if got, want := remaining(t, again), remaining(t, first)-1; got != want {
			t.Fatalf("X-Quota-Remaining after 304 = %d, want %d", got, want)
		}
	})

	t.Run("entity items", func(*testing.T) {
		base := "/api/v1/entities/" + studio + "/items"
		r.client.GetJSON(base).AssertOK().AssertMatchesOpenAPI().AssertJSONPath("data.0.id", r.game)
		r.client.GetJSON(base+"?type=games&role=developer").AssertOK().AssertMatchesOpenAPI().
			AssertJSONPath("data.0.id", r.game)
		r.client.GetJSON(base+"?type=movies").AssertOK().AssertJSONPath("data", []string{})
		r.client.GetJSON(base+"?role=director").AssertOK().AssertJSONPath("data", []string{})
		// The game-genre relation is IGDB-sourced: filtering by it must not reveal it (SDD 01 §6.1).
		r.client.GetJSON("/api/v1/entities/"+genre+"/items").AssertOK().AssertJSONPath("data", []string{})
		r.client.GetJSON("/api/v1/games?genre=metroidvania").AssertOK().AssertJSONPath("data", []string{})
		for _, q := range []string{"?type=films", "?role=bogus", "?bogus=1", "?limit=11"} {
			r.client.GetJSON(base + q).AssertStatus(http.StatusBadRequest)
		}
		r.client.GetJSON("/api/v1/entities/" + uuid.NewV7().String() + "/items").AssertNotFound()
	})
}
