//go:build integration

package tests

import (
	"fmt"
	"testing"
)

// hitBySlug finds a hit in the list prop at path.
func hitBySlug(t *testing.T, items any, slug string) map[string]any {
	t.Helper()

	list, _ := items.([]any)
	for _, it := range list {
		if m, ok := it.(map[string]any); ok && m["slug"] == slug {
			return m
		}
	}

	t.Fatalf("hit %q not in %v", slug, items)

	return nil
}

func TestSearchHitMeta(t *testing.T) {
	app, exec := searchApp(t)

	exec(`UPDATE items SET release_date = '1986-02-21', rating_count = 7, score = 84 WHERE slug = 'game-0'`)
	exec(`UPDATE items SET rating_count = 3, score = 70 WHERE slug = 'game-1'`)

	check := func(t *testing.T, items any) {
		t.Helper()

		scored := hitBySlug(t, items, "game-0")
		if fmt.Sprint(scored["year"]) != "1986" || fmt.Sprint(scored["score"]) != "8.4" {
			t.Fatalf("scored hit: year=%v score=%v", scored["year"], scored["score"])
		}

		gated := hitBySlug(t, items, "game-1")
		if v, ok := gated["score"]; !ok || v != nil {
			t.Fatalf("score under 5 ratings must be present and null, got %v (present %v)", v, ok)
		}

		if v, ok := gated["year"]; !ok || v != nil {
			t.Fatalf("missing release date must give null year, got %v (present %v)", v, ok)
		}
	}

	t.Run("typed", func(t *testing.T) {
		page := app.Client(t).Inertia("").Get("/search?q=zelda&type=game").AssertOK().AssertInertia("Search/Index")
		items, _ := page.Prop("results")
		check(t, items)
	})

	t.Run("grouped", func(t *testing.T) {
		page := app.Client(t).Inertia("").Get("/search?q=zel").AssertOK().AssertInertia("Search/Index")
		items, _ := page.Prop("groups.0.items")
		check(t, items)
	})
}
