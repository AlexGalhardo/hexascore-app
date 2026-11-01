//go:build integration

package tests

import (
	"net/url"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

func TestModerationQueueClaimRelease(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, true)
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	const id = "00000000-0000-7000-8000-0000000000a1"
	if _, err := db.ExecContext(t.Context(), `INSERT INTO suggestions
(id, kind, target_type, target_id, target_domain, locale, changes, status, submitted_at)
VALUES ($1,'edit','item','00000000-0000-0000-0000-0000000000aa','movie','en','{"title":{"before":"a","after":"<b>x</b>"}}','submitted', now() - interval '13 hours')`, id); err != nil {
		t.Fatal(err)
	}

	admin := browserFor(t, app, makeUser(t, app, "q-admin@example.com", models.RoleAdmin, true))
	body := admin.Get("/admin/moderation").AssertOK().Body()
	for _, want := range []string{"badge-amber", "title", "/admin/moderation/" + id + "/claim", `name="version" value="1"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("queue page lacks %q: %s", want, body)
		}
	}

	if got := admin.Get("/admin/moderation?age=red").AssertOK().Body(); strings.Contains(got, id) {
		t.Fatal("age=red must hide a 13 h old suggestion")
	}

	admin.PostForm("/admin/moderation/"+id+"/claim", url.Values{"version": {"1"}}).
		AssertRedirect("/admin/moderation")

	// Same stale version again: 409 stale-before.
	admin.PostForm("/admin/moderation/"+id+"/claim", url.Values{"version": {"1"}}).AssertStatus(409)

	admin.PostForm("/admin/moderation/"+id+"/release", url.Values{"version": {"2"}}).
		AssertRedirect("/admin/moderation")
}

func TestSecurity_A01_ModerationClaimNeedsAbility(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	const id = "00000000-0000-7000-8000-0000000000a2"

	user := browserFor(t, app, makeUser(t, app, "q-user@example.com", "user", true))
	user.PostForm("/admin/moderation/"+id+"/claim", url.Values{"version": {"1"}}).AssertForbidden()

	mod := browserFor(t, app, makeUser(t, app, "q-mod@example.com", models.RoleModerator, true))
	mod.PostForm("/admin/moderation/"+id+"/claim", url.Values{"version": {"1"}}).AssertForbidden()
}
