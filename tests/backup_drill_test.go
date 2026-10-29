//go:build integration

package tests

import (
	"bytes"
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/crypt"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/backup"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// backupFixture is a migrated database with a runner whose dumps go to memory.
func backupFixture(t *testing.T) (*vukemtest.App, *backup.Runner, string) {
	t.Helper()

	for _, bin := range []string{"pg_dump", "pg_restore"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("%s not in PATH (the app image ships postgresql18-client)", bin)
		}
	}

	dsn := pgtest.Database(t, "")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"DATABASE_URL": dsn}})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	enc, err := crypt.New(bytes.Repeat([]byte{9}, 32))
	if err != nil {
		t.Fatal(err)
	}

	r := &backup.Runner{
		DB: db, Disk: storage.NewMemory(""), Enc: enc, DatabaseURL: dsn,
		PgDump: "pg_dump", PgRestore: "pg_restore", Env: "test", Now: time.Now,
	}

	return app, r, dsn
}

func TestBackupDrill(t *testing.T) {
	app, r, dsn := backupFixture(t)
	ctx := t.Context()

	newUser(t, app, "drill-a@example.com")
	newUser(t, app, "drill-b@example.com")

	res, err := r.Run(ctx, "daily")
	if err != nil {
		t.Fatal(err)
	}

	// Verify runs pg_restore --list with PATH as its only variable (FIX-backup-child-path).
	if err := r.Verify(ctx, "daily", res); err != nil {
		t.Fatalf("verify: %v", err)
	}

	// Rows written after the dump are not in it: the manifest still matches the restore.
	newUser(t, app, "drill-late@example.com")

	m, err := r.ReadManifest(ctx, res.Key)
	if err != nil {
		t.Fatal(err)
	}

	if m.Tables["users"] != 2 {
		t.Fatalf("manifest users = %d, want 2", m.Tables["users"])
	}

	asserts := func(ctx context.Context, db *database.DB) (int, error) {
		report, err := catalog.RunDQ(ctx, db, scoring.DQAsserts()...)

		return report.Errors, err
	}

	drill, err := r.Drill(ctx, dsn, asserts)
	if err != nil {
		t.Fatalf("drill: %v (diff %v)", err, drill.Diff)
	}

	if drill.Key != res.Key {
		t.Fatalf("drilled %s, want the latest daily %s", drill.Key, res.Key)
	}

	vukemtest.AssertDatabaseHas(t, app, "backup_runs", map[string]any{
		"id": drill.ID, "kind": "drill", "status": "verified", "object_key": res.Key,
	})

	db, _ := app.DB()
	if n := count(t, db, `SELECT count(*) FROM backup_runs WHERE kind = 'drill' AND duration_s < 3600`); n != 1 {
		t.Fatalf("drill rows within the budget = %d, want 1", n)
	}

	if n := count(t, db, `SELECT count(*) FROM pg_database WHERE datname = $1`, backup.DrillDatabase); n != 0 {
		t.Fatalf("%s was not dropped", backup.DrillDatabase)
	}

	// A failed assert fails the drill and is recorded.
	failing := func(context.Context, *database.DB) (int, error) { return 1, nil }
	if _, err := r.Drill(ctx, dsn, failing); err == nil {
		t.Fatal("a drill with an assert error passed")
	}

	if n := count(t, db, `SELECT count(*) FROM backup_runs WHERE kind = 'drill' AND status = 'failed'`); n != 1 {
		t.Fatalf("failed drill rows = %d, want 1", n)
	}
}

// SDD 13 §8 and open issue 11: deletions made after the dump are re-applied
// to the restored database, and the replay is idempotent.
func TestRestoreReappliesDeletions(t *testing.T) {
	app, r, _ := backupFixture(t)
	ctx := t.Context()

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	acc.Subscriptions = failingCanceller{}

	source, _ := app.DB()
	gone := newUser(t, app, "gone@example.com").ID
	later := newUser(t, app, "later@example.com").ID
	back := newUser(t, app, "back@example.com").ID
	kept := newUser(t, app, "kept@example.com").ID

	if _, err := acc.ScheduleDeletion(ctx, back); err != nil {
		t.Fatal(err)
	}

	res, err := r.Run(ctx, "daily")
	if err != nil {
		t.Fatal(err)
	}

	// After the dump: one account is anonymised, one scheduled, one cancelled.
	if _, err := acc.ScheduleDeletion(ctx, gone); err != nil {
		t.Fatal(err)
	}

	if _, err := source.ExecContext(ctx, `UPDATE users SET deletion_scheduled_at = $2 WHERE id = $1`,
		gone, time.Now().UTC().AddDate(0, 0, -31)); err != nil {
		t.Fatal(err)
	}

	if done, err := acc.PurgeDue(ctx); err != nil || len(done) != 1 {
		t.Fatalf("purged %v, %v", done, err)
	}

	if _, err := acc.ScheduleDeletion(ctx, later); err != nil {
		t.Fatal(err)
	}

	if err := acc.CancelDeletion(ctx, back); err != nil {
		t.Fatal(err)
	}

	targetURL := pgtest.Database(t, "")
	if err := r.Restore(ctx, res.Key, targetURL); err != nil {
		t.Fatal(err)
	}

	target, err := database.Open(ctx, targetURL, database.WithDriver(postgres.Driver()))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = target.Close() })

	// The restore alone brings the anonymised account back.
	if n := count(t, target, `SELECT count(*) FROM users WHERE email = 'gone@example.com'`); n != 1 {
		t.Fatalf("restored gone@ rows = %d, want 1", n)
	}

	stats, err := acc.ReplayDeletions(ctx, source, target)
	if err != nil {
		t.Fatal(err)
	}

	if want := (accounts.ReplayStats{Anonymised: 1, Scheduled: 1, Cancelled: 1}); stats != want {
		t.Fatalf("replay = %+v, want %+v", stats, want)
	}

	if n := count(t, target, `SELECT count(*) FROM users WHERE id = $1 AND anonymized_at IS NOT NULL
AND email = 'deleted-' || id || '@invalid' AND deletion_scheduled_at IS NULL`, gone); n != 1 {
		t.Fatal("gone@ is not anonymised in the restored database")
	}

	if n := count(t, target, `SELECT count(*) FROM users WHERE email LIKE '%@example.com' AND id <> $1`, gone); n != 3 {
		t.Fatalf("other users = %d, want 3", n)
	}

	var srcAt, dstAt time.Time

	_ = source.QueryRowContext(ctx, `SELECT deletion_scheduled_at FROM users WHERE id = $1`, later).Scan(&srcAt)
	if err := target.QueryRowContext(ctx, `SELECT deletion_scheduled_at FROM users WHERE id = $1`, later).
		Scan(&dstAt); err != nil || !dstAt.Equal(srcAt) {
		t.Fatalf("later@ scheduled at %v (%v), want %v", dstAt, err, srcAt)
	}

	if n := count(t, target, `SELECT count(*) FROM users WHERE id IN ($1, $2) AND deletion_scheduled_at IS NULL
AND anonymized_at IS NULL`, back, kept); n != 2 {
		t.Fatal("back@ (cancelled) or kept@ carries a deletion after the replay")
	}

	if again, err := acc.ReplayDeletions(ctx, source, target); err != nil || again != (accounts.ReplayStats{}) {
		t.Fatalf("second replay = %+v, %v; want no change", again, err)
	}
}
