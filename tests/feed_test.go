//go:build integration

package tests

import (
	"fmt"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/social"
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
	ac.PatchJSON("/profile/feed", map[string]any{"feedOptIn": true}).AssertNotFound()
	ac.Inertia("").Get("/profile").AssertOK().AssertInertia("Profile/Index").MissingProp("social")
	ac.Inertia("").Get("/u/OffFeedA?tab=activity").AssertOK().AssertInertia("User/Profile").MissingProp("activity").
		MissingProp("social")

	fresh := makeUser(t, r.app, "fd-new@example.com", "user", true)
	browserFor(t, r.app, fresh).PostJSON("/onboarding", map[string]any{"nickname": "offnew", "feedOptIn": true}).
		AssertRedirect("/dashboard")
	if n := r.count(`SELECT count(*) FROM users WHERE id = $1 AND feed_opt_in`, fresh.ID); n != 0 {
		t.Fatal("onboarding opted in while the flag is off")
	}

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

func TestFeedOptInSetting(t *testing.T) {
	r := newSocialRig(t, true)
	u, uc := r.member("oi@example.com", "OptMe", false)

	opted := func() bool {
		return r.count(`SELECT count(*) FROM users WHERE id = $1 AND feed_opt_in`, u.ID) == 1
	}

	uc.Inertia("").Get("/profile").AssertOK().AssertInertia("Profile/Index").PropEquals("social.feedOptIn", false)

	uc.PatchJSON("/profile/feed", map[string]any{"feedOptIn": true}).AssertRedirect("/profile")
	if !opted() {
		t.Fatal("PATCH /profile/feed true did not opt in")
	}

	uc.Inertia("").Get("/profile").AssertOK().AssertInertia("Profile/Index").PropEquals("social.feedOptIn", true)

	uc.PatchJSON("/profile/feed", map[string]any{"feedOptIn": false}).AssertRedirect("/profile")
	if opted() {
		t.Fatal("PATCH /profile/feed false did not opt out")
	}

	r.app.Client(t).PatchJSON("/profile/feed", map[string]any{"feedOptIn": true}).AssertStatus(http.StatusUnauthorized)

	// Onboarding offers the same switch (FR-14-18); unchecked keeps the default off.
	yes := makeUser(t, r.app, "oi-yes@example.com", "user", true)
	yc := browserFor(t, r.app, yes)
	yc.Inertia("").Get("/onboarding").AssertOK().AssertInertia("Auth/Onboarding").PropEquals("socialEnabled", true)
	yc.PostJSON("/onboarding", map[string]any{"nickname": "optyes", "feedOptIn": true}).AssertRedirect("/dashboard")

	no := makeUser(t, r.app, "oi-no@example.com", "user", true)
	browserFor(t, r.app, no).PostJSON("/onboarding", map[string]any{"nickname": "optno"}).AssertRedirect("/dashboard")

	if n := r.count(`SELECT count(*) FROM users WHERE id IN ($1, $2) AND feed_opt_in`, yes.ID, no.ID); n != 1 ||
		r.count(`SELECT count(*) FROM users WHERE id = $1 AND feed_opt_in`, yes.ID) != 1 {
		t.Fatal("onboarding opt-in: want only the user who ticked the box opted in")
	}
}

// activity reads the deferred `activity` prop of the profile tab as viewer (nil: anonymous); -1 when absent.
func (r *listRig) activity(viewer *models.User, nickname string) int {
	r.t.Helper()

	c := r.app.Client(r.t)
	if viewer != nil {
		c = browserFor(r.t, r.app, *viewer)
	}

	page := c.WithHeader("X-Inertia-Partial-Component", "User/Profile").
		WithHeader("X-Inertia-Partial-Data", "activity").
		Inertia("").Get("/u/" + nickname + "?tab=activity").AssertOK().AssertInertia("User/Profile")

	raw, ok := page.Prop("activity")
	if !ok || raw == nil {
		return -1
	}

	return len(raw.([]any))
}

func TestProfileActivityTab(t *testing.T) {
	r := newSocialRig(t, true)
	a, ac := r.member("pa@example.com", "OpenAct", false)
	p, pc := r.member("pp@example.com", "PrivAct", true)
	f, _ := r.member("pf@example.com", "FanAct", false)
	s, _ := r.member("ps@example.com", "Stranger", false)

	r.optIn(a, true)
	r.optIn(p, true)
	r.acceptFollow(f, p)

	g, h := r.releasedItem("game"), r.releasedItem("game")
	r.rate(ac, g, 8)
	r.rate(ac, h, 7)
	r.rate(pc, g, 6)
	r.exec(`UPDATE items SET visibility = 'hidden' WHERE id = $1`, h)

	cases := []struct {
		name   string
		viewer *models.User
		nick   string
		want   int
	}{
		{"anonymous on a public opted-in profile", nil, "OpenAct", 1},
		{"stranger on a private profile", &s, "PrivAct", -1},
		{"anonymous on a private profile", nil, "PrivAct", -1},
		{"accepted follower on a private profile", &f, "PrivAct", 1},
		{"owner of a private profile", &p, "PrivAct", 1},
	}
	for _, tc := range cases {
		if got := r.activity(tc.viewer, tc.nick); got != tc.want {
			t.Errorf("%s: activity = %d, want %d", tc.name, got, tc.want)
		}
	}

	// Opting out hides past events from others at once; the owner still sees them.
	r.optIn(a, false)
	if got := r.activity(nil, "OpenAct"); got != -1 {
		t.Errorf("after opt-out anonymous activity = %d, want absent", got)
	}

	if got := r.activity(&a, "OpenAct"); got != 1 {
		t.Errorf("after opt-out owner activity = %d, want 1", got)
	}
}

func TestActivityRetention(t *testing.T) {
	r := newSocialRig(t, true)
	a, _ := r.member("ar@example.com", "Retained", false)

	r.exec(`INSERT INTO activity_events (id, actor_id, kind, created_at) VALUES
  ($1, $3, 'list_created', now() - interval '181 days'), ($2, $3, 'list_created', now() - interval '179 days')`,
		uuid.NewV7().String(), uuid.NewV7().String(), a.ID)

	db, err := r.app.DB()
	if err != nil {
		t.Fatal(err)
	}

	n, err := social.PruneActivity(t.Context(), db, time.Now())
	if err != nil || n != 1 {
		t.Fatalf("PruneActivity = %d, %v; want 1 deleted", n, err)
	}

	if left := r.count(`SELECT count(*) FROM activity_events`); left != 1 {
		t.Fatalf("%d events left, want the 179-day one", left)
	}

	sched, err := r.app.Schedule()
	if err != nil {
		t.Fatal(err)
	}

	want := "job:" + jobs.PruneActivity{}.Kind()
	for _, task := range sched.Tasks() {
		if task.Name() == want && task.Expression() == "40 4 * * *" {
			return
		}
	}

	t.Fatalf("schedule missing %s at 04:40", want)
}

func TestFeedPurgedOnAnonymise(t *testing.T) {
	r := newSocialRig(t, true)
	gone, gc := r.member("gone-feed@example.com", "GoneFeed", false)
	stay, _ := r.member("stay-feed@example.com", "StayFeed", false)

	r.optIn(gone, true)
	r.acceptFollow(gone, stay)
	r.acceptFollow(stay, gone)
	reviewed := r.releasedItem("game")
	r.rate(gc, reviewed, 8)
	review(gc, reviewed, "Identifying text", false).AssertStatus(http.StatusSeeOther)
	list := r.publicList(stay, reviewed, true)
	comment(gc, "list", list, "Identifying comment", "").AssertStatus(http.StatusSeeOther)
	react(gc, "list", list, "like").AssertStatus(http.StatusSeeOther)

	r.exec(`UPDATE users SET deletion_scheduled_at = now() - interval '31 days' WHERE id = $1`, gone.ID)

	if done, err := r.acc.PurgeDue(t.Context()); err != nil || len(done) != 1 {
		t.Fatalf("PurgeDue = %v, %v", done, err)
	}

	if n := r.count(`SELECT count(*) FROM follows WHERE $1 IN (follower_id, followee_id)`, gone.ID); n != 0 {
		t.Errorf("%d follows left after the purge", n)
	}

	if n := r.count(`SELECT count(*) FROM activity_events WHERE actor_id = $1`, gone.ID); n != 0 {
		t.Errorf("%d activity events left after the purge", n)
	}

	if n := r.count(`SELECT count(*) FROM users WHERE id = $1 AND feed_opt_in`, gone.ID); n != 0 {
		t.Error("feed opt-in survived the purge")
	}

	if n := r.count(`SELECT count(*) FROM reviews WHERE user_id = $1`, gone.ID); n != 0 {
		t.Errorf("%d reviews left after the purge", n)
	}

	if n := r.count(`SELECT (SELECT count(*) FROM comments WHERE user_id = $1) + (SELECT count(*) FROM reactions
WHERE user_id = $1) + (SELECT like_count FROM lists WHERE id = $2)`, gone.ID, list); n != 0 {
		t.Errorf("%d comments, reactions or list likes left after the purge", n)
	}
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
