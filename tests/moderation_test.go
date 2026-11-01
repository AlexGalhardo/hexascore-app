//go:build integration

package tests

import (
	"net/url"
	"strconv"
	"testing"

	"github.com/AlexGalhardo/vukem/uuid"
)

func TestReportsReviewComment(t *testing.T) {
	r, a, b, _, ac, bc, _, item, _, _ := socialSetup(t)
	third, cc := r.member("cc@example.com", "Third", false)
	admin := makeUser(t, r.app, adminEmail, "admin", true)
	adminC := browserFor(t, r.app, admin)

	// Republish the review once opted in, so the follower's feed has its review_published entry.
	r.optIn(a, true)
	r.acceptFollow(b, a)
	ac.Delete("/items/" + item + "/review").AssertStatus(303)
	review(ac, item, "Great game", false).AssertStatus(303)
	rv := r.str(`SELECT id::text FROM reviews WHERE user_id = $1`, a.ID)
	comment(bc, "review", rv, "Rude words", "").AssertStatus(303)
	cid := r.str(`SELECT id::text FROM comments WHERE user_id = $1`, b.ID)

	// Reporting your own review is refused; unknown ids are 404.
	ac.PostJSON("/reports", reportBody("review", rv)).AssertForbidden()
	cc.PostJSON("/reports", reportBody("review", item)).AssertNotFound()
	cc.PostJSON("/reports", reportBody("comment", "not-a-uuid")).AssertNotFound()

	cc.PostJSON("/reports", reportBody("review", rv))
	cc.PostJSON("/reports", reportBody("comment", cid))
	reviewReport := r.str(`SELECT id::text FROM reports WHERE target_type = 'review' AND target_id = $1`, rv)
	commentReport := r.str(`SELECT id::text FROM reports WHERE target_type = 'comment' AND target_id = $1`, cid)

	if len(r.feed(bc)) == 0 {
		t.Fatal("the follower's feed has no review_published entry before moderation")
	}

	// list_hidden does not apply to a review.
	adminC.PostForm("/admin/reports/"+reviewReport+"/resolve", url.Values{"outcome": {"list_hidden"}}).
		AssertUnprocessable()
	adminC.PostForm("/admin/reports/"+reviewReport+"/resolve", url.Values{"outcome": {"content_hidden"}}).
		AssertRedirect("/admin/reports/" + reviewReport)

	if n := r.count(`SELECT count(*) FROM reviews WHERE id = $1 AND hidden`, rv); n != 1 {
		t.Fatal("review not hidden")
	}
	page := r.app.Client(t).WithHeader("X-Inertia-Partial-Component", "Catalog/Show").
		WithHeader("X-Inertia-Partial-Data", "reviews").Inertia("").Get(r.itemPath(item)).AssertOK().
		AssertInertia("Catalog/Show")
	if raw, _ := page.Prop("reviews"); len(raw.(map[string]any)["data"].([]any)) != 0 {
		t.Fatalf("a hidden review is still on the item page: %v", raw)
	}
	for _, e := range r.feed(bc) {
		if e.Kind == "review_published" {
			t.Fatalf("a hidden review is still in the feed: %+v", e)
		}
	}
	if n := r.notices(a, "content_hidden"); n != 1 {
		t.Fatalf("the review's author has %d content_hidden notices, want 1", n)
	}
	if n := r.notices(third, "report_resolved"); n != 1 {
		t.Fatalf("the reporter has %d report_resolved notices, want 1", n)
	}

	adminC.PostForm("/admin/reports/"+commentReport+"/resolve", url.Values{"outcome": {"content_deleted"}}).
		AssertRedirect("/admin/reports/" + commentReport)
	if n := r.count(`SELECT count(*) FROM comments WHERE id = $1`, cid); n != 0 {
		t.Fatal("comment not deleted")
	}
	if n := r.notices(b, "content_deleted"); n != 1 {
		t.Fatalf("the comment's author has %d content_deleted notices, want 1", n)
	}
}

