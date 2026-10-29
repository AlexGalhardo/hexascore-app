//go:build integration

package catalog_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func slugs(p catalog.ListPage) string {
	var s []string
	for _, i := range p.Items {
		s = append(s, i.Slug)
	}
	return strings.Join(s, ",")
}

func TestListingFilters(t *testing.T) {
	url := pgtest.Database(t, "")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": url}})
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	exec(
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata','Wikidata','CC0-1.0','https://www.wikidata.org')`,
	)

	type fx struct {
		n          int
		slug, typ  string
		pop        float64
		date       *string
		rc         int
		score, c1  *float64
		visibility string
	}
	d := func(s string) *string { return &s }
	fixtures := []fx{
		{1, "a", "game", 5, d("2017-02-24"), 10, new(float64(90)), new(float64(95)), "public"},
		{2, "b", "game", 3, d("2017-06-01"), 20, new(float64(70)), new(float64(60)), "public"},
		{3, "c", "game", 3, d("2020-01-01"), 20, new(float64(50)), new(float64(40)), "public"},
		{4, "d", "game", 1, nil, 2, new(float64(99)), new(float64(99)), "public"}, // under 5 ratings: score hidden
		{5, "e", "game", 0, d("2019-05-05"), 0, nil, nil, "public"},
		{6, "f", "movie", 9, d("2017-01-01"), 50, new(float64(80)), new(float64(80)), "public"},
		{7, "g", "game", 9, d("2017-01-01"), 50, new(float64(99)), new(float64(99)), "hidden"},
	}
	id := func(n int) string { return fmt.Sprintf("00000000-0000-7000-8000-%012d", n) }
	for _, it := range fixtures {
		exec(
			`INSERT INTO items (id, type, slug, wikidata_qid, popularity_prior, release_date, rating_count, score, score_c1, visibility)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			id(it.n),
			it.typ,
			it.slug,
			fmt.Sprintf("Q%d", it.n),
			it.pop,
			it.date,
			it.rc,
			it.score,
			it.c1,
			it.visibility,
		)
	}
	exec(
		`INSERT INTO entities (id, kind, slug, default_name) VALUES ($1,'genre','rpg','RPG'), ($2,'platform','pc','PC')`,
		id(101),
		id(102),
	)
	for _, r := range []struct {
		item, ent int
		role      string
	}{{1, 101, "genre"}, {2, 101, "genre"}, {1, 102, "platform"}, {3, 102, "platform"}} {
		exec(
			`INSERT INTO item_entities (item_id, entity_id, role, source) VALUES ($1,$2,$3,'wikidata')`,
			id(r.item),
			id(r.ent),
			r.role,
		)
	}

	codec := cursorCodec(t, keyA)
	list := func(p catalog.ListParams) catalog.ListPage {
		t.Helper()
		if p.Type == "" {
			p.Type = "game"
		}
		page, err := catalog.List(ctx, db, codec, p)
		if err != nil {
			t.Fatalf("%+v: %v", p, err)
		}
		return page
	}

	// Ties break on id in the sort direction; nulls and hidden scores sort last.
	sorts := map[string]string{
		"": "a,c,b,d,e", "-popularity": "a,c,b,d,e", "popularity": "e,d,b,c,a",
		"-score": "a,b,c,e,d", "score": "c,b,a,d,e",
		"-release_date": "c,e,b,a,d", "release_date": "a,b,e,c,d",
		"-rating_count": "c,b,a,d,e", "rating_count": "e,d,a,b,c",
	}
	for sort, want := range sorts {
		if got := slugs(list(catalog.ListParams{Sort: sort})); got != want {
			t.Errorf("sort %q = %s, want %s", sort, got, want)
		}
	}

	filters := map[string]struct {
		p    catalog.ListParams
		want string
	}{
		"genre":        {catalog.ListParams{Genre: "rpg", Sort: "-score"}, "a,b"},
		"platform":     {catalog.ListParams{Platform: "pc", Sort: "-score"}, "a,c"},
		"genre+plat":   {catalog.ListParams{Genre: "rpg", Platform: "pc"}, "a"},
		"year":         {catalog.ListParams{Year: 2017, Sort: "-score"}, "a,b"},
		"min_score":    {catalog.ListParams{MinScore: new(float64(7)), Sort: "-score"}, "a,b"},
		"min_score 9":  {catalog.ListParams{MinScore: new(float64(9))}, "a"},
		"max_score":    {catalog.ListParams{MaxScore: new(float64(6)), Sort: "-score"}, "c"},
		"min_gameplay": {catalog.ListParams{MinCategory: map[string]float64{"gameplay": 6}, Sort: "-score"}, "a,b"},
		"movie":        {catalog.ListParams{Type: "movie"}, "f"},
	}
	for name, tc := range filters {
		if got := slugs(list(tc.p)); got != tc.want {
			t.Errorf("%s = %s, want %s", name, got, tc.want)
		}
	}

	bad := map[string]catalog.ListParams{
		"year low":      {Type: "game", Year: 1799},
		"year high":     {Type: "game", Year: 2101},
		"score step":    {Type: "game", MinScore: new(float64(7.3))},
		"score range":   {Type: "game", MaxScore: new(float64(10.5))},
		"category type": {Type: "game", MinCategory: map[string]float64{"plot": 5}},
		"sort":          {Type: "game", Sort: "title"},
		"limit":         {Type: "game", Limit: 101},
	}
	for name, p := range bad {
		if _, err := catalog.List(ctx, db, codec, p); !errors.Is(err, catalog.ErrInvalidQuery) {
			t.Errorf("%s: err = %v, want ErrInvalidQuery", name, err)
		}
	}

	// Keyset paging in both directions over mixed null and non-null sort values.
	for _, sort := range []string{"-score", "score", "-release_date", "-popularity"} {
		full := strings.Split(slugs(list(catalog.ListParams{Sort: sort})), ",")
		p1 := list(catalog.ListParams{Sort: sort, Limit: 2})
		if p1.Total == nil || *p1.Total != 5 || p1.PrevCursor != "" {
			t.Fatalf("%s first page total/prev: %+v", sort, p1)
		}
		seen := strings.Split(slugs(p1), ",")
		page := p1
		var pages []catalog.ListPage
		for page.NextCursor != "" {
			page = list(catalog.ListParams{Sort: sort, Limit: 2, Cursor: page.NextCursor})
			if page.Total != nil {
				t.Errorf("%s: total on later page", sort)
			}
			pages = append(pages, page)
			seen = append(seen, strings.Split(slugs(page), ",")...)
		}
		if !slices.Equal(seen, full) {
			t.Errorf("%s walked %v, want %v", sort, seen, full)
		}
		back := list(catalog.ListParams{Sort: sort, Limit: 2, Cursor: pages[len(pages)-1].PrevCursor})
		if got, want := slugs(back), slugs(pages[len(pages)-2]); got != want {
			t.Errorf("%s prev = %s, want %s", sort, got, want)
		}
	}

	next := list(catalog.ListParams{Sort: "-score", Limit: 2}).NextCursor
	other := catalog.ListParams{Type: "game", Sort: "score", Cursor: next}
	if _, err := catalog.List(ctx, db, codec, other); !errors.Is(err, catalog.ErrInvalidCursor) {
		t.Errorf("cursor reused for another sort: %v", err)
	}
}
