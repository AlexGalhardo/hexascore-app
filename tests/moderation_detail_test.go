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
	admin.PostForm("/admin/moderation/"+id+"/approve", url.Values{"version": {"1"}}).AssertStatus(503)
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
