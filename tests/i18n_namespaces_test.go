//go:build integration

package tests

import (
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestI18nPageNamespaces: heavy namespaces (docs, legal bodies, help, billing, API keys, contact) reach only
// their own pages, so the shared messages prop stays small on item and home pages (PERF-i18n-page-namespaces).
func TestI18nPageNamespaces(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)
	c := app.Client(t).Inertia("")

	for _, tc := range []struct {
		path, own string
		others    []string
	}{
		{"/", "nav.", []string{"terms.", "policy.", "docs.", "help.", "billing.", "contact."}},
		{"/terms", "terms.", []string{"policy.", "docs.", "billing."}},
		{"/policy", "policy.", []string{"terms.", "docs."}},
		{"/docs", "docs.", []string{"terms.", "policy."}},
		{"/pt-br/terms", "terms.", []string{"policy.", "docs."}},
		{"/pricing", "billing.", []string{"terms.", "docs."}},
		{"/contact", "contact.", []string{"terms.", "docs."}},
		{"/help/score", "help.", []string{"terms.", "docs."}},
		// Reviews, comments and reactions render on item, list and profile pages (P7.B.1.1, P7.B.1.2): base list.
		{"/", "reviews.", nil},
		{"/", "comments.", nil},
		{"/", "reactions.", nil},
	} {
		body := c.Get(tc.path).AssertOK().Body()
		if !strings.Contains(body, `"`+tc.own) {
			t.Errorf("%s: messages lack %q", tc.path, tc.own)
		}

		for _, o := range tc.others {
			if strings.Contains(body, `"`+o) {
				t.Errorf("%s: messages carry %q, which the page does not use", tc.path, o)
			}
		}
	}
}

// TestI18nOnceKeyPerNamespaceSet: pages with different namespace sets must not share one once-prop key, or a
// client holding the home catalogue would never receive the terms namespace after navigating.
func TestI18nOnceKeyPerNamespaceSet(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)
	c := app.Client(t).Inertia("")

	home := c.Get("/").AssertOK().Body()
	terms := c.Get("/terms").AssertOK().Body()

	key := func(body string) string {
		i := strings.Index(body, `"i18n:`)
		if i < 0 {
			return ""
		}

		rest := body[i+1:]

		return rest[:strings.Index(rest, `"`)]
	}

	if key(home) == "" || key(home) == key(terms) {
		t.Fatalf("once keys: home %q, terms %q; want two different non-empty keys", key(home), key(terms))
	}
}
