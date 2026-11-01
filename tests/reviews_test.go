//go:build integration

package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

type reviewRow struct {
	Author  string
	Text    string
	Spoiler bool
}

// reviews reads a deferred reviews prop (item page or profile tab) through a partial reload; nil when absent.
func (r *listRig) reviews(viewer *models.User, component, path, prop string) ([]reviewRow, map[string]any) {
	r.t.Helper()

	c := r.app.Client(r.t)
	if viewer != nil {
		c = browserFor(r.t, r.app, *viewer)
	}

	page := c.WithHeader("X-Inertia-Partial-Component", component).WithHeader("X-Inertia-Partial-Data", prop).
		Inertia("").Get(path).AssertOK().AssertInertia(component)

	raw, ok := page.Prop(prop)
	if !ok || raw == nil {
		return nil, nil
	}

	list := raw
	var mine map[string]any
	if m, isMap := raw.(map[string]any); isMap {
		list = m["data"]
		mine, _ = m["mine"].(map[string]any)
	}

	var out []reviewRow
	for _, e := range list.([]any) {
		m := e.(map[string]any)
		body, _ := json.Marshal(m["body"])
		out = append(out, reviewRow{
			Author: m["author"].(map[string]any)["nickname"].(string), Text: string(body), Spoiler: m["spoiler"].(bool),
		})
	}

	return out, mine
}

func (r *listRig) itemPath(id string) string {
	r.t.Helper()

	var slug string

	db, err := r.app.DB()
	if err != nil {
		r.t.Fatal(err)
	}
	if err := db.QueryRowContext(r.t.Context(), `SELECT slug FROM items WHERE id = $1`, id).Scan(&slug); err != nil {
		r.t.Fatal(err)
	}

	return "/games/" + slug
}

func review(c *vukemtest.Client, itemID, body string, spoiler bool) *vukemtest.Response {
	return c.PutJSON("/items/"+itemID+"/review", map[string]any{"body": body, "spoiler": spoiler})
}

func TestReviews(t *testing.T) {
	r := newSocialRig(t, true)
	a, ac := r.member("ra@example.com", "Critic", false)
	b, bc := r.member("rb@example.com", "Second", false)
	_, pc := r.member("rp@example.com", "Hidden", true)
	_, nc := r.member("rn@example.com", "NoStars", false)
	r.optIn(a, true)

	g := r.releasedItem("game")
	for _, c := range []*vukemtest.Client{ac, bc, pc} {
		r.rate(c, g, 8)
	}

	review(nc, g, "No rating yet", false).AssertUnprocessable().AssertSee("rating-required")
	review(ac, g, strings.Repeat("x", 5001), false).AssertUnprocessable()
	review(ac, g, "   ", false).AssertUnprocessable()

	review(ac, g, "**Great** game", true).AssertStatus(http.StatusSeeOther)
	review(ac, g, "**Great** game, edited", true).AssertStatus(http.StatusSeeOther)
	review(bc, g, "Fine", false).AssertStatus(http.StatusSeeOther)
	review(pc, g, "From a private profile", false).AssertStatus(http.StatusSeeOther)

	if n := r.count(`SELECT count(*) FROM reviews WHERE item_id = $1`, g); n != 3 {
		t.Fatalf("%d reviews, want 3 (an edit updates in place)", n)
	}

	if n := r.count(`SELECT count(*) FROM activity_events WHERE actor_id = $1 AND kind = 'review_published'`,
		a.ID); n != 1 {
		t.Fatalf("%d review_published events, want 1 (edits do not repeat it)", n)
	}

	// Helpfulness first, then newest: Second's older review outranks Critic's once it has a helpful vote.
	r.exec(`UPDATE reviews SET helpful_count = 2 WHERE user_id = $1`, b.ID)

	got, _ := r.reviews(nil, "Catalog/Show", r.itemPath(g), "reviews")
	if len(got) != 2 || got[0].Author != "Second" || got[1].Author != "Critic" || !got[1].Spoiler ||
		!strings.Contains(got[1].Text, `"t":"strong"`) || !strings.Contains(got[1].Text, "edited") {
		t.Fatalf("item reviews = %+v, want Second then Critic (spoiler, bold); private author left out", got)
	}

	_, mine := r.reviews(&a, "Catalog/Show", r.itemPath(g), "reviews")
	if mine == nil || mine["source"] != "**Great** game, edited" {
		t.Fatalf("mine = %v, want the editable source", mine)
	}

	onProfile, _ := r.reviews(nil, "User/Profile", "/u/Critic?tab=reviews", "reviews")
	if len(onProfile) != 1 || onProfile[0].Author != "Critic" {
		t.Fatalf("profile reviews = %+v", onProfile)
	}

	if hidden, _ := r.reviews(nil, "User/Profile", "/u/Hidden?tab=reviews", "reviews"); hidden != nil {
		t.Fatalf("a private profile shows its reviews to anonymous: %+v", hidden)
	}

	// Deleting the rating hides the review; deleting the review removes it.
	bc.Delete("/items/" + g + "/rating").AssertStatus(http.StatusSeeOther)
	got, _ = r.reviews(nil, "Catalog/Show", r.itemPath(g), "reviews")
	if len(got) != 1 || got[0].Author != "Critic" {
		t.Fatalf("after Second's rating went, reviews = %+v", got)
	}

	ac.Delete("/items/" + g + "/review").AssertStatus(http.StatusSeeOther)
	if n := r.count(`SELECT count(*) FROM reviews WHERE user_id = $1`, a.ID); n != 0 {
		t.Fatal("DELETE kept the review")
	}

	r.app.Client(t).PutJSON("/items/"+g+"/review", map[string]any{"body": "x"}).AssertUnauthorized()
}

func TestReviewsDisabledByDefault(t *testing.T) {
	r := newSocialRig(t, false)
	_, ac := r.member("roff@example.com", "OffCritic", false)
	g := r.releasedItem("game")
	r.rate(ac, g, 8)

	review(ac, g, "hello", false).AssertNotFound()
	ac.Inertia("").Get(r.itemPath(g)).AssertOK().AssertInertia("Catalog/Show").MissingProp("reviews")
}

func TestSecurity_A05_ReviewMarkdown(t *testing.T) {
	r := newSocialRig(t, true)
	_, ac := r.member("rx@example.com", "Xsser", false)
	g := r.releasedItem("game")
	r.rate(ac, g, 8)

	review(ac, g, "<script>alert(1)</script> <img src=x onerror=alert(1)> [x](javascript:alert(1))",
		false).AssertStatus(http.StatusSeeOther)

	got, _ := r.reviews(nil, "Catalog/Show", r.itemPath(g), "reviews")
	if len(got) != 1 {
		t.Fatalf("reviews = %+v", got)
	}

	var nodes []any
	if err := json.Unmarshal([]byte(got[0].Text), &nodes); err != nil {
		t.Fatal(err)
	}

	kinds := map[string]bool{}

	var walk func([]any)
	walk = func(ns []any) {
		for _, n := range ns {
			m := n.(map[string]any)
			kinds[m["t"].(string)] = true
			if c, ok := m["c"].([]any); ok {
				walk(c)
			}
		}
	}
	walk(nodes)

	if len(kinds) != 2 || !kinds["p"] || !kinds["text"] ||
		!strings.Contains(fmt.Sprint(nodes), "<script>alert(1)</script>") {
		t.Fatalf("review nodes = %s (kinds %v), want only escaped text in a paragraph", got[0].Text, kinds)
	}

	r.app.Client(t).Get(r.itemPath(g)).AssertOK().AssertDontSee("<script>alert(1)").AssertDontSee("javascript:alert")
}
