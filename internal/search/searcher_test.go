//go:build integration

package search_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

type fixture struct {
	t  *testing.T
	db *database.DB
	n  int
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	db := migratedDB(t)
	exec(t, db, `INSERT INTO data_sources (code, name, licence, url) VALUES
		('hexascore', 'x', 'proprietary', 'https://x'), ('thirdparty', 'y', 'proprietary', 'https://y')`)

	return &fixture{t: t, db: db}
}

func (f *fixture) item(typ, slug, title, source string, popularity float64) string {
	f.t.Helper()

	f.n++
	id := uuid.NewV7().String()
	exec(f.t, f.db, `INSERT INTO items (id, type, slug, wikidata_qid, popularity_prior) VALUES ($1, $2, $3, $4, $5)`,
		id, typ, slug, fmt.Sprintf("Q%d", 1000+f.n), popularity)
	exec(f.t, f.db, `INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', $2, $3)`,
		id, title, source)

	return id
}

func slugs(h []search.Hit) []string {
	out := make([]string, len(h))
	for i := range h {
		out[i] = h[i].Slug
	}

	return out
}

func searcher(db *database.DB) *search.Postgres { return search.New(db, []string{"hexascore"}) }

func TestSearchTiers(t *testing.T) {
	f := newFixture(t)
	f.item("game", "exact", "Zelda", "hexascore", 0)
	f.item("game", "prefix", "Zelda Breath of the Wild", "hexascore", 0)
	f.item("game", "fuzzy", "The Legend of Zelda", "hexascore", 1000) // popularity cannot beat a tier
	f.item("game", "typo", "Zleda Tales", "hexascore", 0)
	f.item("game", "other", "Mario", "hexascore", 0)

	res, err := searcher(f.db).Search(context.Background(), search.Query{
		Text: "zelda", Types: []string{"game"}, Audience: search.AudienceSite,
	})
	if err != nil {
		t.Fatal(err)
	}

	got := slugs(res.Hits)
	if len(got) < 3 || got[0] != "exact" || got[1] != "prefix" || got[2] != "fuzzy" {
		t.Fatalf("order = %v", got)
	}

	res, err = searcher(
		f.db,
	).Search(context.Background(), search.Query{Text: "zleda", Types: []string{"game"}, Audience: search.AudienceSite})
	if err != nil || len(res.Hits) == 0 || !strings.Contains(res.Hits[0].Title, "Z") {
		t.Fatalf("typo search: %+v %v", res, err)
	}
}

func TestSearchGroupedAndExcluded(t *testing.T) {
	f := newFixture(t)
	f.item("game", "tlou", "The Last of Us", "hexascore", 0)
	f.item("game", "tlou2", "The Last of Us Part II", "hexascore", 0)
	f.item("movie", "tlou-movie", "The Last of Us Movie", "hexascore", 0)

	res, err := searcher(f.db).Search(context.Background(), search.Query{
		Text: `"the last of us" -part`, Audience: search.AudienceSite,
	})
	if err != nil {
		t.Fatal(err)
	}

	var all []string
	for _, g := range res.Groups {
		all = append(all, slugs(g.Hits)...)
	}

	for _, s := range all {
		if s == "tlou2" {
			t.Fatalf("excluded title returned: %v", all)
		}
	}

	if len(res.Groups) != 2 || res.Groups[0].Type != "game" || res.Groups[1].Type != "movie" {
		t.Fatalf("groups = %+v", res.Groups)
	}

	none, err := searcher(
		f.db,
	).Search(context.Background(), search.Query{Text: "-zelda", Audience: search.AudienceSite})
	if err != nil || len(none.Groups) != 0 || none.DidYouMean != nil {
		t.Fatalf("exclusion-only: %+v %v", none, err)
	}
}

func TestAudienceIntegrationAPIOnlyAllowlisted(t *testing.T) {
	f := newFixture(t)
	f.item("game", "allowed", "Hollow Knight", "hexascore", 0)
	f.item("game", "blocked", "Hollow Night", "thirdparty", 0)
	id := f.item("game", "alt-blocked", "Something Else", "hexascore", 0)
	exec(
		f.t,
		f.db,
		`INSERT INTO item_alt_titles (item_id, title, kind, source) VALUES ($1, 'Silksong', 'alias', 'thirdparty')`,
		id,
	)

	s := searcher(f.db)

	api, err := s.Search(
		context.Background(),
		search.Query{Text: "hollow", Types: []string{"game"}, Audience: search.AudienceAPI},
	)
	if err != nil {
		t.Fatal(err)
	}

	if got := slugs(api.Hits); len(got) != 1 || got[0] != "allowed" {
		t.Fatalf("api hits = %v", got)
	}

	site, err := s.Search(
		context.Background(),
		search.Query{Text: "hollow", Types: []string{"game"}, Audience: search.AudienceSite},
	)
	if err != nil || len(site.Hits) != 2 {
		t.Fatalf("site hits = %+v %v", site.Hits, err)
	}

	for audience, want := range map[string]int{search.AudienceAPI: 0, search.AudienceSite: 1} {
		r, err := s.Search(
			context.Background(),
			search.Query{Text: "silksong", Types: []string{"game"}, Audience: audience},
		)
		if err != nil || len(r.Hits) != want {
			t.Fatalf("%s alt title hits = %d (%v), want %d", audience, len(r.Hits), err, want)
		}
	}

	closed := search.New(f.db, nil)
	r, err := closed.Search(
		context.Background(),
		search.Query{Text: "hollow", Types: []string{"game"}, Audience: search.AudienceAPI},
	)
	if err != nil || len(r.Hits) != 0 {
		t.Fatalf("empty allowlist must deny all: %+v %v", r.Hits, err)
	}
}

