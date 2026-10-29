//go:build integration

package search_test

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

type goldenCase struct {
	Query  string
	Locale string
	Types  []string
	Entity bool
	Expect []string
}

func unquote(v string) string {
	if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') && v[len(v)-1] == v[0] {
		return v[1 : len(v)-1]
	}

	return v
}

func list(v string) []string {
	v = strings.TrimSuffix(strings.TrimPrefix(v, "["), "]")

	var out []string

	for part := range strings.SplitSeq(v, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}

	return out
}

// loadGolden reads the restricted YAML subset used by testdata/golden.yaml (a list of flat maps, inline lists)
// so the relevance set needs no YAML dependency.
func loadGolden(t *testing.T) []goldenCase {
	t.Helper()

	f, err := os.Open("testdata/golden.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var cases []goldenCase

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if rest, ok := strings.CutPrefix(line, "- "); ok {
			cases = append(cases, goldenCase{})
			line = rest
		}

		key, val, ok := strings.Cut(line, ":")
		if !ok || len(cases) == 0 {
			t.Fatalf("golden.yaml: bad line %q", line)
		}

		val = strings.TrimSpace(val)
		c := &cases[len(cases)-1]

		switch key {
		case "q":
			c.Query = unquote(val)
		case "locale":
			c.Locale = val
		case "types":
			c.Types = list(val)
		case "entity":
			c.Entity = val == "true"
		case "expect":
			c.Expect = list(val)
		default:
			t.Fatalf("golden.yaml: unknown key %q", key)
		}
	}

	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}

	return cases
}

func loadFixture(t *testing.T, f *fixture) {
	t.Helper()

	raw, err := os.ReadFile("testdata/fixture.sql")
	if err != nil {
		t.Fatal(err)
	}

	var body strings.Builder

	for line := range strings.SplitSeq(string(raw), "\n") {
		if !strings.HasPrefix(line, "--") {
			body.WriteString(line + "\n")
		}
	}

	for stmt := range strings.SplitSeq(body.String(), ";\n") {
		if strings.TrimSpace(stmt) != "" {
			exec(t, f.db, stmt)
		}
	}
}

func topSlugs(t *testing.T, s *search.Postgres, c goldenCase) []string {
	t.Helper()

	ctx := context.Background()

	if c.Entity {
		res, err := s.SearchEntities(
			ctx,
			search.EntityQuery{Text: c.Query, Locale: c.Locale, Audience: search.AudienceSite},
		)
		if err != nil {
			t.Fatal(err)
		}

		out := make([]string, len(res.Hits))
		for i, h := range res.Hits {
			out[i] = h.Slug
		}

		return out[:min(3, len(out))]
	}

	res, err := s.Search(
		ctx,
		search.Query{Text: c.Query, Locale: c.Locale, Types: c.Types, Audience: search.AudienceSite},
	)
	if err != nil {
		t.Fatal(err)
	}

	hits := res.Hits
	for _, g := range res.Groups {
		hits = append(hits, g.Hits...)
	}

	sort.SliceStable(hits, func(i, j int) bool { return hits[i].Rank > hits[j].Rank })

	out := slugs(hits)

	return out[:min(3, len(out))]
}

func TestGolden(t *testing.T) {
	cases := loadGolden(t)

	perLocale := map[string]int{}
	for _, c := range cases {
		perLocale[c.Locale]++

		if len(c.Expect) == 0 || len(c.Expect) > 3 {
			t.Fatalf("%q: expect needs 1 to 3 slugs", c.Query)
		}
	}

	if len(cases) != 60 || perLocale["en"] != 20 || perLocale["pt-BR"] != 20 || perLocale["es"] != 20 {
		t.Fatalf("golden set = %d queries %v, want 60 (20 per locale)", len(cases), perLocale)
	}

	f := &fixture{t: t, db: migratedDB(t)}
	loadFixture(t, f)

	if n := count(t, f.db, `SELECT count(*) FROM items`); n != 3000 {
		t.Fatalf("fixture items = %d, want 3000", n)
	}

	s := searcher(f.db)

	for _, c := range cases {
		t.Run(c.Locale+"/"+c.Query, func(t *testing.T) {
			got := topSlugs(t, s, c)
			for _, want := range c.Expect {
				if !slices.Contains(got, want) {
					t.Errorf("top 3 = %v, missing %q", got, want)
				}
			}
			if t.Failed() {
				t.Log(debugTop(s, c))
			}
		})
	}
}

// debugTop lists the top 6 hits with ranks to make a failing golden line diagnosable.
func debugTop(s *search.Postgres, c goldenCase) string {
	if c.Entity {
		return ""
	}

	res, err := s.Search(context.Background(), search.Query{
		Text: c.Query, Locale: c.Locale, Types: c.Types, Audience: search.AudienceSite, PerPage: 6,
	})
	if err != nil {
		return err.Error()
	}

	hits := res.Hits
	for _, g := range res.Groups {
		hits = append(hits, g.Hits...)
	}

	var b strings.Builder
	for _, h := range hits {
		fmt.Fprintf(&b, "%s=%.2f(%s) ", h.Slug, h.Rank, h.MatchedTitle)
	}

	return b.String()
}
