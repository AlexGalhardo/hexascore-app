//go:build integration

package community

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/uuid"
)

func TestSaveDraft(t *testing.T) {
	db := openCommunityDB(t)
	ctx := context.Background()
	now := time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC)
	author := insertAuthor(t, db, "draft-author@example.com", now.Add(-30*24*time.Hour), true, false)
	other := insertAuthor(t, db, "draft-other@example.com", now.Add(-30*24*time.Hour), true, false)
	target := Target{Kind: "item", ID: uuid.NewV7().String()}
	src := "https://www.wikidata.org/wiki/Q127367"

	id, v, err := SaveDraft(ctx, db, author, "", 0, target, edit(change("title", nil, "First"), src), now)
	if err != nil || v != 1 {
		t.Fatalf("create: %v, version %d", err, v)
	}

	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM suggestions WHERE id = $1`, id).
		Scan(&status); err != nil ||
		status != "draft" {
		t.Fatalf("status %q, %v", status, err)
	}

	if n := eventCount(t, db, id); n != 1 {
		t.Fatalf("%d events, want 1", n)
	}

	if _, v, err = SaveDraft(
		ctx,
		db,
		author,
		id,
		1,
		target,
		edit(change("title", nil, "Second"), src),
		now,
	); err != nil ||
		v != 2 {
		t.Fatalf("update: %v, version %d", err, v)
	}

	var title string
	if err := db.QueryRowContext(ctx, `SELECT changes->'title'->>'after' FROM suggestions WHERE id = $1`, id).
		Scan(&title); err != nil ||
		title != "Second" {
		t.Fatalf("title %q, %v", title, err)
	}

	in := edit(change("title", nil, "Third"), src)

	if _, _, err = SaveDraft(ctx, db, author, id, 1, target, in, now); !isCode(err, "stale-before") {
		t.Fatalf("stale version: %v", err)
	}

	if _, _, err = SaveDraft(ctx, db, other, id, 2, target, in, now); !errors.Is(err, ErrDraftNotFound) {
		t.Fatalf("foreign draft: %v", err)
	}

	if _, _, err = SaveDraft(ctx, db, author, id, 2, target, edit(change("title", nil, ""), src), now); err == nil {
		t.Fatal("invalid draft saved")
	}

	if err := Transition(ctx, db, id, 2, "withdrawn", Actor{Role: "author"}); err != nil {
		t.Fatal(err)
	}

	if _, _, err = SaveDraft(ctx, db, author, id, 3, target, in, now); !isCode(err, "invalid-transition") {
		t.Fatalf("withdrawn draft: %v", err)
	}
}

func isCode(err error, code string) bool {
	e, ok := errors.AsType[*TransitionError](err)

	return ok && e.Code == code
}
