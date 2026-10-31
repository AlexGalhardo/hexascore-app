//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/schema"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/database/migrations"
)

func TestCommunitySchema(t *testing.T) {
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

	for _, table := range []string{"suggestions", "suggestion_events", "catalog_audit", "contributor_stats", "catalog_locks"} {
		var got string
		if err := db.QueryRowContext(ctx, `SELECT to_regclass($1)`, "public."+table).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != table {
			t.Fatalf("to_regclass(%s) = %q", table, got)
		}
	}

	var urls models.TextArray
	if err := db.QueryRowContext(ctx, `SELECT $1::text[]`, models.TextArray{"https://a.example/a,b", "https://b.example"}).
		Scan(&urls); err != nil {
		t.Fatal(err)
	}
	if len(urls) != 2 || urls[0] != "https://a.example/a,b" || urls[1] != "https://b.example" {
		t.Fatalf("source urls = %#v", urls)
	}

	rolled, err := m.Rollback(ctx, 1)
	if err != nil || len(rolled) != 1 || rolled[0] != "20261030130000_create_community_tables" {
		t.Fatalf("rollback: %v %v", rolled, err)
	}
	var gone *string
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.suggestions')`).Scan(&gone); err != nil {
		t.Fatal(err)
	}
	if gone != nil {
		t.Fatalf("suggestions still present: %v", *gone)
	}
}
