//go:build integration

package tests

import (
	"net/url"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Owner Q-A15 (SDD 04 FR-04-17): a moderator reviews but never applies or rejects for good, even with the
// trusted-moderators flag on; claim and release stay theirs.
func TestSecurity_A01_ModeratorCannotApply(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, true)
	seedModerationItem(t, app)
	const id = "00000000-0000-7000-8000-0000000000c1"
	seedSuggestion(t, app, id, "submitted", `{"title":{"before":"a","after":"b"}}`)

	mod := browserFor(t, app, makeUser(t, app, "r1-mod@example.com", models.RoleModerator, true))
	mod.PostForm("/admin/moderation/"+id+"/claim", url.Values{"version": {"1"}}).AssertRedirect("/admin/moderation")

	mod.PostForm("/admin/moderation/"+id+"/approve", url.Values{"version": {"2"}}).AssertForbidden()
	mod.PostForm("/admin/moderation/"+id+"/reject", url.Values{"version": {"2"}, "reason": {"spam"}}).AssertForbidden()
	for _, action := range []string{"approve", "reject"} {
		mod.PostForm("/admin/moderation/bulk", url.Values{"action": {action}, "ids": {id}, "reason": {"spam"}}).
			AssertForbidden()
	}
	mod.PostForm("/admin/moderation/bulk", url.Values{"action": {"release"}, "ids": {id}}).AssertOK()

	// The catalogue tools outside the queue (item merge, affiliate links) are panel-admin only: the panel's 404.
	const item = "00000000-0000-0000-0000-0000000000aa"
	mod.PostForm("/admin/items/"+item+"/merge",
		url.Values{"into": {"00000000-0000-0000-0000-0000000000bb"}, "confirm": {"merge"}}).AssertNotFound()
	mod.PostForm("/admin/items/"+item+"/affiliate-links",
		url.Values{"store": {"gog"}, "url": {"https://x.example/a"}}).AssertNotFound()
	mod.PostForm("/admin/items/"+item+"/affiliate-links/gog/delete", url.Values{}).AssertNotFound()

	if got := itemTitle(t, app); got != "a" {
		t.Fatalf("moderator wrote the catalogue: %q", got)
	}

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	var status string
	if err := db.QueryRowContext(t.Context(), `SELECT status::text FROM suggestions WHERE id = $1`, id).
		Scan(&status); err != nil || status != "submitted" {
		t.Fatalf("status = %q, %v", status, err)
	}

	// The pages do not offer what the gate refuses.
	for _, page := range []string{"/admin/moderation", "/admin/moderation/" + id} {
		body := mod.Get(page).AssertOK().Body()
		if strings.Contains(body, `/approve"`) || strings.Contains(body, `value="approve">Approve`) {
			t.Fatalf("%s offers approve to a moderator: %s", page, body)
		}
	}

	admin := browserFor(t, app, makeUser(t, app, "r1-admin@example.com", models.RoleAdmin, true))
	admin.PostForm("/admin/moderation/"+id+"/approve", url.Values{"version": {"3"}}).AssertRedirect("/admin/moderation")
	if got := itemTitle(t, app); got != "b" {
		t.Fatalf("admin approve did not apply: %q", got)
	}
}

