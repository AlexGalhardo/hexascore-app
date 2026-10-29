//go:build integration

package legacyimport_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/database/legacyimport"
)

// fixtureDB is a migrated app database with the synthetic legacy dump.
func fixtureDB(t *testing.T) (*vukemtest.App, *database.DB) {
	t.Helper()

	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	sql, err := os.ReadFile("testdata/fixture.sql")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(), string(sql)); err != nil {
		t.Fatalf("load fixture: %v", err)
	}

	return app, db
}

// copyUsers is a stand-in users stage: it copies legacy user ids in key
// order, and fails once at failAt to play a process killed mid-stage.
func copyUsers(calls *int, failAt int) legacyimport.BatchFunc {
	return func(ctx context.Context, tx *database.Tx, after string, n int) (string, int, error) {
		*calls++

		rows, err := tx.QueryContext(
			ctx,
			`SELECT id FROM legacy_import."user" WHERE id > $1 ORDER BY id LIMIT $2`,
			after,
			n,
		)
		if err != nil {
			return "", 0, err
		}
		defer func() { _ = rows.Close() }()

		var ids []string

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return "", 0, err
			}

			ids = append(ids, id)
		}

		if err := rows.Err(); err != nil {
			return "", 0, err
		}

		for _, id := range ids {
			// A plain INSERT: a batch imported twice would hit the primary key.
			if _, err := tx.ExecContext(ctx, `INSERT INTO import_target (id) VALUES ($1)`, id); err != nil {
				return "", 0, err
			}
		}

		if *calls == failAt {
			return "", 0, errors.New("killed")
		}

		if len(ids) == 0 {
			return "", 0, nil
		}

		return ids[len(ids)-1], len(ids), nil
	}
}

func TestRunnerResumes(t *testing.T) {
	t.Parallel()

	app, db := fixtureDB(t)
	ctx := t.Context()

	if _, err := db.ExecContext(ctx, `CREATE TABLE import_target (id text PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}

	calls := 0
	stage := []legacyimport.Stage{{Name: "users", Batch: copyUsers(&calls, 3)}}

	var out bytes.Buffer

	r := &legacyimport.Runner{DB: db, Batch: 2, Out: &out, Now: app.Now}

	if err := r.Run(ctx, stage); err == nil || !strings.Contains(err.Error(), "killed") {
		t.Fatalf("first run: err = %v, want the kill", err)
	}

	var (
		status, last string
		rows, copied int
	)

	read := func() {
		t.Helper()

		if err := db.QueryRowContext(ctx, `SELECT status, coalesce(last_key, ''), rows FROM legacy_import_runs
WHERE stage = 'users'`).Scan(&status, &last, &rows); err != nil {
			t.Fatal(err)
		}

		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM import_target`).Scan(&copied); err != nil {
			t.Fatal(err)
		}
	}

	read()

	// Two batches of two committed with their checkpoint; the third rolled back.
	if status != "failed" || last != "u04" || rows != 4 || copied != 4 {
		t.Fatalf("after kill: status=%s last=%s rows=%d copied=%d, want failed u04 4 4", status, last, rows, copied)
	}

	out.Reset()

	if err := r.Run(ctx, stage); err != nil {
		t.Fatalf("resume: %v\n%s", err, out.String())
	}

	read()

	if status != "done" || last != "u07" || rows != 7 || copied != 7 {
		t.Fatalf("after resume: status=%s last=%s rows=%d copied=%d, want done u07 7 7", status, last, rows, copied)
	}

	if !strings.Contains(out.String(), "resuming after key u04") {
		t.Fatalf("resume output:\n%s", out.String())
	}

	// A finished stage is skipped on the next run.
	before := calls
	if err := r.Run(ctx, stage); err != nil || calls != before {
		t.Fatalf("rerun of a done stage: err=%v calls %d -> %d", err, before, calls)
	}

	// Unimplemented stages refuse to run instead of marking themselves done.
	if err := r.Run(ctx, []legacyimport.Stage{{Name: "catalog"}}); !errors.Is(err, legacyimport.ErrNotImplemented) {
		t.Fatalf("stage without Batch: err = %v, want ErrNotImplemented", err)
	}
}

func TestPreflightAborts(t *testing.T) {
	t.Parallel()

	_, db := fixtureDB(t)
	ctx := t.Context()

	res, err := legacyimport.Preflight(ctx, db)
	if err != nil {
		t.Fatalf("clean fixture: %v", err)
	}

	if res.PasskeyOnly != 1 {
		t.Fatalf("passkey-only = %d, want 1", res.PasskeyOnly)
	}

	for _, q := range []string{
		// Same address in another case: one duplicate group.
		`INSERT INTO legacy_import."user" (id, name, email, email_verified) VALUES ('u08', 'Ana 2', 'ANA@Example.test', true)`,
		`INSERT INTO legacy_import.account (id, account_id, provider_id, user_id, password) VALUES ('a08', 'u08', 'credential', 'u08', 'x')`,
		// 2FA flagged on with no secret row.
		`INSERT INTO legacy_import."user" (id, name, email, two_factor_enabled) VALUES ('u09', 'Hugo', 'hugo@example.test', true)`,
		`INSERT INTO legacy_import.account (id, account_id, provider_id, user_id, password) VALUES ('a09', 'u09', 'credential', 'u09', 'x')`,
		// Nothing to sign in with: no password, no Google, no passkey.
		`INSERT INTO legacy_import."user" (id, name, email) VALUES ('u10', 'Iris', 'iris@example.test')`,
		`INSERT INTO legacy_import.account (id, account_id, provider_id, user_id, password) VALUES ('a10', 'u10', 'credential', 'u10', NULL)`,
	} {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	res, err = legacyimport.Preflight(ctx, db)

	var pf *legacyimport.PreflightError
	if !errors.As(err, &pf) {
		t.Fatalf("err = %v, want *PreflightError", err)
	}

	want := map[string]int64{"duplicate_emails": 1, "twofactor_without_row": 1, "no_login_method": 1}
	for k, v := range want {
		if res.Failures[k] != v {
			t.Errorf("%s = %d, want %d (all: %v)", k, res.Failures[k], v, res.Failures)
		}
	}

	for _, s := range []string{"duplicate_emails=1", "twofactor_without_row=1", "no_login_method=1"} {
		if !strings.Contains(err.Error(), s) {
			t.Errorf("error %q lacks %s", err, s)
		}
	}
}