func TestSearchEditionByISBN(t *testing.T) {
	f := newFixture(t)
	work := f.item("book", "dom-casmurro", "Dom Casmurro", "hexascore", 0)
	exec(f.t, f.db, `INSERT INTO editions (id, work_id, isbn13, title, language, source)
		VALUES ($1, $2, '9780306406157', 'Dom Casmurro Edicao Especial', 'pt-BR', 'hexascore')`, uuid.NewV7().String(), work)

	s := searcher(f.db)

	for _, q := range []string{"978-0-306-40615-7", "edicao especial"} {
		res, err := s.Search(
			context.Background(),
			search.Query{Text: q, Types: []string{"book"}, Audience: search.AudienceSite},
		)
		if err != nil || len(res.Hits) != 1 || res.Hits[0].Slug != "dom-casmurro" {
			t.Fatalf("%q: %+v %v", q, res, err)
		}

		if m := res.Hits[0].MatchedEdition; m == nil || m.ISBN13 != "9780306406157" {
			t.Fatalf("%q matchedEdition = %+v", q, m)
		}
	}
}

func TestSearchPageBeyondCap(t *testing.T) {
	f := newFixture(t)

	_, err := searcher(
		f.db,
	).Search(context.Background(), search.Query{Text: "zelda", Page: 11, PerPage: 20, Audience: search.AudienceAPI})
	if !errors.Is(err, search.ErrInvalidQuery) {
		t.Fatalf("err = %v", err)
	}
}

func TestSearchDidYouMean(t *testing.T) {
	f := newFixture(t)
	f.item("game", "zelda", "Zelda", "hexascore", 0)

	if err := search.DrainSuggest(context.Background(), f.db); err != nil {
		t.Fatal(err)
	}

	res, err := searcher(f.db).Search(context.Background(), search.Query{Text: "zeldx", Audience: search.AudienceSite})
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Groups) == 0 && (res.DidYouMean == nil || *res.DidYouMean != "Zelda") {
		t.Fatalf("zero results without suggestion: %+v", res)
	}
}

func TestSuggestAndEntities(t *testing.T) {
	f := newFixture(t)
	f.item("game", "zelda-1", "Zelda", "hexascore", 5)
	f.item("game", "zelda-2", "Zelda II", "hexascore", 1)

	for i := range 3 {
		id := uuid.NewV7().String()
		exec(f.t, f.db, `INSERT INTO entities (id, kind, slug, default_name) VALUES ($1, 'person', $2, $3)`,
			id, fmt.Sprintf("zel-%d", i), fmt.Sprintf("Zelman %d", i))
		exec(
			f.t,
			f.db,
			`INSERT INTO entity_translations (entity_id, locale, name, name_source) VALUES ($1, 'en', $2, 'hexascore')`,
			id,
			fmt.Sprintf("Zelman %d", i),
		)
	}

	if err := search.DrainSuggest(context.Background(), f.db); err != nil {
		t.Fatal(err)
	}

	s := searcher(f.db)

	sug, err := s.Suggest(context.Background(), search.SuggestQuery{Text: "zel"})
	if err != nil {
		t.Fatal(err)
	}

	ents := 0
	for _, x := range sug {
		if x.Kind == "person" {
			ents++
		}
	}

	if len(sug) > 8 || ents != 2 || sug[0].Title != "Zelda" {
		t.Fatalf("suggest = %+v", sug)
	}

	if short, err := s.Suggest(context.Background(), search.SuggestQuery{Text: "z"}); err != nil || len(short) != 0 {
		t.Fatalf("short suggest = %+v %v", short, err)
	}

	res, err := s.SearchEntities(
		context.Background(),
		search.EntityQuery{Text: "zelman 1", Audience: search.AudienceSite},
	)
	if err != nil || len(res.Hits) == 0 || res.Hits[0].Name != "Zelman 1" {
		t.Fatalf("entities = %+v %v", res, err)
	}

	api, err := s.SearchEntities(context.Background(), search.EntityQuery{Text: "zelman", Audience: search.AudienceAPI})
	if err != nil || len(api.Hits) != 3 {
		t.Fatalf("api entities = %+v %v", api, err)
	}
}

func TestSearchExplainUsesIndexes(t *testing.T) {
	f := newFixture(t)
	f.item("game", "zelda", "Zelda", "hexascore", 0)

	if err := search.DrainSuggest(context.Background(), f.db); err != nil {
		t.Fatal(err)
	}

	for query, index := range map[string]string{
		`SELECT 1 FROM item_translations WHERE search_tsv @@ websearch_to_tsquery('hs_en', 'zelda')`: "item_translations_tsv_gin",
		`SELECT 1 FROM item_translations WHERE hs_norm('zleda') <% title_norm`:                       "item_translations_title_trgm",
		`SELECT 1 FROM search_suggest WHERE label_norm ^@ 'zel'`:                                     "search_suggest_prefix",
	} {
		var plan strings.Builder

		err := f.db.Tx(context.Background(), func(tx *database.Tx) error {
			if _, err := tx.ExecContext(context.Background(), `SET LOCAL enable_seqscan = off`); err != nil {
				return err
			}

			rows, err := tx.QueryContext(context.Background(), `EXPLAIN `+query)
			if err != nil {
				return err
			}
			defer rows.Close()

			for rows.Next() {
				var line string
				if err := rows.Scan(&line); err != nil {
					return err
				}

				plan.WriteString(line + "\n")
			}

			return rows.Err()
		})
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(plan.String(), index) {
			t.Errorf("plan for %q misses %s:\n%s", query, index, plan.String())
		}
	}
}
