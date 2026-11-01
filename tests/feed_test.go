//go:build integration

package tests

import (
	"fmt"
	"net/http"
	"slices"
	"testing"

	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

func (r *listRig) releasedItem(typ string) string {
	r.t.Helper()

	r.n++
	id := uuid.NewV7().String()
	r.exec(
		`INSERT INTO items (id, type, slug, wikidata_qid, visibility, release_date, release_precision, release_status)
VALUES ($1, $2, $3, $4, 'public', '2020-01-01', 'day', 'released')`,
		id,
		typ,
		"s-"+id,
		fmt.Sprintf("Q%d", 9500+r.n),
	)

	return id
}

func (r *listRig) rate(c *vukemtest.Client, itemID string, score float64) {
	r.t.Helper()

	c.PutJSON("/items/"+itemID+"/rating", map[string]any{"scores": map[string]float64{
		"gameplay": score, "fun": score, "sound": score, "storytelling": score, "technical": score, "visual": score,
	}}).AssertStatus(http.StatusSeeOther)
}

func (r *listRig) acceptFollow(follower, followee models.User) {
	r.t.Helper()
	r.exec(`INSERT INTO follows (follower_id, followee_id, state, accepted_at) VALUES ($1, $2, 'accepted', now())`,
		follower.ID, followee.ID)
}

func (r *listRig) optIn(u models.User, on bool) {
	r.t.Helper()
	r.exec(`UPDATE users SET feed_opt_in = $2 WHERE id = $1`, u.ID, on)
}

type feedEntry struct {
	Kind, Actor, Item, List string
}

func (r *listRig) feed(c *vukemtest.Client) []feedEntry {
	r.t.Helper()

	page := c.Inertia("").Get("/feed").AssertOK().AssertInertia("Feed/Index")
	raw, ok := page.Prop("events")
	if !ok {
		r.t.Fatal("feed: no events prop")
	}

	var out []feedEntry
	for _, e := range raw.([]any) {
		m := e.(map[string]any)
		entry := feedEntry{Kind: m["kind"].(string), Actor: m["actor"].(map[string]any)["nickname"].(string)}
		if it, ok := m["item"].(map[string]any); ok {
			entry.Item = it["id"].(string)
		}
		if l, ok := m["list"].(map[string]any); ok {
			entry.List = l["title"].(string)
		}
		out = append(out, entry)
	}

	return out
}

func TestFeedDisabledByDefault(t *testing.T) {
	r := newSocialRig(t, false)
	a, ac := r.member("fd-a@example.com", "OffFeedA", false)
	r.optIn(a, true)
	r.rate(ac, r.releasedItem("game"), 8)

	ac.Inertia("").Get("/feed").AssertNotFound()

	if n := r.count(`SELECT count(*) FROM activity_events`); n != 0 {
		t.Fatalf("activity events with the flag off = %d, want 0", n)
	}
}

func TestFeedPrivacy(t *testing.T) {
	r := newSocialRig(t, true)
	v, vc := r.member("fv@example.com", "Viewer", false)
	a, ac := r.member("fa2@example.com", "OpenA", false)
	b, bc := r.member("fb2@example.com", "ShyB", false)
	p, pc := r.member("fp2@example.com", "PrivP", true)
	x, xc := r.member("fx2@example.com", "BanX", false)
	_, wc := r.member("fw@example.com", "Nobody", false)

	for _, u := range []models.User{a, p, x} {
		r.optIn(u, true)
	}

	for _, u := range []models.User{a, b, p, x} {
		r.acceptFollow(v, u)
	}

	g1, g2, g3 := r.releasedItem("game"), r.releasedItem("game"), r.releasedItem("game")
	r.rate(ac, g1, 8)
	r.rate(ac, g1, 9) // collapses into the first event (10-minute window)
	ac.PutJSON("/items/"+g2+"/status", map[string]any{"status": "completed"}).AssertStatus(http.StatusSeeOther)
	ac.PostJSON("/lists", map[string]any{"type": "game", "title": "Secret", "isPublic": false}).
		AssertStatus(http.StatusSeeOther)
	ac.PostJSON("/lists", map[string]any{"type": "game", "title": "Shared", "isPublic": true}).
		AssertStatus(http.StatusSeeOther)

	var secret, shared string
	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}
	_ = db.QueryRowContext(t.Context(), `SELECT id FROM lists WHERE title = 'Secret'`).Scan(&secret)
	_ = db.QueryRowContext(t.Context(), `SELECT id FROM lists WHERE title = 'Shared'`).Scan(&shared)
	ac.PutJSON("/lists/"+secret+"/items/"+g3, map[string]any{}).AssertStatus(http.StatusSeeOther)
	ac.PutJSON("/lists/"+shared+"/items/"+g3, map[string]any{}).AssertStatus(http.StatusSeeOther)

	r.rate(bc, g1, 7) // not opted in
	r.rate(pc, g2, 6) // private, opted in, viewer is an accepted follower
	r.rate(xc, g3, 5)
	r.exec(`UPDATE users SET banned_at = now(), ban_reason = 'spam test' WHERE id = $1`, x.ID)

	got := r.feed(vc)
	want := []feedEntry{
		{Kind: "rating_saved", Actor: "PrivP", Item: g2},
		{Kind: "list_item_added", Actor: "OpenA", Item: g3, List: "Shared"},
		{Kind: "list_created", Actor: "OpenA", List: "Shared"},
		{Kind: "status_changed", Actor: "OpenA", Item: g2},
		{Kind: "rating_saved", Actor: "OpenA", Item: g1},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("feed =\n%v\nwant\n%v", got, want)
	}

	if n := len(r.feed(wc)); n != 0 {
		t.Fatalf("feed of a user who follows nobody has %d entries", n)
	}

	// Read-time filters: an item leaving `public`, a list turning private, the actor opting out.
	r.exec(`UPDATE items SET visibility = 'hidden' WHERE id = $1`, g1)
	r.exec(`UPDATE lists SET is_public = false WHERE id = $1`, shared)
	r.optIn(p, false)

	got = r.feed(vc)
	want = []feedEntry{{Kind: "status_changed", Actor: "OpenA", Item: g2}}
	if !slices.Equal(got, want) {
		t.Fatalf("after filters feed =\n%v\nwant\n%v", got, want)
	}

	r.app.Client(t).Inertia("").Get("/feed").AssertStatus(http.StatusFound)
}

// TestSocialFlagShared: the nav reads `flags.social` to show the Feed link (IMPL-web-feed-nav).
func TestSocialFlagShared(t *testing.T) {
	for _, on := range []bool{false, true} {
		r := newSocialRig(t, on)
		_, c := r.member(fmt.Sprintf("flag-%v@example.com", on), fmt.Sprintf("Flag%v", on), false)

		raw, ok := c.Inertia("").Get("/library").AssertOK().AssertInertia("Library/Index").Prop("flags")
		if !ok {
			t.Fatalf("social=%v: no flags prop", on)
		}
		if got := raw.(map[string]any)["social"]; got != on {
			t.Fatalf("flags.social = %v, want %v", got, on)
		}
	}
}
