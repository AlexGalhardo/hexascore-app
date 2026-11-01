//go:build integration

package tests

import (
	"net/url"
	"testing"
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
