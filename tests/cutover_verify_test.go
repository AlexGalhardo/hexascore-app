//go:build integration

package tests

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

// importedFixture runs every stage, finalise included, over the synthetic
// legacy dump and returns the app (stdout captured), its database and a disk
// holding every cover key the import wrote.
func importedFixture(t *testing.T) (*vukemtest.App, *database.DB, *storage.Memory, *bytes.Buffer) {
	t.Helper()

	var out bytes.Buffer

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env:        map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
		AppOptions: []vukem.Option{vukem.WithStdout(&out)},
	})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()

	fixture, err := os.ReadFile("../database/legacyimport/testdata/fixture.sql")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, string(fixture)); err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	stages := legacyimport.Stages()
	for i := range stages {
		if stages[i].Name == "twofactor" {
			// The secret shape matches Better Auth's; the export covers both legacy two_factor rows.
			export := `{"userId":"u02","secret":"q8Zr1LmN4pXv7TbW2yHc9KdF3sJg6EaU","backupCodes":["aB3dE-fG4hI"]}
{"userId":"u04","secret":"q8Zr1LmN4pXv7TbW2yHc9KdF3sJg6EaU","backupCodes":[]}
`
			stages[i].Batch = legacyimport.TwoFactorBatch(strings.NewReader(export), time.Now)
		}
	}

	r := &legacyimport.Runner{DB: db, Batch: 2, Out: io.Discard, Now: time.Now}
	if err := r.Run(ctx, stages); err != nil {
		t.Fatalf("import: %v", err)
	}

	disk := storage.NewMemory("")

	rows, err := db.QueryContext(
		ctx,
		`SELECT k FROM (SELECT cover_key k FROM items UNION ALL SELECT cover_thumb_key FROM items) c
WHERE k IS NOT NULL`,
	)
	if err != nil {
		t.Fatal(err)
	}

	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			t.Fatal(err)
		}

		if err := disk.Put(ctx, k, strings.NewReader("webp")); err != nil {
			t.Fatal(err)
		}
	}

	_ = rows.Close()

	return app, db, disk, &out
}

func TestCutoverVerify_PassesOnFixtureImport(t *testing.T) {
	t.Parallel()

	app, db, disk, out := importedFixture(t)
	ctx := t.Context()

	var scored int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM items WHERE score_computed_at IS NOT NULL`).
		Scan(&scored); err != nil {
		t.Fatal(err)
	}

	if scored == 0 {
		t.Fatal("finalise recomputed no item")
	}

	var report bytes.Buffer
	if err := legacyimport.Verify(
		ctx,
		db,
		legacyimport.VerifyOptions{SamplePct: 100, Disk: disk, Out: &report},
	); err != nil {
		t.Fatalf("verify: %v\n%s", err, report.String())
	}

	t.Log(report.String())

	start := time.Now()
	if err := app.Call(ctx, "hexascore:cutover-verify", []string{"--counts-only"}); err != nil {
		t.Fatalf("--counts-only: %v\n%s", err, out.String())
	}

	if d := time.Since(start); d > time.Minute {
		t.Fatalf("--counts-only took %v, want < 60 s", d)
	}

	if !strings.Contains(out.String(), "all checks passed") {
		t.Fatalf("command output:\n%s", out.String())
	}
}

func TestCutoverVerify_DetectsSeededMismatch(t *testing.T) {
	t.Parallel()

	_, db, disk, _ := importedFixture(t)
	ctx := t.Context()

	var cover string
	if err := db.QueryRowContext(ctx, `SELECT cover_key FROM items WHERE cover_key IS NOT NULL ORDER BY id LIMIT 1`).
		Scan(&cover); err != nil {
		t.Fatal(err)
	}

	if err := disk.Delete(ctx, cover); err != nil {
		t.Fatal(err)
	}

	err := db.Tx(ctx, func(tx *database.Tx) error {
		if _, err := tx.ExecContext(ctx, `UPDATE items SET slug = slug || '-changed'
WHERE id = (SELECT min(id::text)::uuid FROM items)`); err != nil {
			return err
		}

		// Replica mode skips the FK triggers, the way a careless bulk load would.
		if _, err := tx.ExecContext(ctx, `SET LOCAL session_replication_role = replica`); err != nil {
			return err
		}

		_, err := tx.ExecContext(ctx, `UPDATE item_status SET item_id = gen_random_uuid()
WHERE (user_id, item_id) = (SELECT user_id, item_id FROM item_status ORDER BY user_id, item_id LIMIT 1)`)

		return err
	})
	if err != nil {
		t.Fatalf("seed mismatch: %v", err)
	}

	var report bytes.Buffer

	err = legacyimport.Verify(ctx, db, legacyimport.VerifyOptions{SamplePct: 100, Disk: disk, Out: &report})
	if err == nil {
		t.Fatalf("verify passed on a seeded mismatch:\n%s", report.String())
	}

	t.Log(err)

	for _, want := range []string{"sample items", "covers items", "orphan item_status"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not name %q\n%s", err, want, report.String())
		}
	}
}
