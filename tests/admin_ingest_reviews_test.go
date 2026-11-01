//go:build integration

package tests

import (
	"net/url"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

const (
	reviewItem   = "00000000-0000-0000-0000-0000000000b1"
	reviewEntity = "00000000-0000-0000-0000-0000000000c1"
	reviewOther  = "00000000-0000-0000-0000-0000000000c2"
)

// seedReviews makes one movie, two person entities and three open reviews (ids 1..3) for it.
func seedReviews(t *testing.T, app *vukemtest.App) {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), `INSERT INTO data_sources (code, name, licence, url)
VALUES ('wikidata','Wikidata','CC0-1.0','https://www.wikidata.org') ON CONFLICT DO NOTHING;
INSERT INTO items (id, type, slug, wikidata_qid, original_title, provenance)
VALUES ('`+reviewItem+`','movie','review-movie','Q10','Review movie','{"original_title":"wikidata"}');
INSERT INTO entities (id, kind, slug, default_name) VALUES
 ('`+reviewEntity+`','person','jane-doe','Jane Doe'),
 ('`+reviewOther+`','person','jane-doe-2','Jane <b>Doe</b>');
INSERT INTO entity_match_reviews (item_id, role, source, raw_name, raw_year, raw_external_id, candidates) VALUES
 ('`+reviewItem+`','director','wikidata','Jane Doe',1970,'Q777','{`+reviewEntity+`,`+reviewOther+`}'),
 ('`+reviewItem+`','cast','wikidata','John Smith',NULL,NULL,'{`+reviewEntity+`}'),
 ('`+reviewItem+`','writer','wikidata','Third Person',NULL,NULL,'{`+reviewEntity+`}')`); err != nil {
		t.Fatal(err)
	}
}

type reviewState struct {
	Status   string
	Resolved bool
	Linked   int
}

func reviewOf(t *testing.T, app *vukemtest.App, id int, role string) reviewState {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var s reviewState
	if err := db.QueryRowContext(t.Context(), `SELECT r.status, r.resolved_by IS NOT NULL AND r.resolved_at IS NOT NULL,
  (SELECT count(*) FROM item_entities ie WHERE ie.item_id = r.item_id AND ie.role = r.role)
FROM entity_match_reviews r WHERE r.id = $1 AND r.role = $2`, id, role).Scan(&s.Status, &s.Resolved, &s.Linked); err != nil {
		t.Fatalf("review %d: %v", id, err)
	}

	return s
}

func reviewPath(id int) string { return "/admin/ingest/reviews/" + itoa(int64(id)) }

// P6.B.3.3: the queue lists open reviews escaped; link, create and dismiss close one and a second try answers 409.
func TestAdminIngestReviews(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedReviews(t, app)
	admin := browserFor(t, app, makeUser(t, app, "reviews-admin@example.com", models.RoleAdmin, true))

	body := admin.Get("/admin/ingest/reviews").AssertOK().Body()
	for _, want := range []string{"Jane Doe", "John Smith", "Q777", "Jane &lt;b&gt;Doe&lt;/b&gt;"} {
		if !strings.Contains(body, want) {
			t.Fatalf("queue misses %q: %s", want, body)
		}
	}

	if strings.Contains(body, "<b>Doe</b>") {
		t.Fatalf("candidate name not escaped: %s", body)
	}

	admin.PostForm(reviewPath(1), url.Values{"action": {"link"}, "entity_id": {reviewEntity}}).
		AssertRedirect("/admin/ingest/reviews")

	if got := reviewOf(t, app, 1, "director"); got != (reviewState{"linked", true, 1}) {
		t.Fatalf("after link = %+v", got)
	}

	vukemtest.AssertDatabaseHas(t, app, "item_entities", map[string]any{
		"item_id": reviewItem, "entity_id": reviewEntity, "role": "director", "source": "wikidata",
	})
	// The QID the credit carried is taught to the entity that did not have one.
	vukemtest.AssertDatabaseHas(t, app, "entities", map[string]any{"id": reviewEntity, "wikidata_qid": "Q777"})

	admin.PostForm(reviewPath(2), url.Values{"action": {"create"}}).AssertRedirect("/admin/ingest/reviews")

	if got := reviewOf(t, app, 2, "cast"); got != (reviewState{"created", true, 1}) {
		t.Fatalf("after create = %+v", got)
	}

	vukemtest.AssertDatabaseHas(t, app, "entities", map[string]any{"kind": "person", "default_name": "John Smith"})

	admin.PostForm(reviewPath(3), url.Values{"action": {"dismiss"}}).AssertRedirect("/admin/ingest/reviews")

	if got := reviewOf(t, app, 3, "writer"); got != (reviewState{"dismissed", true, 0}) {
		t.Fatalf("after dismiss = %+v", got)
	}

	// A resolved review is final.
	admin.PostForm(reviewPath(1), url.Values{"action": {"dismiss"}}).AssertStatus(409)

	if strings.Contains(admin.Get("/admin/ingest/reviews").AssertOK().Body(), "John Smith") {
		t.Fatal("resolved review still listed")
	}
}