func TestModerationRecommend(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, true)
	const id = "00000000-0000-7000-8000-0000000000c2"
	seedSuggestion(t, app, id, "submitted", `{"title.en":{"before":"a","after":"b"}}`)

	modUser := makeUser(t, app, "r2-mod@example.com", models.RoleModerator, true)
	mod := browserFor(t, app, modUser)
	path := "/admin/moderation/" + id + "/recommend"

	json := mod.WithHeader("Accept", "application/json")
	json.PostForm(path, url.Values{"version": {"1"}, "recommendation": {"maybe"}}).AssertStatus(422)
	json.PostForm(path, url.Values{
		"version": {"1"}, "recommendation": {"approve"},
		"note": {strings.Repeat("x", 1001)},
	}).AssertStatus(422)
	mod.PostForm(path, url.Values{"version": {"9"}, "recommendation": {"approve"}}).AssertStatus(409)

	mod.PostForm(path, url.Values{"version": {"1"}, "recommendation": {"approve"}, "note": {"Source checks out"}}).
		AssertRedirect("/admin/moderation/" + id)
	// A second recommendation replaces the reviewer's first one.
	mod.PostForm(path, url.Values{"version": {"1"}, "recommendation": {"reject"}, "note": {"<b>Wrong year</b>"}}).
		AssertRedirect("/admin/moderation/" + id)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	var (
		n        int
		rec, ver string
	)
	if err := db.QueryRowContext(t.Context(), `SELECT count(*), max(recommendation), max(suggestion_version)::text
FROM suggestion_reviews WHERE suggestion_id = $1 AND reviewer_id = $2`, id, modUser.ID).Scan(&n, &rec, &ver); err != nil {
		t.Fatal(err)
	}
	if n != 1 || rec != "reject" || ver != "1" {
		t.Fatalf("reviews = %d %q v%s", n, rec, ver)
	}

	// Recommending is not a transition: the version the admin acts on is unchanged.
	var version int
	if err := db.QueryRowContext(t.Context(), `SELECT version FROM suggestions WHERE id = $1`, id).
		Scan(&version); err != nil ||
		version != 1 {
		t.Fatalf("version = %d, %v", version, err)
	}

	admin := browserFor(t, app, makeUser(t, app, "r2-admin@example.com", models.RoleAdmin, true))
	detail := admin.Get("/admin/moderation/" + id).AssertOK().Body()
	for _, want := range []string{"Recommendations", "reject", "&lt;b&gt;Wrong year&lt;/b&gt;", "#" + itoa(modUser.ID)} {
		if !strings.Contains(detail, want) {
			t.Fatalf("detail missing %q: %s", want, detail)
		}
	}
	if !strings.Contains(admin.Get("/admin/moderation").AssertOK().Body(), "reject 1") {
		t.Fatal("queue does not show the recommendation")
	}

	// A decided suggestion takes no more recommendations.
	admin.PostForm("/admin/moderation/"+id+"/reject", url.Values{"version": {"1"}, "reason": {"incorrect"}}).
		AssertRedirect("/admin/moderation")
	mod.PostForm(path, url.Values{"version": {"2"}, "recommendation": {"approve"}}).AssertStatus(409)

	user := browserFor(t, app, makeUser(t, app, "r2-user@example.com", "user", true))
	user.PostForm(path, url.Values{"version": {"1"}, "recommendation": {"approve"}}).AssertForbidden()
	app.Client(t).PostForm(path, url.Values{"version": {"1"}, "recommendation": {"approve"}}).AssertRedirect("/login")
}

func TestSecurity_A01_ModeratorCannotApplyEntityCreate(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, true)
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}
	const id = "00000000-0000-7000-8000-0000000000c9"
	if _, err := db.ExecContext(t.Context(), `INSERT INTO suggestions
(id, kind, target_type, target_domain, locale, changes, status, version, submitted_at)
VALUES ($1, 'create', 'entity', 'movie', 'en', $2::jsonb, 'submitted', 1, now())`, id,
		`{"title":{"before":null,"after":"Mod Made"},"entity_kind":{"before":null,"after":"person"}}`); err != nil {
		t.Fatal(err)
	}
	entities := func() int {
		var n int
		if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM entities WHERE default_name = 'Mod Made'`).
			Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}

	mod := browserFor(t, app, makeUser(t, app, "r9-mod@example.com", models.RoleModerator, true))
	mod.PostForm("/admin/moderation/"+id+"/approve", url.Values{"version": {"1"}}).AssertForbidden()
	if n := entities(); n != 0 {
		t.Fatalf("moderator created %d entities", n)
	}

	admin := browserFor(t, app, makeUser(t, app, "r9-admin@example.com", models.RoleAdmin, true))
	admin.PostForm("/admin/moderation/"+id+"/approve", url.Values{"version": {"1"}}).AssertRedirect("/admin/moderation")
	if n := entities(); n != 1 {
		t.Fatalf("admin approve created %d entities", n)
	}
}