func TestReportsReviewBanAuthor(t *testing.T) {
	r, a, _, _, _, _, _, _, rv, _ := socialSetup(t)
	_, cc := r.member("cc@example.com", "Third", false)
	admin := makeUser(t, r.app, adminEmail, "admin", true)

	cc.PostJSON("/reports", reportBody("review", rv))
	id := r.str(`SELECT id::text FROM reports WHERE target_type = 'review'`)

	browserFor(t, r.app, admin).PostForm("/admin/reports/"+id+"/resolve", url.Values{"outcome": {"user_banned"}}).
		AssertRedirect("/admin/reports/" + id)

	if !reload(t, r.app, a.ID).BannedAt.Valid {
		t.Fatal("the review's author is not banned")
	}
}

func TestReportsReviewCommentFlagOff(t *testing.T) {
	r := newSocialRig(t, false)
	a, _ := r.member("off-a@example.com", "Critic", false)
	_, bc := r.member("off-b@example.com", "Second", false)
	item := r.releasedItem("game")
	rv := uuid.NewV7().String()
	r.exec(`INSERT INTO reviews (id, user_id, item_id, body) VALUES ($1, $2, $3, 'Great game')`, rv, a.ID, item)

	bc.PostJSON("/reports", reportBody("review", rv)).AssertNotFound()
	bc.PostJSON("/reports", reportBody("comment", rv)).AssertNotFound()

	if n := r.count(`SELECT count(*) FROM reports`); n != 0 {
		t.Fatalf("%d reports filed with the social flag off", n)
	}

	// Other targets keep working.
	if res := bc.PostJSON("/reports", reportBody("user", "critic")); res.Status() >= 400 {
		t.Fatalf("a user report is refused with the flag off: %d", res.Status())
	}
}

func TestReportsUnhide(t *testing.T) {
	r, a, _, _, _, _, _, item, rv, _ := socialSetup(t)
	third, _ := r.member("cc@example.com", "Third", false)
	admin := makeUser(t, r.app, adminEmail, "admin", true)
	adminC := browserFor(t, r.app, admin)

	r.exec(`UPDATE reviews SET hidden = true WHERE id = $1`, rv)
	rid := uuid.NewV7().String()
	r.exec(
		`INSERT INTO reports (id, reporter_id, target_type, target_id, reason) VALUES ($1, $2, 'review', $3, 'spam')`,
		rid, third.ID, rv)

	// content_unhidden does not apply to a user report.
	uid := uuid.NewV7().String()
	r.exec(`INSERT INTO reports (id, reporter_id, target_type, target_id, reason) VALUES ($1, $2, 'user', $3, 'spam')`,
		uid, third.ID, strconv.FormatInt(a.ID, 10))
	adminC.PostForm("/admin/reports/"+uid+"/resolve", url.Values{"outcome": {"content_unhidden"}}).AssertUnprocessable()

	adminC.PostForm("/admin/reports/"+rid+"/resolve", url.Values{"outcome": {"content_unhidden"}}).
		AssertRedirect("/admin/reports/" + rid)

	if n := r.count(`SELECT count(*) FROM reviews WHERE id = $1 AND NOT hidden`, rv); n != 1 {
		t.Fatal("review still hidden")
	}
	if n := r.count(
		`SELECT count(*) FROM reports WHERE id = $1 AND status = 'resolved' AND outcome = 'content_unhidden'`, rid,
	); n != 1 {
		t.Fatal("report not closed with the unhide outcome")
	}
	page := r.app.Client(t).WithHeader("X-Inertia-Partial-Component", "Catalog/Show").
		WithHeader("X-Inertia-Partial-Data", "reviews").Inertia("").Get(r.itemPath(item)).AssertOK().
		AssertInertia("Catalog/Show")
	if raw, _ := page.Prop("reviews"); len(raw.(map[string]any)["data"].([]any)) != 1 {
		t.Fatalf("the restored review is not on the item page: %v", raw)
	}
}