func TestAdminIngestReviewsInvalidInput(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedReviews(t, app)
	admin := browserFor(t, app, makeUser(t, app, "reviews-invalid@example.com", models.RoleAdmin, true))

	admin.PostForm(reviewPath(1), url.Values{"action": {"explode"}}).AssertStatus(422)
	admin.PostForm(reviewPath(1), url.Values{"action": {"link"}}).AssertStatus(422)
	admin.PostForm(reviewPath(1), url.Values{"action": {"link"}, "entity_id": {"not-a-uuid"}}).AssertStatus(422)
	admin.PostForm(reviewPath(1), url.Values{
		"action": {"link"}, "entity_id": {"00000000-0000-0000-0000-0000000000ff"},
	}).AssertStatus(422)
	admin.PostForm(reviewPath(9999), url.Values{"action": {"dismiss"}}).AssertNotFound()
	admin.PostForm("/admin/ingest/reviews/abc", url.Values{"action": {"dismiss"}}).AssertNotFound()

	if got := reviewOf(t, app, 1, "director"); got.Status != "open" || got.Resolved {
		t.Fatalf("invalid input changed the review: %+v", got)
	}
}

// Only admins hold runIngestion, so even a trusted moderator is refused (403) and nothing changes.
func TestSecurity_A01_IngestReviewsAdminOnly(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, true)
	seedReviews(t, app)
	mod := browserFor(t, app, makeUser(t, app, "reviews-mod@example.com", models.RoleModerator, true))
	member := browserFor(t, app, makeUser(t, app, "reviews-member@example.com", "user", true))

	for _, c := range []*vukemtest.Client{mod, member} {
		c.Get("/admin/ingest/reviews").AssertForbidden()
		c.PostForm(reviewPath(1), url.Values{"action": {"link"}, "entity_id": {reviewEntity}}).AssertForbidden()
		c.PostForm(reviewPath(2), url.Values{"action": {"create"}}).AssertForbidden()
		c.PostForm(reviewPath(3), url.Values{"action": {"dismiss"}}).AssertForbidden()
	}

	app.Client(t).Get("/admin/ingest/reviews").AssertStatus(302)

	for i, role := range []string{"director", "cast", "writer"} {
		if got := reviewOf(t, app, i+1, role); got.Status != "open" || got.Resolved || got.Linked != 0 {
			t.Fatalf("review %d changed by a non-admin: %+v", i+1, got)
		}
	}
}

// A cross-site POST never reaches the handler.
func TestSecurity_A01_IngestReviewsRequireCSRF(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedReviews(t, app)
	admin := browserFor(t, app, makeUser(t, app, "reviews-csrf@example.com", models.RoleAdmin, true))

	admin.WithHeader("Sec-Fetch-Site", "cross-site").PostForm(reviewPath(3), url.Values{"action": {"dismiss"}}).
		AssertForbidden()

	if got := reviewOf(t, app, 3, "writer"); got.Status != "open" {
		t.Fatalf("cross-site post resolved the review: %+v", got)
	}
}
