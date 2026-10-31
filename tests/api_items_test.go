//go:build integration

package tests

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// apiCatalog is a key rig with a licence-mixed catalogue: one game whose synopsis, runtime, relation and
// external ids come from TMDB/IGDB next to Wikidata fields, and one book with an edition.
type apiCatalog struct {
	*keyRig
	client          *vukemtest.Client
	key, game, book string
}

func newAPICatalog(t *testing.T) *apiCatalog {
	t.Helper()

	// The free plan's 1 rps throttle is not under test here (a documented test fake, WithRateLimiters).
	r := newKeyRigWith(t, vukem.WithRateLimiters(func(_ string, _ int, window time.Duration) ratelimit.Limiter {
		return ratelimit.NewMemory(1000, window)
	}))
	plain := r.create(r.signIn("reader@example.com"), "ci")
	game, book, genre, studio, pub, named := uuid.NewV7().
		String(),
		uuid.NewV7().
			String(),
		uuid.NewV7().
			String(),
		uuid.NewV7().
			String(),
		uuid.NewV7().
			String(),
		uuid.NewV7().
			String()

	for _, q := range [][]any{
		{`INSERT INTO data_sources (code, name, licence, url) VALUES
  ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org'), ('tmdb', 'TMDB', 'proprietary', 'https://tmdb.org'),
  ('igdb', 'IGDB', 'proprietary', 'https://igdb.com'), ('hexascore', 'HexaScore', 'HEXASCORE-API-TERMS', 'https://x'),
  ('openlibrary', 'Open Library', 'pending', 'https://openlibrary.org')`},
		{`INSERT INTO items (id, type, slug, wikidata_qid, release_date, release_precision, runtime_minutes,
  external_ids, provenance, rating_count, score, score_c1, score_confidence, score_formula_version)
VALUES ($1, 'game', 'hollow-knight', 'Q20984815', '2017-02-24', 'day', 40,
  '{"wikidata":"Q20984815","tmdb":"999","igdb":1234}', '{"release_date":"wikidata","runtime_minutes":"tmdb"}',
  12, 87.25, 90, 'high', 1)`, game},
		{`INSERT INTO item_translations (item_id, locale, title, title_source, synopsis, synopsis_source, synopsis_licence)
VALUES ($1, 'en', 'Hollow Knight', 'wikidata', 'TMDB overview text', 'tmdb', 'proprietary')`, game},
		{`INSERT INTO entities (id, kind, slug, default_name, provenance) VALUES ($1, 'genre', 'metroidvania', 'Metroidvania', '{}'),
  ($2, 'organization', 'team-cherry', 'Team Cherry', '{}'),
  ($3, 'organization', 'igdb-published', 'IGDB Published', '{"default_name":"igdb"}'),
  ($4, 'organization', 'sourced-studio', 'Sourced Studio', '{"default_name":"wikidata"}')`, genre, studio, pub, named},
		{`INSERT INTO item_entities (item_id, role, entity_id, source) VALUES ($1, 'genre', $2, 'igdb'),
  ($1, 'developer', $3, 'wikidata'), ($1, 'publisher', $4, 'wikidata'), ($1, 'developer', $5, 'wikidata')`, game, genre, studio, pub, named},
		{`INSERT INTO items (id, type, slug, wikidata_qid, provenance) VALUES ($1, 'book', 'dune', 'Q190192', '{}')`, book},
		{`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', 'Dune', 'wikidata')`, book},
		{`INSERT INTO editions (id, work_id, openlibrary_id, title, language, page_count, isbn13, source)
VALUES ($1, $2, 'OL1M', 'Dune', 'en', 412, '9780441013593', 'openlibrary')`, uuid.NewV7().String(), book},
	} {
		if _, err := r.acc.DB.ExecContext(t.Context(), q[0].(string), q[1:]...); err != nil {
			t.Fatalf("%s: %v", q[0], err)
		}
	}

	return &apiCatalog{keyRig: r, client: r.app.Client(t).WithToken(plain), key: plain, game: game, book: book}
}

