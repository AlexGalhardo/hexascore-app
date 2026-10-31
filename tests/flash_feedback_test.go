//go:build integration

package tests

import (
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"
)

// TestMutationsFlashStatus: every web mutation of ratings, status, time to beat, lists and the book edition
// answers 303 with a "status" flash the layout shows as a toast (SDD 09 §7.9).
func TestMutationsFlashStatus(t *testing.T) {
	r := newListRig(t)
	u, _ := r.user("flash@example.com")
	web := browserFor(t, r.app, u)
	game, book := r.item("game", "public"), r.item("book", "public")

	list := "0197f3a0-0000-7000-8000-0000000000c3"
	r.exec(`INSERT INTO lists (id, user_id, type, title, is_public) VALUES ($1, $2, 'game', 'L', true)`, list, u.ID)

	steps := []struct {
		name, want string
		do         func() *vukemtest.Response
	}{
		{"rating put", "Rating saved.", func() *vukemtest.Response {
			return web.PutJSON(
				"/items/"+game+"/rating",
				map[string]any{
					"scores": map[string]float64{
						"gameplay":     7.5,
						"fun":          7.5,
						"sound":        7.5,
						"storytelling": 7.5,
						"technical":    7.5,
						"visual":       7.5,
					},
				},
			)
		}},
		{
			"rating delete",
			"Rating removed.",
			func() *vukemtest.Response { return web.Delete("/items/" + game + "/rating") },
		},
		{"status put", "Status saved.", func() *vukemtest.Response {
			return web.PutJSON("/items/"+game+"/status", map[string]any{"status": "completed"})
		}},
		{"time to beat put", "Times saved.", func() *vukemtest.Response {
			return web.PutJSON("/items/"+game+"/time-to-beat", map[string]any{"mainHours": 10})
		}},
		{
			"time to beat delete",
			"Times removed.",
			func() *vukemtest.Response { return web.Delete("/items/" + game + "/time-to-beat") },
		},
		{
			"status delete",
			"Status removed.",
			func() *vukemtest.Response { return web.Delete("/items/" + game + "/status") },
		},
		{"list create", "List created.", func() *vukemtest.Response {
			return web.PostJSON("/lists", map[string]any{"type": "game", "title": "T", "isPublic": true})
		}},
		{"list update", "List updated.", func() *vukemtest.Response {
			return web.PatchJSON("/lists/"+list, map[string]any{"title": "New"})
		}},
		{"list item add", "Added to the list.", func() *vukemtest.Response {
			return web.PutJSON("/lists/"+list+"/items/"+game, map[string]any{})
		}},
		{"list item remove", "Removed from the list.", func() *vukemtest.Response {
			return web.Delete("/lists/" + list + "/items/" + game)
		}},
		{"list delete", "List deleted.", func() *vukemtest.Response { return web.Delete("/lists/" + list) }},
		{
			"edition delete",
			"Edition forgotten.",
			func() *vukemtest.Response { return web.Delete("/items/" + book + "/edition") },
		},
	}

	for _, s := range steps {
		s.do().AssertStatus(303).AssertFlash("status", s.want)
	}
}
