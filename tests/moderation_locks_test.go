//go:build integration

package tests

import (
	"net/url"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

const lockItem = "00000000-0000-0000-0000-0000000000aa"

// SDD 04 §6: admins lock a field (or a whole record) with a reason and optional expiry, and unlock it.
func TestLocks(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedModerationItem(t, app)
	admin := browserFor(t, app, makeUser(t, app, "locks-admin@example.com", models.RoleAdmin, true))

	admin.PostForm("/admin/locks", url.Values{
		"target": {"item"}, "target_id": {lockItem}, "field": {"title"}, "reason": {"edit war <b>"}, "days": {"7"},
	}).AssertRedirect("/admin/locks")
	admin.PostForm("/admin/locks", url.Values{
		"target": {"item"}, "target_id": {lockItem}, "field": {"title"}, "reason": {"again"},
	}).AssertStatus(409)
	admin.PostForm("/admin/locks", url.Values{
		"target": {"item"}, "target_id": {"00000000-0000-0000-0000-0000000000ff"}, "reason": {"x"},
	}).AssertNotFound()
	admin.PostForm("/admin/locks", url.Values{"target": {"item"}, "target_id": {lockItem}}).AssertStatus(303)

	body := admin.Get("/admin/locks").AssertOK().Body()
	if !strings.Contains(body, "edit war &lt;b&gt;") || strings.Contains(body, "edit war <b>") {
		t.Fatalf("lock reason not listed escaped: %s", body)
	}

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var id string
	var expires bool
	if err := db.QueryRowContext(t.Context(), `SELECT id::text, expires_at IS NOT NULL FROM catalog_locks
WHERE target_id = $1 AND field = 'title'`, lockItem).Scan(&id, &expires); err != nil || !expires {
		t.Fatalf("lock = %q expires=%v, %v", id, expires, err)
	}

	// A locked field refuses an apply (FR-04-7).
	const sid = "00000000-0000-7000-8000-0000000000d1"
	seedSuggestion(t, app, sid, "submitted", `{"title":{"before":"a","after":"b"}}`)
	admin.PostForm("/admin/moderation/"+sid+"/approve", url.Values{"version": {"1"}}).AssertStatus(423)

	admin.PostForm("/admin/locks/"+id+"/delete", url.Values{}).AssertRedirect("/admin/locks")

	var n int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM catalog_locks`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("locks left = %d, %v", n, err)
	}
}

// Locks and the audit export are admin-only (lockCatalog, viewCatalogAudit), even for a trusted moderator.
func TestSecurity_A01_LocksAndExportAdminOnly(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, true)
	seedModerationItem(t, app)
	mod := browserFor(t, app, makeUser(t, app, "locks-mod@example.com", models.RoleModerator, true))

	mod.Get("/admin/locks").AssertForbidden()
	mod.PostForm("/admin/locks", url.Values{
		"target": {"item"}, "target_id": {lockItem}, "field": {"title"}, "reason": {"x"},
	}).AssertForbidden()
	mod.PostForm("/admin/locks/00000000-0000-7000-8000-0000000000e1/delete", url.Values{}).AssertForbidden()
	mod.Get("/admin/catalog-audit/export.csv").AssertForbidden()
	mod.Get("/admin/catalog-audit/export.ndjson").AssertForbidden()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var n int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM catalog_locks`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("moderator created a lock: %d, %v", n, err)
	}
}

// FR-04-15: the export streams every matching row (100k here) with the CSV formula guard on.
func TestAuditExportStreams(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	admin := browserFor(t, app, makeUser(t, app, "export-admin@example.com", models.RoleAdmin, true))

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), `INSERT INTO catalog_audit
(actor_role, source, target_type, target_id, field, before, after)
SELECT 'admin', 'admin', 'item', '00000000-0000-0000-0000-0000000000aa',
  CASE WHEN g = 1 THEN '=HYPERLINK("x")' ELSE 'title' END, '"a"', to_jsonb(g::text)
FROM generate_series(1, 100000) g`); err != nil {
		t.Fatal(err)
	}

	csv := admin.Get("/admin/catalog-audit/export.csv").AssertOK()
	if !strings.Contains(csv.Header("Content-Type"), "text/csv") {
		t.Fatalf("content type = %q", csv.Header("Content-Type"))
	}

	lines := strings.Split(strings.TrimRight(csv.Body(), "\r\n"), "\n")
	if len(lines) != 100001 {
		t.Fatalf("csv lines = %d", len(lines))
	}

	if !strings.Contains(csv.Body(), `'=HYPERLINK`) {
		t.Fatal("formula cell not guarded")
	}

	nd := admin.Get("/admin/catalog-audit/export.ndjson?source=admin&field=title").AssertOK()
	if n := strings.Count(nd.Body(), "\n"); n != 99999 {
		t.Fatalf("ndjson lines = %d", n)
	}
}
