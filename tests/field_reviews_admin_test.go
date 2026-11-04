//go:build integration

package tests

import (
	"net/url"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

const fieldReviewItem = "00000000-0000-0000-0000-0000000000d1"

func seedFieldReview(t *testing.T, app *vukemtest.App) {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), `INSERT INTO data_sources (code, name, licence, url)
VALUES ('llm','LLM','proprietary','https://x') ON CONFLICT DO NOTHING;
INSERT INTO items (id, type, slug, wikidata_qid, runtime_minutes, provenance)
VALUES ('`+fieldReviewItem+`','movie','field-review-movie','Q11',90,'{"runtime_minutes":"wikidata"}');
INSERT INTO field_reviews (item_id, field, value, source, confidence, reason)
VALUES ('`+fieldReviewItem+`','runtime_minutes','123','llm',50,'<b>llm</b> guess')`); err != nil {
		t.Fatal(err)
	}
}

func TestFieldReviews_AdminResolvesAndItemUpdates(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedFieldReview(t, app)
	admin := browserFor(t, app, makeUser(t, app, "field-reviews-admin@example.com", models.RoleAdmin, true))

	queue := inertiaProp(t, admin.Get("/admin/ingest/field-reviews").AssertOK().Body(), "Admin/IngestFieldReviews", "reviews")
	if !strings.Contains(queue, "runtime_minutes") || !strings.Contains(queue, "<b>llm</b>") {
		t.Fatalf("queue page: %s", queue)
	}

	admin.PostForm("/admin/ingest/field-reviews/1", url.Values{"action": {"explode"}}).AssertStatus(422)
	admin.PostForm("/admin/ingest/field-reviews/1", url.Values{"action": {"accept"}}).
		AssertRedirect("/admin/ingest/field-reviews")

	vukemtest.AssertDatabaseHas(t, app, "items", map[string]any{"id": fieldReviewItem, "runtime_minutes": 123})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var src string

	var conf int

	if err := db.QueryRowContext(t.Context(), `SELECT provenance ->> 'runtime_minutes',
  (provenance_v2 -> 'runtime_minutes' ->> 'confidence')::int FROM items WHERE id = $1`, fieldReviewItem).
		Scan(&src, &conf); err != nil || src != "llm" || conf != 100 {
		t.Fatalf("provenance after accept: %q %d %v", src, conf, err)
	}

	admin.PostForm("/admin/ingest/field-reviews/1", url.Values{"action": {"reject"}}).AssertStatus(409)
	admin.PostForm("/admin/ingest/field-reviews/99", url.Values{"action": {"reject"}}).AssertNotFound()
}

func TestFieldReviews_RejectLeavesItem(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedFieldReview(t, app)
	admin := browserFor(t, app, makeUser(t, app, "field-reviews-reject@example.com", models.RoleAdmin, true))

	admin.PostForm("/admin/ingest/field-reviews/1", url.Values{"action": {"reject"}}).
		AssertRedirect("/admin/ingest/field-reviews")
	vukemtest.AssertDatabaseHas(t, app, "items", map[string]any{"id": fieldReviewItem, "runtime_minutes": 90})
}

// Only admins hold runIngestion: a moderator and a member are refused and the item does not change.
func TestSecurity_A01_FieldReviewsAdminOnly(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, true)
	seedFieldReview(t, app)
	mod := browserFor(t, app, makeUser(t, app, "fr-mod@example.com", models.RoleModerator, true))
	member := browserFor(t, app, makeUser(t, app, "fr-member@example.com", "user", true))

	for _, c := range []*vukemtest.Client{mod, member} {
		c.Get("/admin/ingest/field-reviews").AssertForbidden()
		c.PostForm("/admin/ingest/field-reviews/1", url.Values{"action": {"accept"}}).AssertForbidden()
	}

	app.Client(t).Get("/admin/ingest/field-reviews").AssertStatus(302)
	vukemtest.AssertDatabaseHas(t, app, "items", map[string]any{"id": fieldReviewItem, "runtime_minutes": 90})

	admin := browserFor(t, app, makeUser(t, app, "fr-csrf@example.com", models.RoleAdmin, true))
	admin.WithHeader("Sec-Fetch-Site", "cross-site").
		PostForm("/admin/ingest/field-reviews/1", url.Values{"action": {"accept"}}).AssertForbidden()
	vukemtest.AssertDatabaseHas(t, app, "items", map[string]any{"id": fieldReviewItem, "runtime_minutes": 90})
}