func remaining(t *testing.T, res *vukemtest.Response) int {
	t.Helper()

	n, err := strconv.Atoi(res.Header("X-Quota-Remaining"))
	if err != nil {
		t.Fatalf("X-Quota-Remaining = %q", res.Header("X-Quota-Remaining"))
	}

	return n
}

// TMDB/IGDB-sourced fields and ids never reach /api/v1, and the same fixture
// shows them on /app/v1 (SDD 05 FR-05-6, FR-05-16; ADR 0006; SDD 05 §9).
func TestAPILicenceFilter(t *testing.T) {
	r := newAPICatalog(t)

	res := r.client.GetJSON("/api/v1/games/" + r.game).AssertOK().AssertMatchesOpenAPI()
	body := res.Body()
	for _, leak := range []string{"tmdb", "igdb", "TMDB overview", "Metroidvania", "metroidvania"} {
		if strings.Contains(body, leak) {
			t.Errorf("api body leaks %q: %s", leak, body)
		}
	}
	assertNoExternalIDs(t, body)
	res.AssertJSONPath("title", "Hollow Knight").
		AssertJSONPath("releaseDate", "2017-02-24").
		AssertJSONPath("externalIds.wikidata", "Q20984815").
		AssertJSONPath("provenance.title.licence", "CC0-1.0").
		AssertJSONPath("provenance.score.source", "hexascore").
		AssertJSONPath("score", 8.7).
		AssertJSONPath("scoreByCategory.gameplay", 9.0).
		AssertJSONPath("entities.0.name", "Sourced Studio").
		AssertJSONPath("meta.omittedFields", []string{"entities", "entities.name", "externalIds", "runtimeMinutes", "synopsis"})
	// Allowed relation, but an IGDB-sourced or unsourced name: the relation stays, the name goes (ADR 0006).
	for _, name := range []string{"IGDB Published", "Team Cherry"} {
		if strings.Contains(body, name) {
			t.Errorf("api body leaks name %q: %s", name, body)
		}
	}

	site, err := catalog.LoadItemPage(t.Context(), r.acc.DB, "game", "hollow-knight", "en")
	if err != nil {
		t.Fatal(err)
	}
	if site.Item.Synopsis == nil || site.Item.RuntimeMinutes == nil || len(site.Entities["genre"]) != 1 ||
		!siteHasName(site, "IGDB Published") {
		t.Fatalf("site read model lost licensed fields: %+v", site.Item)
	}

	appBody := appCatalogClient(t, r, "catalog:read").GetJSON("/app/v1/games/" + r.game).AssertOK().Body()
	for _, want := range []string{"TMDB overview", "Metroidvania", "IGDB Published", "tmdb", "igdb"} {
		if !strings.Contains(appBody, want) {
			t.Errorf("/app/v1 body missing %q: %s", want, appBody)
		}
	}
}

// Every /api/v1 read operation of this slice answers bodies that match the OpenAPI document.
func TestAPIContract(t *testing.T) {
	r := newAPICatalog(t)

	for _, path := range []string{
		"/api/v1/games", "/api/v1/games?limit=1&sort=-score&locale=pt-BR&min_gameplay=8.5", "/api/v1/games/" + r.game,
		"/api/v1/games/genres", "/api/v1/books/" + r.book + "/editions", "/api/v1/me/usage",
		"/api/v1/games/" + uuid.NewV7().String(), "/api/v1/games?bogus=1",
	} {
		r.client.GetJSON(path).AssertMatchesOpenAPI()
	}

	r.client.GetJSON("/api/v1/books/"+r.book+"/editions").AssertOK().
		AssertJSONPath("data.0.isbn13", "9780441013593").
		AssertJSONPath("data.0.title", "Dune").
		AssertJSONPath("data.0.meta.omittedFields", []string{"format", "pageCount", "publicationYear", "publisher", "subtitle"})
	// The only genre relation is IGDB-sourced, so the API counts none (SDD 01 §6.1 "entity relations").
	r.client.GetJSON("/api/v1/games/genres").AssertOK().AssertJSONPath("data", []string{})
}

