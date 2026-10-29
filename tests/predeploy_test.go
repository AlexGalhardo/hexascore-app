//go:build integration

package tests

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/schema"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/console"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// Two pre-deploys on an empty database: without serialisation the second
// would race the first and fail on existing tables; with vukem's runner
// lock one migrates everything and the other finds nothing to do.
func TestPredeploy_ConcurrentRunsSerialise(t *testing.T) {
	dsn := pgtest.Database(t, "")

	var outs [2]bytes.Buffer

	apps := make([]*vukemtest.App, 2)
	for i := range apps {
		apps[i] = vukemtest.New(t, bootstrap.New, vukemtest.Options{
			Env:            map[string]string{"DATABASE_URL": dsn},
			SkipMigrations: true,
			AppOptions:     []vukem.Option{vukem.WithStdout(&outs[i])},
		})
	}

	var wg sync.WaitGroup

	errs := make([]error, 2)
	for i, app := range apps {
		wg.Go(func() { errs[i] = app.Call(context.Background(), "hexascore:predeploy", nil) })
	}

	wg.Wait()

	migrated := 0

	for i := range apps {
		if errs[i] != nil {
			t.Fatalf("run %d: %v\n%s", i, errs[i], outs[i].String())
		}

		switch out := outs[i].String(); {
		case strings.Contains(out, "Migrated:"):
			migrated++
		case !strings.Contains(out, "Nothing to migrate."):
			t.Fatalf("run %d printed neither Migrated nor Nothing to migrate:\n%s", i, out)
		}
	}

	if migrated != 1 {
		t.Fatalf("runs that migrated = %d, want 1:\n%s\n---\n%s", migrated, outs[0].String(), outs[1].String())
	}
}

// A failed concurrent unique index leaves an INVALID index; `IF NOT EXISTS`
// alone would skip it and record the migration. The pre-deploy drops the
// leftover first, so the retried migration rebuilds a valid index.
func TestIndexConcurrently_DropsInvalid(t *testing.T) {
	ctx := t.Context()

	db, err := database.Open(ctx, pgtest.Database(t, ""), database.WithDriver(postgres.Driver()))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = db.Close() })

	exec := func(q string) {
		t.Helper()

		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	exec(`CREATE TABLE predeploy_t (k int)`)
	exec(`INSERT INTO predeploy_t VALUES (1), (1)`)

	m, err := schema.NewMigrator(db, []schema.Migration{{
		ID:            "20990101000000_index_predeploy_t_k",
		NoTransaction: true,
		Up: func(s *schema.Schema) error {
			return s.Raw("CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS predeploy_t_k_uq ON predeploy_t (k)")
		},
	}})
	if err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer

	if err := console.RunPredeploy(ctx, db, m, &out); err == nil {
		t.Fatal("unique index over duplicates did not fail")
	}

	valid := func() string {
		t.Helper()

		var v string
		if err := db.SQL().QueryRowContext(ctx, `SELECT coalesce((SELECT i.indisvalid::text FROM pg_index i
JOIN pg_class c ON c.oid = i.indexrelid WHERE c.relname = 'predeploy_t_k_uq'), 'missing')`).Scan(&v); err != nil {
			t.Fatal(err)
		}

		return v
	}

	if got := valid(); got != "false" {
		t.Fatalf("after failed build indisvalid = %s, want false", got)
	}

	exec(`DELETE FROM predeploy_t WHERE ctid = (SELECT max(ctid) FROM predeploy_t)`)

	out.Reset()

	if err := console.RunPredeploy(ctx, db, m, &out); err != nil {
		t.Fatalf("retry: %v\n%s", err, out.String())
	}

	if !strings.Contains(out.String(), "Dropped INVALID index: predeploy_t_k_uq") ||
		!strings.Contains(out.String(), "Migrated: 20990101000000_index_predeploy_t_k") {
		t.Fatalf("retry output:\n%s", out.String())
	}

	if got := valid(); got != "true" {
		t.Fatalf("after retry indisvalid = %s, want true", got)
	}
}
