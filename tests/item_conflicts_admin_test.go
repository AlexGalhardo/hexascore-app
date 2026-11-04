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
	conflictOwner = "00000000-0000-0000-0000-0000000000e1"
	conflictQID   = "Q777"
)

func seedConflict(t *testing.T, app *vukemtest.App) {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(
		t.Context(),
		`INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata', 'Wikidata', 'CC0-1.0', 'https://www.wikidata.org') ON CONFLICT DO NOTHING;
INSERT INTO items (id, type, slug, wikidata_qid, external_ids)
VALUES ('`+conflictOwner+`', 'movie', 'owner-film', 'Q11', '{"imdb":"tt1"}');
INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ('`+conflictOwner+`', 'en', 'Owner <Film>', 'wikidata');
INSERT INTO item_conflicts (wikidata_qid, item_type, kind, key, value, other_item_id)
VALUES ('`+conflictQID+`', 'movie', 'external_id', 'imdb', 'tt1', '`+conflictOwner+`');
INSERT INTO source_records (source, external_id, payload, hash, normalized_at)
VALUES ('wikidata', '`+conflictQID+`', '{}', encode(sha256('h'::bytea), 'hex'), now())`,
	); err != nil {
		t.Fatal(err)
	}
}

func TestItemConflicts_AdminListsAndAllows(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedConflict(t, app)
	admin := browserFor(t, app, makeUser(t, app, "conflicts-admin@example.com", models.RoleAdmin, true))
	plain := browserFor(t, app, makeUser(t, app, "conflicts-plain@example.com", "user", true))

	plain.Get("/admin/ingest/conflicts").AssertStatus(403) // like the other ingest queues (runIngestion)
	plain.PostForm("/admin/ingest/conflicts/1", url.Values{"action": {"allow"}}).AssertStatus(403)

	queue := inertiaProp(t, admin.Get("/admin/ingest/conflicts").AssertOK().Body(), "Admin/IngestConflicts", "conflicts")
	if !strings.Contains(queue, conflictQID) || !strings.Contains(queue, "tt1") ||
		!strings.Contains(queue, "Owner <Film>") {
		t.Fatalf("queue page: %s", queue)
	}

	admin.PostForm("/admin/ingest/conflicts/1", url.Values{"action": {"explode"}}).AssertStatus(422)
	admin.PostForm("/admin/ingest/conflicts/1", url.Values{"action": {"allow"}}).
		AssertRedirect("/admin/ingest/conflicts")

	vukemtest.AssertDatabaseHas(t, app, "item_conflicts", map[string]any{"id": 1, "status": "allowed"})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var pending bool
	if err := db.QueryRowContext(t.Context(), `SELECT normalized_at IS NULL FROM source_records WHERE external_id = $1`, conflictQID).
		Scan(&pending); err != nil ||
		!pending {
		t.Fatalf("allowing must put the record back for the normaliser: %v %v", pending, err)
	}

	admin.PostForm("/admin/ingest/conflicts/1", url.Values{"action": {"dismiss"}}).AssertStatus(409)
	admin.PostForm("/admin/ingest/conflicts/99", url.Values{"action": {"dismiss"}}).AssertNotFound()
}

func TestItemConflicts_DismissLeavesTheRecordAlone(t *testing.T) {
	t.Parallel()

	app := moderationApp(t, false)
	seedConflict(t, app)
	admin := browserFor(t, app, makeUser(t, app, "conflicts-dismiss@example.com", models.RoleAdmin, true))

	admin.PostForm("/admin/ingest/conflicts/1", url.Values{"action": {"dismiss"}}).
		AssertRedirect("/admin/ingest/conflicts")
	vukemtest.AssertDatabaseHas(t, app, "item_conflicts", map[string]any{"id": 1, "status": "dismissed"})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var pending bool
	if err := db.QueryRowContext(t.Context(), `SELECT normalized_at IS NULL FROM source_records WHERE external_id = $1`, conflictQID).
		Scan(&pending); err != nil ||
		pending {
		t.Fatalf("dismissing must not reopen the record: %v %v", pending, err)
	}

	if body := admin.Get("/admin/ingest/conflicts").AssertOK().Body(); strings.Contains(body, conflictQID) {
		t.Fatal("a resolved conflict is still listed")
	}
}