func TestAPIItems(t *testing.T) {
	r := newAPICatalog(t)
	path := "/api/v1/games/" + r.game

	t.Run("etag 304 counts", func(t *testing.T) {
		first := r.client.GetJSON(path).AssertOK()
		etag := first.Header("ETag")
		if !strings.HasPrefix(etag, `"`) {
			t.Fatalf("ETag = %q, want a strong tag", etag)
		}
		first.AssertHeader("Cache-Control", "private, no-cache")

		again := r.app.Client(t).WithToken(r.key).WithHeader("If-None-Match", etag).GetJSON(path).
			AssertStatus(http.StatusNotModified)
		if got, want := remaining(t, again), remaining(t, first)-1; got != want {
			t.Fatalf("X-Quota-Remaining after 304 = %d, want %d", got, want)
		}
	})

	t.Run("list pages with a signed cursor", func(t *testing.T) {
		page := r.client.GetJSON("/api/v1/games?limit=1").AssertOK().AssertJSONPath("data.0.id", r.game)
		var body struct {
			Meta struct {
				NextCursor string `json:"next_cursor"`
			} `json:"meta"`
		}
		page.JSON(&body)
		if body.Meta.NextCursor != "" {
			t.Fatalf("one game: next_cursor = %q, want none", body.Meta.NextCursor)
		}
		r.client.GetJSON("/api/v1/games?cursor=forged").AssertStatus(http.StatusBadRequest)
	})

	t.Run("bad queries are 400 and not counted", func(t *testing.T) {
		before := remaining(t, r.client.GetJSON("/api/v1/me/usage").AssertOK())
		for _, q := range []string{"?bogus=1", "?limit=11", "?year=abc", "?locale=fr", "?genre="} {
			r.client.GetJSON("/api/v1/games" + q).AssertStatus(http.StatusBadRequest)
		}
		r.client.GetJSON("/api/v1/films").AssertStatus(http.StatusBadRequest)
		after := r.client.GetJSON("/api/v1/me/usage").AssertOK()
		if got := remaining(t, after); got != before-1 {
			t.Fatalf("remaining = %d after 400s, want %d (only the usage call counted)", got, before-1)
		}
		used := 100 - remaining(t, after)
		after.AssertJSONPath("plan", "free").
			AssertJSONPath("quota.used", used).
			AssertJSONPath("quota.period", "lifetime")
	})

	t.Run("not found and merged ids", func(t *testing.T) {
		r.client.GetJSON("/api/v1/games/not-a-uuid").AssertNotFound()
		r.client.GetJSON("/api/v1/movies/" + r.game).AssertNotFound()

		gone := uuid.NewV7().String()
		if err := catalog.RecordMerge(t.Context(), r.acc.DB, catalog.ScopeItem, "game",
			catalog.Target{ID: gone, Slug: "hollow-knight-dup"}, r.game); err != nil {
			t.Fatal(err)
		}
		r.client.GetJSON("/api/v1/games/"+gone).AssertStatus(http.StatusMovedPermanently).
			AssertHeader("Location", path)
	})

	t.Run("hidden items are 404", func(t *testing.T) {
		if _, err := r.acc.DB.ExecContext(
			t.Context(),
			`UPDATE items SET visibility = 'hidden' WHERE id = $1`,
			r.book,
		); err != nil {
			t.Fatal(err)
		}
		r.client.GetJSON("/api/v1/books/" + r.book).AssertNotFound()
		r.client.GetJSON("/api/v1/books/" + r.book + "/editions").AssertNotFound()
	})
}

func siteHasName(p catalog.ItemPage, name string) bool {
	for _, links := range p.Entities {
		for _, l := range links {
			if l.Name == name {
				return true
			}
		}
	}

	return false
}

// leakedExternalID matches the fixture's TMDB/IGDB ids only as whole JSON values: random UUIDs may contain "999" or "1234".
var leakedExternalID = regexp.MustCompile(`[:\[,]\s*"?(999|1234)"?\s*[,}\]]`)

func assertNoExternalIDs(t *testing.T, body string) {
	t.Helper()

	if m := leakedExternalID.FindString(body); m != "" {
		t.Errorf("api body leaks a TMDB/IGDB id (%s): %s", m, body)
	}
}
