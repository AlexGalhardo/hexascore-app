//go:build integration

package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/schema"

	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/database/migrations"
)

const changedItem = "00000000-0000-0000-0000-0000000000aa"

func openJobsDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.Open(t.Context(), pgtest.Database(t, ""), database.WithDriver(postgres.Driver()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	m, err := schema.NewMigrator(db, migrations.All())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(t.Context()); err != nil {
		t.Fatal(err)
	}

	return db
}

func exec(t *testing.T, db *database.DB, q string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), q, args...); err != nil {
		t.Fatal(err)
	}
}

func dirtyCount(t *testing.T, db *database.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(context.Background(), `SELECT count(*) FROM search_dirty WHERE ref_id = $1`,
		changedItem).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

func TestCatalogChangedIdempotent(t *testing.T) {
	db := openJobsDB(t)
	exec(t, db, `INSERT INTO items (id, type, slug, wikidata_qid, updated_at)
VALUES ($1, 'movie', 'm', 'Q1', '2020-01-01T00:00:00Z')`, changedItem)
	for range 3 {
		exec(t, db, `INSERT INTO catalog_audit (actor_role, source, target_type, target_id, field, before, after)
VALUES ('admin', 'community', 'item', $1, 'title', '"a"', '"b"')`, changedItem)
	}
	exec(t, db, `DELETE FROM search_dirty`)

	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	job := community.CatalogChanged{
		TargetType: "item",
		TargetID:   changedItem,
		Fields:     []string{"title"},
		AuditIDs:   []int64{1, 2},
	}

	ran, err := RunCatalogChanged(context.Background(), db, job, now)
	if err != nil || !ran {
		t.Fatalf("first run ran=%v err=%v", ran, err)
	}
	if n := dirtyCount(t, db); n != 1 {
		t.Fatalf("reindex rows after first run = %d", n)
	}
	var updated time.Time
	if err := db.QueryRowContext(context.Background(), `SELECT updated_at FROM items WHERE id = $1`, changedItem).
		Scan(&updated); err != nil || !updated.Equal(now) {
		t.Fatalf("lastmod source = %v %v", updated, err)
	}

	// The search sync drained the row; a replay of the same audit ids does nothing.
	exec(t, db, `DELETE FROM search_dirty`)
	ran, err = RunCatalogChanged(context.Background(), db, job, now.Add(time.Hour))
	if err != nil || ran {
		t.Fatalf("replay ran=%v err=%v", ran, err)
	}
	if n := dirtyCount(t, db); n != 0 {
		t.Fatalf("replay reindexed again: %d", n)
	}

	// A job with one new audit id runs again.
	job.AuditIDs = []int64{2, 3}
	if ran, err := RunCatalogChanged(context.Background(), db, job, now.Add(time.Hour)); err != nil || !ran {
		t.Fatalf("new audit id ran=%v err=%v", ran, err)
	}
	if n := dirtyCount(t, db); n != 1 {
		t.Fatalf("reindex rows after new audit id = %d", n)
	}
}

func TestCatalogChangedEditionReindexesWork(t *testing.T) {
	db := openJobsDB(t)
	exec(t, db, `INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata','W','CC0-1.0','https://w')`)
	exec(t, db, `INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'book', 'b', 'Q1')`, changedItem)
	exec(t, db, `INSERT INTO editions (id, work_id, isbn13, title, language, source)
VALUES ('00000000-0000-0000-0000-0000000000ed', $1, '9780306406157', 'E', 'en', 'wikidata')`, changedItem)
	exec(t, db, `INSERT INTO catalog_audit (actor_role, source, target_type, target_id, field, before, after)
VALUES ('admin', 'community', 'edition', '00000000-0000-0000-0000-0000000000ed', 'title', '"E"', '"F"')`)
	exec(t, db, `DELETE FROM search_dirty`)

	job := community.CatalogChanged{
		TargetType: "edition", TargetID: "00000000-0000-0000-0000-0000000000ed", Fields: []string{"title"},
		AuditIDs: []int64{1},
	}
	if ran, err := RunCatalogChanged(context.Background(), db, job, time.Now()); err != nil || !ran {
		t.Fatalf("ran=%v err=%v", ran, err)
	}
	if n := dirtyCount(t, db); n != 1 {
		t.Fatalf("work not queued for reindex: %d", n)
	}
}
