//go:build integration

package tests

import (
	"net/url"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func ingestAdminApp(t *testing.T, env map[string]string) *vukemtest.App {
	t.Helper()

	env["DATABASE_URL"] = pgtest.Database(t, "")
	env["FLAG_COMMUNITY_TRUSTED_MODERATORS"] = "true"

	return vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})
}

func runsOf(t *testing.T, app *vukemtest.App, where string) int {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var n int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM job_runs WHERE `+where).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

// P6.B.4.3: a trigger returns 303 and creates a run triggered_by admin:<id>; 409 while one is active, 423 for a
// disabled source, 422 for bad input (more than 500 qids, unknown job or mode).
func TestAdminIngestTrigger(t *testing.T) {
	t.Parallel()

	app := ingestAdminApp(t, map[string]string{"FLAG_INGEST_ENABLED": "true"})
	user := makeUser(t, app, "ingest-admin@example.com", models.RoleAdmin, true)
	admin := browserFor(t, app, user)

	admin.PostForm("/admin/ingest/runs", url.Values{"job": {"normalize"}, "mode": {"manual"}}).
		AssertRedirect("/admin/ingest")

	if n := runsOf(t, app, `job = 'normalize' AND triggered_by = 'admin:`+itoa(user.ID)+`'`); n != 1 {
		t.Fatalf("runs by the admin = %d, want 1", n)
	}

	// The test queue runs jobs inline, so an active run is seeded.
	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), `INSERT INTO job_runs (job, mode, status, triggered_by)
VALUES ('dq.asserts', 'manual', 'queued', 'test')`); err != nil {
		t.Fatal(err)
	}

	admin.PostForm("/admin/ingest/runs", url.Values{"job": {"dq.asserts"}, "mode": {"manual"}}).AssertStatus(409)

	// Source tmdb has its own kill switch, which is off.
	admin.PostForm("/admin/ingest/runs", url.Values{"job": {"fetch.tmdb"}, "source": {"tmdb"}, "mode": {"incremental"}}).
		AssertStatus(423)

	qids := make([]string, 501)
	for i := range qids {
		qids[i] = "Q" + itoa(int64(i+1))
	}

	admin.PostForm("/admin/ingest/runs", url.Values{
		"job": {"dq.asserts"}, "mode": {"manual"}, "qids": {strings.Join(qids, " ")},
	}).AssertStatus(422)
	admin.PostForm("/admin/ingest/runs", url.Values{"job": {"drop.table"}, "mode": {"manual"}}).AssertStatus(422)
	admin.PostForm("/admin/ingest/runs", url.Values{"job": {"dq.asserts"}, "mode": {"nope"}}).AssertStatus(422)
	admin.PostForm("/admin/ingest/runs", url.Values{"job": {"dq.asserts"}, "mode": {"manual"}, "limit": {"-1"}}).
		AssertStatus(422)
	admin.PostForm("/admin/ingest/runs", url.Values{"job": {"dq.asserts"}, "mode": {"manual"}, "allow_count_drop": {"on"}}).
		AssertStatus(422)

	if n := runsOf(t, app, `true`); n != 2 {
		t.Fatalf("rejected triggers left runs behind: %d", n)
	}
}

func TestAdminIngestDisabledMaster(t *testing.T) {
	t.Parallel()

	app := ingestAdminApp(t, map[string]string{})
	admin := browserFor(t, app, makeUser(t, app, "ingest-off@example.com", models.RoleAdmin, true))

	admin.PostForm("/admin/ingest/runs", url.Values{"job": {"normalize"}, "mode": {"manual"}}).AssertStatus(423)

	if n := runsOf(t, app, `true`); n != 0 {
		t.Fatalf("runs = %d, want 0", n)
	}
}

func TestAdminIngestCancelAndPage(t *testing.T) {
	t.Parallel()

	app := ingestAdminApp(t, map[string]string{})
	admin := browserFor(t, app, makeUser(t, app, "ingest-cancel@example.com", models.RoleAdmin, true))

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), `INSERT INTO job_runs (job, mode, status, triggered_by, error) VALUES
 ('normalize', 'manual', 'queued', 'test', NULL),
 ('dq.asserts', 'manual', 'failed', 'test', '<script>x</script>')`); err != nil {
		t.Fatal(err)
	}

	body := admin.Get("/admin/ingest").AssertOK().Body()
	if !strings.Contains(body, "normalize") || !strings.Contains(body, "dq.asserts") {
		t.Fatalf("page misses the runs: %s", body)
	}

	if strings.Contains(body, "<script>x</script>") {
		t.Fatalf("run error not escaped: %s", body)
	}

	admin.PostForm("/admin/ingest/runs/1/cancel", url.Values{}).AssertRedirect("/admin/ingest")

	if n := runsOf(t, app, `id = 1 AND status = 'cancelled' AND finished_at IS NOT NULL`); n != 1 {
		t.Fatal("run not cancelled")
	}

	admin.PostForm("/admin/ingest/runs/2/cancel", url.Values{}).AssertStatus(409)
	admin.PostForm("/admin/ingest/runs/999/cancel", url.Values{}).AssertNotFound()
	admin.PostForm("/admin/ingest/runs/abc/cancel", url.Values{}).AssertNotFound()
}

// The read-only panel resources exist for admins.
func TestAdminIngestPanelResources(t *testing.T) {
	t.Parallel()

	app := ingestAdminApp(t, map[string]string{})
	admin := browserFor(t, app, makeUser(t, app, "ingest-panel@example.com", models.RoleAdmin, true))

	for _, p := range []string{"/admin/job-runs", "/admin/ingest-failures", "/admin/dq-results"} {
		admin.Get(p).AssertOK()
	}
}

func TestSecurity_A01_IngestRunsAdminOnly(t *testing.T) {
	t.Parallel()

	app := ingestAdminApp(t, map[string]string{"FLAG_INGEST_ENABLED": "true"})
	mod := browserFor(t, app, makeUser(t, app, "ingest-mod@example.com", models.RoleModerator, true))
	member := browserFor(t, app, makeUser(t, app, "ingest-member@example.com", "user", true))

	for _, c := range []*vukemtest.Client{mod, member} {
		c.Get("/admin/ingest").AssertForbidden()
		c.PostForm("/admin/ingest/runs", url.Values{"job": {"normalize"}, "mode": {"manual"}}).AssertForbidden()
		c.PostForm("/admin/ingest/runs/1/cancel", url.Values{}).AssertForbidden()
	}

	if n := runsOf(t, app, `true`); n != 0 {
		t.Fatalf("a non-admin created runs: %d", n)
	}
}

func TestSecurity_A01_IngestRunsRequireCSRF(t *testing.T) {
	t.Parallel()

	app := ingestAdminApp(t, map[string]string{"FLAG_INGEST_ENABLED": "true"})
	admin := browserFor(t, app, makeUser(t, app, "ingest-csrf@example.com", models.RoleAdmin, true))

	admin.WithHeader("Sec-Fetch-Site", "cross-site").
		PostForm("/admin/ingest/runs", url.Values{"job": {"normalize"}, "mode": {"manual"}}).AssertForbidden()

	if n := runsOf(t, app, `true`); n != 0 {
		t.Fatalf("a cross-site post created a run: %d", n)
	}
}
