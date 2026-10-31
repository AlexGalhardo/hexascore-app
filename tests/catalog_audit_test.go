//go:build integration

package tests

import (
	"net/url"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

func firstAuditID(t *testing.T, app *vukemtest.App) string {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var id string
	if err := db.QueryRowContext(t.Context(), `SELECT min(id)::text FROM catalog_audit`).Scan(&id); err != nil {
		t.Fatal(err)
	}

	return id
}

// The audit log and revert stay admin-only, also for moderators with the trusted flag on (SDD 04 section 6, Q-A15).
func TestSecurity_A01_CatalogAuditAdminOnly(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, true)
	seedModerationItem(t, app)
	const id = "00000000-0000-7000-8000-0000000000d1"
	seedSuggestion(t, app, id, "submitted", `{"title":{"before":"a","after":"b"}}`)
	admin := browserFor(t, app, makeUser(t, app, "a1-admin@example.com", models.RoleAdmin, true))
	admin.PostForm("/admin/moderation/"+id+"/approve", url.Values{"version": {"1"}}).AssertRedirect("/admin/moderation")
	revert := "/admin/catalog-audit/" + firstAuditID(t, app) + "/revert"
	form := url.Values{"mode": {"single"}, "reason": {"incorrect"}}

	for _, role := range []string{"user", models.RoleModerator} {
		c := browserFor(t, app, makeUser(t, app, "a1-"+role+"@example.com", role, true))
		c.Get("/admin/catalog-audit").AssertForbidden()
		c.PostForm(revert, form).AssertForbidden()
	}

	app.Client(t).Get("/admin/catalog-audit").AssertRedirect("/login")

	if got := itemTitle(t, app); got != "b" {
		t.Fatalf("unauthorised revert wrote the catalogue: %q", got)
	}
}

func TestCatalogAuditRevert(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedModerationItem(t, app)
	const id = "00000000-0000-7000-8000-0000000000d2"
	seedSuggestion(t, app, id, "submitted", `{"title":{"before":"a","after":"<i>b</i>"}}`)
	admin := browserFor(t, app, makeUser(t, app, "a2-admin@example.com", models.RoleAdmin, true))
	admin.PostForm("/admin/moderation/"+id+"/approve", url.Values{"version": {"1"}}).AssertRedirect("/admin/moderation")

	body := admin.Get("/admin/catalog-audit?source=community").AssertOK().Body()
	auditID := firstAuditID(t, app)
	revert := "/admin/catalog-audit/" + auditID + "/revert"
	if !strings.Contains(body, `action="`+revert+`"`) || strings.Contains(body, "<i>b</i>") {
		t.Fatalf("audit list lacks the revert form or renders values as markup: %s", body)
	}

	admin.WithHeader("Accept", "application/json").
		PostForm(revert, url.Values{"mode": {"sideways"}, "reason": {"incorrect"}}).AssertStatus(422)
	admin.PostForm(revert, url.Values{"mode": {"single"}, "reason": {"incorrect"}}).
		AssertRedirect("/admin/catalog-audit")

	if got := itemTitle(t, app); got != "a" {
		t.Fatalf("title after revert = %q", got)
	}

	// The revert row has no revert form and the reverted row says so; a second revert is a conflict.
	body = admin.Get("/admin/catalog-audit").AssertOK().Body()
	if !strings.Contains(body, "reverted") || strings.Count(body, `/revert"`) != 0 {
		t.Fatalf("audit list after revert: %s", body)
	}
	admin.PostForm(revert, url.Values{"mode": {"single"}, "reason": {"incorrect"}}).AssertStatus(409)
	admin.PostForm("/admin/catalog-audit/999999/revert", url.Values{"mode": {"single"}, "reason": {"incorrect"}}).
		AssertNotFound()
}
