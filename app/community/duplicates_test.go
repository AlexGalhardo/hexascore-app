//go:build integration

package community

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

func TestDuplicate(t *testing.T) {
	db := openCommunityDB(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	target := "00000000-0000-0000-0000-0000000000b1"
	author := insertAuthor(t, db, "dup-a@example.test", now.Add(-30*24*time.Hour), true, false)
	other := insertAuthor(t, db, "dup-b@example.test", now.Add(-30*24*time.Hour), true, false)
	seedTitle(t, db, target, "dup-item", "Q424201", "Old")
	change := `{"translations.en.title":{"before":"Old","after":"New"}}`
	first := insertChangedDraft(t, db, author, 401, target, change)
	if status, err := Submit(ctx, db, first, author, 1, now); err != nil || status != http.StatusSeeOther {
		t.Fatalf("first: status %d err %v", status, err)
	}

	own := insertChangedDraft(t, db, author, 402, target, change)
	status, err := Submit(ctx, db, own, author, 1, now)
	var problem *Problem
	if status != http.StatusConflict || !errors.As(err, &problem) || problem.Code != "suggestion-duplicate" || problem.ExistingID != first {
		t.Fatalf("own duplicate: status %d err %#v", status, err)
	}
	assertStatus(t, db, own, "draft", 1)

	same := insertChangedDraft(t, db, other, 403, target, change)
	status, err = Submit(ctx, db, same, other, 1, now)
	problem = nil
	if status != http.StatusConflict || !errors.As(err, &problem) || problem.Code != "suggestion-duplicate" || problem.ExistingID != "" {
		t.Fatalf("same after: status %d err %#v", status, err)
	}
	// A refused submit stays draft, and an open draft on the same field is still a duplicate.
	if err := Transition(ctx, db, same, 1, "withdrawn", Actor{Role: "author"}); err != nil {
		t.Fatal(err)
	}

	alt := `{"translations.en.title":{"before":"Old","after":"Other"}}`
	rival := insertChangedDraft(t, db, other, 404, target, alt)
	status, err = Submit(ctx, db, rival, other, 1, now)
	if err != nil || status != http.StatusSeeOther {
		t.Fatalf("competing: status %d err %v", status, err)
	}
	if !competing(t, db, first) || !competing(t, db, rival) {
		t.Fatal("competing flag missing")
	}
}

func TestFieldLock(t *testing.T) {
	db := openCommunityDB(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	target := "00000000-0000-0000-0000-0000000000b2"
	author := insertAuthor(t, db, "lock@example.test", now.Add(-30*24*time.Hour), true, false)
	change := `{"translations.en.title":{"before":null,"after":"New"}}`
	if _, err := db.ExecContext(ctx, `INSERT INTO catalog_locks (id, target_type, target_id, field, reason, expires_at)
VALUES ('00000000-0000-7000-8000-0000000000c1', 'item', $1, 'translations.en.title', 'frozen', $2)`, target, now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	expired := insertChangedDraft(t, db, author, 411, target, change)
	if status, err := Submit(ctx, db, expired, author, 1, now); err != nil || status != http.StatusSeeOther {
		t.Fatalf("expired lock: status %d err %v", status, err)
	}

	if _, err := db.ExecContext(ctx, `UPDATE catalog_locks SET expires_at = NULL, reason = 'dispute' WHERE id = '00000000-0000-7000-8000-0000000000c1'`); err != nil {
		t.Fatal(err)
	}
	locked := insertChangedDraft(t, db, author, 412, target, `{"translations.en.title":{"before":null,"after":"Newer"}}`)
	status, err := Submit(ctx, db, locked, author, 1, now)
	var problem *Problem
	if status != http.StatusLocked || !errors.As(err, &problem) || problem.Code != "field-locked" || problem.LockReason != "dispute" || problem.Field != "translations.en.title" {
		t.Fatalf("lock: status %d err %#v", status, err)
	}
	assertStatus(t, db, locked, "draft", 1)
}

func TestStaleBeforeAtSubmit(t *testing.T) {
	db := openCommunityDB(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	target := "00000000-0000-0000-0000-0000000000b3"
	author := insertAuthor(t, db, "stale@example.test", now.Add(-30*24*time.Hour), true, false)
	seedTitle(t, db, target, "stale-item", "Q424242", "Live")
	id := insertChangedDraft(t, db, author, 421, target, `{"translations.en.title":{"before":"Old","after":"New"}}`)
	status, err := Submit(ctx, db, id, author, 1, now)
	var problem *Problem
	if status != http.StatusConflict || !errors.As(err, &problem) || problem.Code != "stale-before" || problem.Field != "translations.en.title" || problem.Current != `"Live"` {
		t.Fatalf("stale: status %d err %#v", status, err)
	}
	assertStatus(t, db, id, "draft", 1)
	if err := Transition(ctx, db, id, 1, "withdrawn", Actor{Role: "author"}); err != nil {
		t.Fatal(err)
	}

	fresh := insertChangedDraft(t, db, author, 422, target, `{"translations.en.title":{"before":"Live","after":"New"}}`)
	status, err = Submit(ctx, db, fresh, author, 1, now)
	if err != nil || status != http.StatusSeeOther {
		t.Fatalf("fresh: status %d err %v", status, err)
	}
}

func seedTitle(t *testing.T, db *database.DB, id, slug, qid, title string) {
	t.Helper()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `INSERT INTO data_sources (code, name, licence, url) VALUES ('wikidata', 'Wikidata', 'CC0', 'https://www.wikidata.org') ON CONFLICT (code) DO NOTHING`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'movie', $2, $3)`, id, slug, qid); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', $2, 'wikidata')`, id, title); err != nil {
		t.Fatal(err)
	}
}

func insertChangedDraft(t *testing.T, db *database.DB, authorID int64, n int, target, changes string) string {
	t.Helper()
	id := fmt.Sprintf("00000000-0000-7000-8000-%012d", n)
	_, err := db.ExecContext(context.Background(), `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, status, version)
VALUES ($1, $2, 'edit', 'item', $3, 'movie', 'en', $4::jsonb, 'draft', 1)`, id, authorID, target, changes)
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func competing(t *testing.T, db *database.DB, id string) bool {
	t.Helper()
	var flag bool
	if err := db.QueryRowContext(context.Background(), `SELECT coalesce(meta->>'competing','') = 'true' FROM suggestions WHERE id = $1`, id).Scan(&flag); err != nil {
		t.Fatal(err)
	}

	return flag
}
