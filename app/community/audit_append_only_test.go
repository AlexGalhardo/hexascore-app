//go:build integration

package community

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/schema"

	"github.com/AlexGalhardo/hexascore-app/database/migrations"
)

func TestAuditAppendOnly(t *testing.T) {
	ctx := context.Background()

	db, err := database.Open(ctx, pgtest.Database(t, ""), database.WithDriver(postgres.Driver()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	m, err := schema.NewMigrator(db, migrations.All())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(ctx); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO catalog_audit
(actor_role, source, target_type, target_id, field, after)
VALUES ('system', 'admin', 'item', '00000000-0000-0000-0000-000000000001', 'title', '"x"'::jsonb)`); err != nil {
		t.Fatal(err)
	}

	_, err = db.ExecContext(ctx, `UPDATE catalog_audit SET after = 'null' WHERE id = 1`)
	assertAppendOnly(t, err)

	_, err = db.ExecContext(ctx, `UPDATE catalog_audit SET after = 'null'`)
	assertAppendOnly(t, err)

	_, err = db.ExecContext(ctx, `DELETE FROM catalog_audit`)
	assertAppendOnly(t, err)

	var n int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM catalog_audit`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("insert survived, count = %d", n)
	}
}

// The guard is a row trigger: a statement trigger fired by the ON DELETE SET
// NULL foreign key on actor_id with zero rows and blocked every user delete.
func TestAuditGuardDoesNotBlockUserDelete(t *testing.T) {
	ctx := context.Background()

	db, err := database.Open(ctx, pgtest.Database(t, ""), database.WithDriver(postgres.Driver()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	m, err := schema.NewMigrator(db, migrations.All())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(ctx); err != nil {
		t.Fatal(err)
	}

	const addUser = `INSERT INTO users (name, email, password, created_at, updated_at)
VALUES ('U', $1, 'x', now(), now()) RETURNING id`

	var plain, author int64
	if err := db.QueryRowContext(ctx, addUser, "plain@example.test").Scan(&plain); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, addUser, "author@example.test").Scan(&author); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, plain); err != nil {
		t.Fatalf("delete user without audit rows: %v", err)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO catalog_audit
(actor_id, actor_role, source, target_type, target_id, field, after)
VALUES ($1, 'system', 'admin', 'item', '00000000-0000-0000-0000-000000000002', 'title', '"x"'::jsonb)`, author); err != nil {
		t.Fatal(err)
	}

	// Erasing an author nulls actor_id on their audit rows and nothing else.
	if _, err := db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, author); err != nil {
		t.Fatalf("delete user with audit rows: %v", err)
	}

	var actor sql.NullInt64
	if err := db.QueryRowContext(ctx, `SELECT actor_id FROM catalog_audit`).Scan(&actor); err != nil {
		t.Fatal(err)
	}
	if actor.Valid {
		t.Fatalf("actor_id = %d, want NULL", actor.Int64)
	}

	_, err = db.ExecContext(ctx, `UPDATE catalog_audit SET actor_id = 1`)
	assertAppendOnly(t, err)
}

func assertAppendOnly(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected append-only error")
	}
	if got := sqlState(err); got != "P0001" {
		t.Fatalf("SQLSTATE %q, err %v", got, err)
	}
	if !strings.Contains(err.Error(), "catalog_audit is append-only") {
		t.Fatalf("message %v", err)
	}
}

func sqlState(err error) string {
	type stater interface{ SQLState() string }
	for err != nil {
		if s, ok := err.(stater); ok {
			return s.SQLState()
		}
		err = errors.Unwrap(err)
	}
	return ""
}
