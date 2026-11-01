//go:build integration

package tests

import (
	"math"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/config"
)

// P3.A.2.2: qa/sql/api_coverage.sql must count a field exactly when the API serialiser exposes it, so the
// published coverage numbers (SDD 05 §6.2, D-33) cannot drift from what customers receive. The expected
// percentages are derived from catalog.LoadAPIItems over the same fixtures; a mismatch is a bug in one of them.
func TestAPICoverageSQL(t *testing.T) {
	t.Parallel()

	c := newAPICatalog(t)
	db := c.acc.DB
	movie, book2, hidden := uuid.NewV7().String(), uuid.NewV7().String(), uuid.NewV7().String()
	genre := uuid.NewV7().String()

	for _, q := range [][]any{
		{`INSERT INTO data_sources (code, name, licence, url) VALUES
  ('wikipedia', 'Wikipedia', 'CC-BY-SA-4.0', 'https://wikipedia.org'), ('community', 'Community', 'A-07', 'https://x')`},
		{`INSERT INTO items (id, type, slug, wikidata_qid, original_title, runtime_minutes, provenance, external_ids)
VALUES ($1, 'movie', 'arrival', 'Q20022969', 'Arrival', 116, '{"original_title":"community","runtime_minutes":"wikidata"}',
  '{"tmdb":"329865"}')`, movie},
		{`INSERT INTO item_translations (item_id, locale, title, title_source, synopsis, synopsis_source, synopsis_licence)
VALUES ($1, 'en', 'Arrival', 'tmdb','Linguist meets aliens.', 'wikipedia', 'CC-BY-SA-4.0')`, movie},
		{`INSERT INTO entities (id, kind, slug, default_name, provenance) VALUES ($1, 'genre', 'sci-fi', 'Sci-fi',
  '{"default_name":"wikidata"}')`, genre},
		{`INSERT INTO item_entities (item_id, role, entity_id, source) VALUES ($1, 'genre', $2, 'wikidata')`, movie, genre},
		{`INSERT INTO items (id, type, slug, wikidata_qid, page_count, provenance, external_ids)
VALUES ($1, 'book', 'emma', 'Q216018', 320, '{"page_count":"wikidata"}', '{"isbn13":"9780141439587"}')`, book2},
		{`INSERT INTO item_translations (item_id, locale, title, title_source, synopsis, synopsis_source, synopsis_licence)
VALUES ($1, 'en', 'Emma', 'openlibrary', 'Open Library text', 'openlibrary', 'pending')`, book2},
		{`INSERT INTO items (id, type, slug, wikidata_qid, provenance, visibility) VALUES ($1, 'series', 'gone', 'Q1', '{}', 'hidden')`, hidden},
	} {
		if _, err := db.ExecContext(t.Context(), q[0].(string), q[1:]...); err != nil {
			t.Fatalf("%s: %v", q[0], err)
		}
	}

	raw, err := os.ReadFile("../qa/sql/api_coverage.sql")
	if err != nil {
		t.Fatal(err)
	}
	if def := strings.Join(config.DefaultCatalogAPISources, ","); !strings.Contains(string(raw), "'"+def+"'") {
		t.Fatalf("api_coverage.sql default allowlist must be %q (config.DefaultCatalogAPISources)", def)
	}

	rows, err := db.QueryContext(t.Context(), string(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	got := map[string]float64{}
	for rows.Next() {
		var (
			typ, field string
			items      int
			pct        float64
		)
		if err := rows.Scan(&typ, &field, &items, &pct); err != nil {
			t.Fatal(err)
		}
		if pct < 0 || pct > 100 {
			t.Fatalf("%s.%s pct = %v, want 0-100", typ, field, pct)
		}
		got[typ+"."+field] = pct
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	policy := catalog.NewPolicy(catalog.AudienceAPI, config.DefaultCatalogAPISources)
	items, err := catalog.LoadAPIItems(
		t.Context(),
		db,
		[]string{c.game, c.book, movie, book2, hidden},
		"en",
		policy,
		true,
	)
	if err != nil {
		t.Fatal(err)
	}

	perType, hits := map[string]int{}, map[string]int{}
	for _, it := range items {
		d := it.Detail
		perType[d.Type]++
		has := map[string]bool{"externalIds": len(d.ExternalIDs) > 0}
		for field := range d.Provenance {
			has[field] = true
		}
		for _, e := range d.Entities {
			has[map[bool]string{true: "genres", false: "credits"}[e.Role == "genre"]] = true
		}
		for field := range has {
			if has[field] {
				hits[d.Type+"."+field]++
			}
		}
	}
	want := map[string]float64{}
	for typ, n := range perType {
		for _, field := range coverageFields {
			want[typ+"."+field] = math.Round(1000*float64(hits[typ+"."+field])/float64(n)) / 10
		}
	}

	// Hand-checked anchors, so a shared mistake in both readers still fails.
	for key, pct := range map[string]float64{
		"movie.synopsis": 100, "movie.title": 0, "game.title": 100, "book.synopsis": 0, "book.pageCount": 50, "game.genres": 0, "movie.genres": 100,
		"game.runtimeMinutes": 0, "movie.runtimeMinutes": 100, "game.cover": 0,
	} {
		if got[key] != pct {
			t.Errorf("%s = %v, want %v", key, got[key], pct)
		}
	}
	if _, ok := got["series.title"]; ok {
		t.Error("hidden items must not be counted")
	}

	keys := make([]string, 0, len(want))
	for k := range want {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		if g, ok := got[k]; !ok || g != want[k] {
			t.Errorf("%s: sql %v (present %v), serialiser %v", k, g, ok, want[k])
		}
	}
	if len(got) != len(want) {
		t.Errorf("sql returned %d rows, serialiser fields %d", len(got), len(want))
	}
}

// coverageFields are the API fields api_coverage.sql reports, named as in the JSON response.
var coverageFields = []string{
	"title", "synopsis", "releaseDate", "runtimeMinutes", "pageCount", "cover", "externalIds", "genres", "credits",
}
