//go:build integration

package tests

import (
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

func seedSuggestion(t *testing.T, app *vukemtest.App, id, status, changes string) {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), `INSERT INTO suggestions
(id, kind, target_type, target_id, target_domain, locale, changes, status, submitted_at)
VALUES ($1,'edit','item','00000000-0000-0000-0000-0000000000aa','movie','en',$2::jsonb,$3, now())`,
		id, changes, status); err != nil {
		t.Fatal(err)
	}
}

func TestSecurity_Community_A03_DiffEscapesHTML(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	const id = "00000000-0000-7000-8000-0000000000b1"
	seedSuggestion(t, app, id, "submitted",
		`{"title.en":{"before":"Old","after":"<script>alert(1)</script> New"}}`)

	admin := browserFor(t, app, makeUser(t, app, "d-admin@example.com", models.RoleAdmin, true))
	body := admin.Get("/admin/moderation/" + id).AssertOK().Body()

	if strings.Contains(body, "<script>alert(1)") {
		t.Fatalf("after value rendered as markup: %s", body)
	}

	if !strings.Contains(body, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("after value not rendered as text: %s", body)
	}
}

func TestModerationDetailActions(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, true)
	const id = "00000000-0000-7000-8000-0000000000b2"
	seedSuggestion(t, app, id, "submitted", `{"title.en":{"before":"a","after":"b"}}`)

	admin := browserFor(t, app, makeUser(t, app, "d2-admin@example.com", models.RoleAdmin, true))

	// Browser forms are redirected back with a flash; JSON clients see the 422.
	admin.WithHeader("Accept", "application/json").
		PostForm("/admin/moderation/"+id+"/reject", url.Values{"version": {"1"}}).AssertStatus(422)
	// The seeded target item does not exist, so the apply transaction refuses it.
	admin.PostForm("/admin/moderation/"+id+"/approve", url.Values{"version": {"1"}}).AssertStatus(409)
	admin.PostForm("/admin/moderation/"+id+"/reject", url.Values{"version": {"1"}, "reason": {"spam"}}).
		AssertRedirect("/admin/moderation")

	mod := browserFor(t, app, makeUser(t, app, "d2-mod@example.com", models.RoleModerator, true))
	mod.PostForm("/admin/moderation/"+id+"/clone", url.Values{}).AssertForbidden()
	// The seeded suggestion has no author, so there is nobody to own the draft.
	admin.PostForm("/admin/moderation/"+id+"/clone", url.Values{}).AssertStatus(409)
}

func TestModerationBulkLimit(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	admin := browserFor(t, app, makeUser(t, app, "d3-admin@example.com", models.RoleAdmin, true))

	ids := make([]string, 51)
	for i := range ids {
		ids[i] = "00000000-0000-7000-8000-" + strings.Repeat("0", 12-len(strconv.Itoa(i))) + strconv.Itoa(i)
	}

	admin.WithHeader("Accept", "application/json").
		PostForm("/admin/moderation/bulk", url.Values{"action": {"claim"}, "ids": ids}).AssertStatus(422)

	res := admin.PostForm("/admin/moderation/bulk", url.Values{"action": {"claim"}, "ids": ids[:2]}).AssertOK().Body()
	if strings.Count(res, "not-found") != 2 {
		t.Fatalf("per-id results missing: %s", res)
	}
}

func seedModerationItem(t *testing.T, app *vukemtest.App) {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), `INSERT INTO data_sources (code, name, licence, url)
VALUES ('wikidata','Wikidata','CC0-1.0','https://www.wikidata.org') ON CONFLICT DO NOTHING;
INSERT INTO items (id, type, slug, wikidata_qid, original_title, provenance)
VALUES ('00000000-0000-0000-0000-0000000000aa','movie','old','Q1','a','{"original_title":"wikidata"}')`); err != nil {
		t.Fatal(err)
	}
}

func itemTitle(t *testing.T, app *vukemtest.App) string {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var title string
	if err := db.QueryRowContext(t.Context(),
		`SELECT original_title FROM items WHERE id = '00000000-0000-0000-0000-0000000000aa'`).Scan(&title); err != nil {
		t.Fatal(err)
	}

	return title
}

func TestModerationApproveApplies(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedModerationItem(t, app)
	const id = "00000000-0000-7000-8000-0000000000b4"
	seedSuggestion(t, app, id, "submitted", `{"title":{"before":"a","after":"b"}}`)

	admin := browserFor(t, app, makeUser(t, app, "d4-admin@example.com", models.RoleAdmin, true))
	admin.PostForm("/admin/moderation/"+id+"/approve", url.Values{"version": {"1"}}).AssertRedirect("/admin/moderation")

	if got := itemTitle(t, app); got != "b" {
		t.Fatalf("title after approve = %q", got)
	}

	// A second approve with the old version is stale.
	admin.PostForm("/admin/moderation/"+id+"/approve", url.Values{"version": {"1"}}).AssertStatus(409)
}

func TestSecurity_Community_A01_ApproveNeedsModerate(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedModerationItem(t, app)
	const id = "00000000-0000-7000-8000-0000000000b5"
	seedSuggestion(t, app, id, "submitted", `{"title":{"before":"a","after":"b"}}`)

	for _, role := range []string{"user", models.RoleModerator} {
		c := browserFor(t, app, makeUser(t, app, "d5-"+role+"@example.com", role, true))
		c.PostForm("/admin/moderation/"+id+"/approve", url.Values{"version": {"1"}}).AssertForbidden()
		c.PostForm("/admin/moderation/bulk", url.Values{"action": {"approve"}, "ids": {id}}).AssertForbidden()
	}

	app.Client(t).PostForm("/admin/moderation/"+id+"/approve", url.Values{"version": {"1"}}).AssertRedirect("/login")

	if got := itemTitle(t, app); got != "a" {
		t.Fatalf("unauthorised approve wrote the catalogue: %q", got)
	}
}

func TestModerationMergeNeedsTypedTitle(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedModerationItem(t, app)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	const id = "00000000-0000-7000-8000-0000000000b6"
	if _, err := db.ExecContext(t.Context(), `INSERT INTO items (id, type, slug, wikidata_qid, original_title)
VALUES ('00000000-0000-0000-0000-0000000000bb','movie','keeper','Q2','Keeper');
INSERT INTO suggestions (id, kind, target_type, target_id, target_domain, locale, changes, status, submitted_at)
VALUES ('`+id+`','merge','item','00000000-0000-0000-0000-0000000000aa','movie','en',
 '{"*merge*":{"before":"00000000-0000-0000-0000-0000000000aa","after":"00000000-0000-0000-0000-0000000000bb"}}',
 'submitted', now())`); err != nil {
		t.Fatal(err)
	}

	admin := browserFor(t, app, makeUser(t, app, "d6-admin@example.com", models.RoleAdmin, true))
	admin.Get("/admin/moderation/" + id).AssertOK().AssertSee(`name="confirm_title"`)
	admin.WithHeader("Accept", "application/json").
		PostForm("/admin/moderation/"+id+"/approve", url.Values{"version": {"1"}}).AssertStatus(422)
	admin.PostForm("/admin/moderation/"+id+"/approve", url.Values{"version": {"1"}, "confirm_title": {"a"}}).
		AssertRedirect("/admin/moderation")

	var left int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM items`).Scan(&left); err != nil || left != 1 {
		t.Fatalf("items after merge = %d %v", left, err)
	}
}
