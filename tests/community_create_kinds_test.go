//go:build integration

package tests

import (
	"testing"
)

// The create form switches between item, entity and edition creates with ?target=; each kind loads its own
// fields (the create-only ones included) and its own latest draft.
func TestCommunityCreateTargetKinds(t *testing.T) {
	t.Parallel()

	app := communityApp(t, true)
	u := makeUser(t, app, "create-kinds@example.com", "user", true)
	c := app.Client(t).ActingAs("web", u)

	c.Get("/suggest/new/game").AssertInertia("Community/SuggestCreate").PropEquals("target", "item")
	c.Get("/suggest/new/game?target=entity").AssertInertia("Community/SuggestCreate").
		PropEquals("target", "entity").
		PropEquals("fields", []any{"title", "cover", "external_ids.wikidata", "entity_kind"}).
		PropEquals("draft", nil)
	c.Get("/suggest/new/book?target=edition").AssertInertia("Community/SuggestCreate").
		PropEquals("fields", []any{
			"title", "cover", "external_ids.isbn13", "external_ids.wikidata", "external_ids.openlibrary",
			"work_id", "language",
		})
	c.Get("/suggest/new/movie?target=edition").AssertNotFound()
	c.Get("/suggest/new/game?target=user").AssertNotFound()
	c.Get("/suggest/new/game?kind=merge&target=entity").AssertNotFound()

	body := map[string]any{
		"kind": "create", "domain": "game", "targetKind": "entity", "version": 0, "locale": "en",
		"sourceUrls": []string{"https://www.wikidata.org/wiki/Q42"},
		"changes": map[string]any{
			"title":       map[string]any{"before": nil, "after": "Douglas Adams"},
			"entity_kind": map[string]any{"before": nil, "after": "person"},
		},
	}
	c.PostJSON("/suggestions", body).AssertRedirect("/suggest/new/game?kind=create&target=entity")
	c.Get("/suggest/new/game?target=entity").AssertInertia("Community/SuggestCreate").
		PropEquals("draft.changes.entity_kind.after", "person")
	c.Get("/suggest/new/game").AssertInertia("Community/SuggestCreate").PropEquals("draft", nil)
}
