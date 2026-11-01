//go:build integration

package tests

import (
	"fmt"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// friendsProp reads the deferred prop of component at path as viewer (nil: anonymous); partial false asks for the
// initial page, which must not carry it.
func (r *listRig) friendsProp(
	viewer *models.User, path, component, prop string, partial bool,
) (any, bool) {
	r.t.Helper()

	c := r.app.Client(r.t)
	if viewer != nil {
		c = browserFor(r.t, r.app, *viewer)
	}

	if partial {
		c = c.WithHeader("X-Inertia-Partial-Component", component).WithHeader("X-Inertia-Partial-Data", prop)
	}

	return c.Inertia("").Get(path).AssertOK().AssertInertia(component).Prop(prop)
}

func TestFriendsScore(t *testing.T) {
	r := newSocialRig(t, true)
	v, _ := r.member("fs-v@example.com", "Viewer", false)
	a, ac := r.member("fs-a@example.com", "OpenA", false)
	p, pc := r.member("fs-p@example.com", "PrivP", true)
	x, xc := r.member("fs-x@example.com", "BanX", false)
	n, nc := r.member("fs-n@example.com", "Pending", false)
	_, sc := r.member("fs-s@example.com", "Stranger", false)

	item := r.releasedItem("game")
	r.rate(ac, item, 8)
	r.rate(pc, item, 4)
	r.rate(xc, item, 2)
	r.rate(nc, item, 6)
	r.rate(sc, item, 10)

	for _, u := range []models.User{a, p, x} {
		r.acceptFollow(v, u)
	}

	r.exec(`INSERT INTO follows (follower_id, followee_id, state) VALUES ($1, $2, 'pending')`, v.ID, n.ID)
	r.exec(`UPDATE users SET banned_at = now() WHERE id = $1`, x.ID)

	// Only followed, accepted, public, active users count: OpenA alone.
	raw, ok := r.friendsProp(&v, r.itemPath(item), "Catalog/Show", "friendsScore", true)
	if !ok {
		t.Fatal("no friendsScore prop for a signed-in viewer")
	}

	got := raw.(map[string]any)
	if fmt.Sprint(got["count"]) != "1" || fmt.Sprint(got["average"]) != "8" {
		t.Fatalf("friendsScore = %v, want count 1 average 8", got)
	}

	friends := got["friends"].([]any)
	if len(friends) != 1 || friends[0].(map[string]any)["nickname"] != "OpenA" {
		t.Fatalf("friends = %v, want OpenA only", friends)
	}

	// The initial HTML never carries it (deferred, group personal), so the public page stays cacheable.
	if _, ok := r.friendsProp(&v, r.itemPath(item), "Catalog/Show", "friendsScore", false); ok {
		t.Fatal("friendsScore is in the initial page")
	}

	// Anonymous visitors and viewers who follow nobody get nothing.
	if raw, ok := r.friendsProp(nil, r.itemPath(item), "Catalog/Show", "friendsScore", true); ok && raw != nil {
		t.Fatalf("anonymous friendsScore = %v", raw)
	}

	// The home strip lists what followed people rated.
	home, ok := r.friendsProp(&v, "/", "Home/Index", "friends", true)
	if !ok {
		t.Fatal("no friends prop on the home")
	}

	rows := home.([]any)
	if len(rows) != 1 || rows[0].(map[string]any)["nickname"] != "OpenA" {
		t.Fatalf("home friends = %v, want OpenA's rating only", rows)
	}
}

func TestFriendsScore_FlagOff(t *testing.T) {
	r := newSocialRig(t, false)
	v, _ := r.member("fso-v@example.com", "Viewer", false)
	a, ac := r.member("fso-a@example.com", "OpenA", false)
	item := r.releasedItem("game")
	r.rate(ac, item, 8)
	r.acceptFollow(v, a)

	if raw, ok := r.friendsProp(&v, r.itemPath(item), "Catalog/Show", "friendsScore", true); ok && raw != nil {
		t.Fatalf("friendsScore with the flag off = %v", raw)
	}

	if raw, ok := r.friendsProp(&v, "/", "Home/Index", "friends", true); ok && raw != nil {
		t.Fatalf("home friends with the flag off = %v", raw)
	}
}
