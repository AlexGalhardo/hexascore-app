//go:build integration

package tests

import (
	"encoding/json"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/admin"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// ADR 0026: the React dashboard sits behind the panel guard. Non-admins see 404 (the panel's existence stays hidden)
// and an admin gets the snapshot as JSON with the headline counts and the recent runs.
func TestAdminOverview(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
	})

	member := browserFor(t, app, makeUser(t, app, "overview-member@example.com", models.RoleModerator, true))
	member.Get("/admin/overview").AssertStatus(404)
	member.Get("/admin/overview/data").AssertStatus(404)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(
		t.Context(),
		`INSERT INTO job_runs (job, mode, status, triggered_by, started_at, finished_at)
VALUES ('fetch.wikidata', 'bootstrap', 'failed', 'test', now() - interval '90 seconds', now() - interval '30 seconds')`,
	); err != nil {
		t.Fatal(err)
	}

	owner := browserFor(t, app, makeUser(t, app, "overview-admin@example.com", models.RoleAdmin, true))
	owner.Get("/admin/overview").AssertOK()

	var data admin.OverviewData
	if err := json.Unmarshal([]byte(owner.Get("/admin/overview/data").AssertOK().Body()), &data); err != nil {
		t.Fatal(err)
	}

	if data.Counts.FailedRuns24h != 1 {
		t.Fatalf("failed runs in 24h = %d, want 1", data.Counts.FailedRuns24h)
	}

	if len(data.Runs) != 1 || data.Runs[0].Job != "fetch.wikidata" || data.Runs[0].DurationSec == nil ||
		*data.Runs[0].DurationSec != 60 {
		t.Fatalf("runs = %+v, want one fetch.wikidata run of 60 s", data.Runs)
	}

	if data.Worker.Status != "never" {
		t.Fatalf("worker status = %q, want never", data.Worker.Status)
	}
}
