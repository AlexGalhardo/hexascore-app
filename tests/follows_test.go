//go:build integration

package tests

import (
	"net/http"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// newSocialRig is newListRig with the feature flag `social` set (SDD 14 §12.3, default off).
func newSocialRig(t *testing.T, on bool) *listRig {
	t.Helper()

	env := map[string]string{"DATABASE_URL": pgtest.Database(t, "")}
	if on {
		env["FLAG_SOCIAL"] = "true"
	}

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	return &listRig{t: t, app: app, acc: acc}
}

func (r *listRig) member(email, nick string, private bool) (models.User, *vukemtest.Client) {
	r.t.Helper()

	u := makeUser(r.t, r.app, email, "user", true)
	r.setNickname(u, nick, private)

	return u, browserFor(r.t, r.app, u)
}

func (r *listRig) followState(follower, followee models.User) string {
	r.t.Helper()

	db, err := r.app.DB()
	if err != nil {
		r.t.Fatal(err)
	}

	var state string
	_ = db.QueryRowContext(r.t.Context(), `SELECT state FROM follows WHERE follower_id = $1 AND followee_id = $2`,
		follower.ID, followee.ID).Scan(&state)

	return state
}

func (r *listRig) notices(u models.User, kind string) int {
	r.t.Helper()

	return r.count(`SELECT count(*) FROM user_notifications WHERE user_id = $1 AND kind = $2`, u.ID, kind)
}

func TestSocialDisabledByDefault(t *testing.T) {
	r := newSocialRig(t, false)
	_, a := r.member("off-a@example.com", "OffA", false)
	r.member("off-b@example.com", "OffB", false)

	a.PutJSON("/u/offb/follow", map[string]any{}).AssertNotFound()
	a.Delete("/u/offb/follow").AssertNotFound()
	r.app.Client(t).Inertia("").Get("/u/offb").AssertOK().AssertInertia("User/Profile").MissingProp("social")
}

func TestFollows(t *testing.T) {
	r := newSocialRig(t, true)
	a, ac := r.member("fa@example.com", "FanA", false)
	b, bc := r.member("fb@example.com", "StarB", false)
	_, pc := r.member("fp@example.com", "QuietP", true)
	x, _ := r.member("fx@example.com", "BannedX", false)
	r.exec(`UPDATE users SET banned_at = now(), ban_reason = 'spam test' WHERE id = $1`, x.ID)

	ac.PutJSON("/u/starb/follow", map[string]any{}).AssertStatus(http.StatusSeeOther)
	ac.PutJSON("/u/StarB/follow", map[string]any{}).AssertStatus(http.StatusSeeOther)

	if got := r.followState(a, b); got != "accepted" {
		t.Fatalf("follow state = %q, want accepted", got)
	}

	if n := r.notices(b, "new_follower"); n != 1 {
		t.Fatalf("new_follower notices = %d, want 1 (idempotent follow)", n)
	}

	// A private follower is not counted on the followee's public profile.
	pc.PutJSON("/u/starb/follow", map[string]any{}).AssertStatus(http.StatusSeeOther)

	anon := r.app.Client(t).Inertia("")
	anon.Get("/u/starb").AssertOK().AssertInertia("User/Profile").
		PropEquals("social.followers", 1).PropEquals("social.following", 0).MissingProp("social.state")
	anon.Get("/u/fana").AssertOK().AssertInertia("User/Profile").PropEquals("social.following", 1)
	ac.Inertia("").Get("/u/starb").AssertOK().AssertInertia("User/Profile").PropEquals("social.state", "accepted")
	bc.Inertia("").Get("/u/fana").AssertOK().AssertInertia("User/Profile").PropEquals("social.state", "")

	ac.PutJSON("/u/fana/follow", map[string]any{}).AssertUnprocessable()
	ac.PutJSON("/u/nobody/follow", map[string]any{}).AssertNotFound()
	ac.PutJSON("/u/bannedx/follow", map[string]any{}).AssertNotFound()

	ac.Delete("/u/starb/follow").AssertStatus(http.StatusSeeOther)
	ac.Delete("/u/starb/follow").AssertStatus(http.StatusSeeOther)

	if got := r.followState(a, b); got != "" {
		t.Fatalf("after unfollow state = %q, want no row", got)
	}
}

func TestSecurity_A01_PrivateProfileFollowRequiresApproval(t *testing.T) {
	r := newSocialRig(t, true)
	a, ac := r.member("ra@example.com", "AskerA", false)
	c, cc := r.member("rc@example.com", "AskerC", false)
	p, pc := r.member("rp@example.com", "LockedP", true)

	ac.PutJSON("/u/lockedp/follow", map[string]any{}).AssertStatus(http.StatusSeeOther)
	cc.PutJSON("/u/lockedp/follow", map[string]any{}).AssertStatus(http.StatusSeeOther)

	if got := r.followState(a, p); got != "pending" {
		t.Fatalf("follow of a private profile = %q, want pending", got)
	}

	if n := r.notices(p, "follow_request"); n != 2 {
		t.Fatalf("follow_request notices = %d, want 2", n)
	}

	// The notice carries the asker's nickname so the inbox can link the profile and answer the request.
	if n := r.count(`SELECT count(*) FROM user_notifications WHERE user_id = $1 AND kind = 'follow_request'
  AND payload->>'nickname' = 'askera'`, p.ID); n != 1 {
		t.Fatalf("follow_request notices naming askera = %d, want 1", n)
	}

	// The private profile still shows nothing but the nickname, and a pending request is not a follower.
	ac.Inertia("").Get("/u/lockedp").AssertOK().AssertInertia("User/Profile").
		PropEquals("social.state", "pending").MissingProp("social.followers").MissingProp("stats")
	pc.Inertia("").Get("/u/lockedp").AssertOK().AssertInertia("User/Profile").PropEquals("social.followers", 0)

	// Only the followee decides: the asker cannot accept its own request, nor a third party.
	ac.PostJSON("/me/follow-requests/askera/accept", map[string]any{}).AssertNotFound()
	cc.PostJSON("/me/follow-requests/askera/accept", map[string]any{}).AssertNotFound()

	if got := r.followState(a, p); got != "pending" {
		t.Fatalf("after foreign accept state = %q, want pending", got)
	}

	pc.PostJSON("/me/follow-requests/askera/accept", map[string]any{}).AssertStatus(http.StatusSeeOther)
	pc.PostJSON("/me/follow-requests/askerc/decline", map[string]any{}).AssertStatus(http.StatusSeeOther)
	pc.PostJSON("/me/follow-requests/askerc/decline", map[string]any{}).AssertNotFound()

	if got := r.followState(a, p); got != "accepted" {
		t.Fatalf("after accept state = %q, want accepted", got)
	}

	if got := r.followState(c, p); got != "" {
		t.Fatalf("after decline state = %q, want no row", got)
	}

	pc.Inertia("").Get("/u/lockedp").AssertOK().AssertInertia("User/Profile").PropEquals("social.followers", 1)
}
