//go:build integration

package community

import (
	"context"
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

	// The ROADMAP command updates id 1. A fresh table has no such row, and a
	// statement trigger still refuses the write.
	_, err = db.ExecContext(ctx, `UPDATE catalog_audit SET after = 'null' WHERE id = 1`)
	assertAppendOnly(t, err)

	if _, err := db.ExecContext(ctx, `INSERT INTO catalog_audit
(actor_role, source, target_type, target_id, field, after)
VALUES ('system', 'admin', 'item', '00000000-0000-0000-0000-000000000001', 'title', '"x"'::jsonb)`); err != nil {
		t.Fatal(err)
	}

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
