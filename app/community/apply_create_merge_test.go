//go:build integration

package community

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"
)

const mergeInto = "00000000-0000-0000-0000-0000000000bb"

func mustDaysAgo(n int) time.Time { return time.Now().AddDate(0, 0, -n) }

func TestApplyCreate(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	author := insertAuthor(t, db, "create-author@example.com", mustDaysAgo(30), true, false)
	const id, dep = "00000000-0000-7000-8000-00000000b001", "00000000-0000-7000-8000-00000000b002"
	if _, err := db.ExecContext(context.Background(), `INSERT INTO suggestions
(id, author_id, kind, target_type, target_domain, locale, changes, status, version, submitted_at)
VALUES ($1, $2, 'create', 'item', 'movie', 'en', $3::jsonb, 'in_review', 1, now())`, id, author,
		`{"title":{"before":null,"after":"Brand New Film"},"external_ids.wikidata":{"before":null,"after":"Q42"},
"release_date":{"before":null,"after":{"value":"2020-05-01","precision":"day"}}}`); err != nil {
		t.Fatal(err)
	}
	// An edit that waited for the create: approved, blocked on depends_on.
	if _, err := db.ExecContext(context.Background(), `INSERT INTO suggestions
(id, kind, target_type, target_id, target_domain, locale, changes, status, version, submitted_at, depends_on)
VALUES ($1, 'edit', 'item', $2, 'movie', 'en', '{"title":{"before":"Old","after":"Newer"}}', 'approved', 2, now(), $3)`,
		dep, applyItem, id); err != nil {
		t.Fatal(err)
	}

	a, fake := newApplier(t, db)
	if err := a.Approve(context.Background(), id, 1, ApproveOptions{}, adminActor); err != nil {
		t.Fatal(err)
	}

	var newID, slug, qid, title, prov string
	if err := db.QueryRowContext(context.Background(), `SELECT i.id::text, i.slug, i.wikidata_qid, i.original_title,
  i.provenance->>'original_title'
FROM items i JOIN suggestions s ON s.meta->>'created_id' = i.id::text WHERE s.id = $1`, id).
		Scan(&newID, &slug, &qid, &title, &prov); err != nil {
		t.Fatal(err)
	}
	if slug != "brand-new-film" || qid != "Q42" || title != "Brand New Film" || prov != "community" {
		t.Fatalf("created item %s %s %s %s", slug, qid, title, prov)
	}

	var rows, nonNullBefore int
	if err := db.QueryRowContext(context.Background(), `SELECT count(*), count(*) FILTER (WHERE before IS NOT NULL
  AND before <> 'null'::jsonb) FROM catalog_audit WHERE suggestion_id = $1 AND source = 'create' AND target_id = $2`,
		id, newID).Scan(&rows, &nonNullBefore); err != nil || rows != 3 || nonNullBefore != 0 {
		t.Fatalf("create audit rows=%d nonNullBefore=%d err=%v", rows, nonNullBefore, err)
	}
	assertStatus(t, db, id, "applied", 3)

	var action string
	if err := db.QueryRowContext(context.Background(), `SELECT action FROM api_changes WHERE item_id = $1`, newID).
		Scan(&action); err != nil || action != "created" {
		t.Fatalf("api_changes %q %v", action, err)
	}

	// The dependent edit is unblocked and applied right after the create.
	assertStatus(t, db, dep, "applied", 3)
	if n := len(queue.Dispatched[CatalogChanged](fake)); n != 2 {
		t.Fatalf("catalog.changed jobs = %d, want create + dependent", n)
	}
}

func TestApplyCreateNeedsWikidata(t *testing.T) {
	db := openCommunityDB(t)
	const id = "00000000-0000-7000-8000-00000000b003"
	if _, err := db.ExecContext(context.Background(), `INSERT INTO suggestions
(id, kind, target_type, target_domain, locale, changes, status, version, submitted_at)
VALUES ($1, 'create', 'item', 'movie', 'en', '{"title":{"before":null,"after":"No Qid"}}', 'submitted', 1, now())`,
		id); err != nil {
		t.Fatal(err)
	}

	a, _ := newApplier(t, db)
	var ve *ValidationError
	if err := a.Approve(context.Background(), id, 1, ApproveOptions{}, adminActor); !errors.As(err, &ve) ||
		ve.Field != "external_ids.wikidata" {
		t.Fatalf("err = %#v", err)
	}
	assertStatus(t, db, id, "submitted", 1)
}

func TestApplyMerge(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	author := insertAuthor(t, db, "merge-author@example.com", mustDaysAgo(30), true, false)
	other := insertAuthor(t, db, "other-author@example.com", mustDaysAgo(30), true, false)
	if _, err := db.ExecContext(context.Background(), `INSERT INTO items (id, type, slug, wikidata_qid, original_title)
VALUES ($1, 'movie', 'keeper', 'Q2', 'Keeper')`, mergeInto); err != nil {
		t.Fatal(err)
	}
	const id, open = "00000000-0000-7000-8000-00000000b004", "00000000-0000-7000-8000-00000000b005"
	if _, err := db.ExecContext(context.Background(), `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, status, version, submitted_at)
VALUES ($1, $2, 'merge', 'item', $3, 'movie', 'en', $4::jsonb, 'submitted', 1, now())`, id, author, applyItem,
		`{"*merge*":{"before":"`+applyItem+`","after":"`+mergeInto+`"}}`); err != nil {
		t.Fatal(err)
	}
	seedApplySuggestion(t, db, open, other, "submitted", `{"title":{"before":"Old","after":"X"}}`)

	a, fake := newApplier(t, db)
	var ve *ValidationError
	if err := a.Approve(context.Background(), id, 1, ApproveOptions{ConfirmTitle: "wrong"}, adminActor); !errors.As(err,
		&ve) || ve.Field != "confirm_title" {
		t.Fatalf("missing typed title err = %#v", err)
	}

	if err := a.Approve(context.Background(), id, 1, ApproveOptions{ConfirmTitle: " old "}, adminActor); err != nil {
		t.Fatal(err)
	}

	var gone bool
	if err := db.QueryRowContext(context.Background(), `SELECT NOT EXISTS (SELECT 1 FROM items WHERE id = $1)`,
		applyItem).Scan(&gone); err != nil || !gone {
		t.Fatalf("source still there: %v %v", gone, err)
	}

	var src, snap string
	if err := db.QueryRowContext(context.Background(), `SELECT source::text, before->'source'->'item'->>'slug'
FROM catalog_audit WHERE field = '*merge*' AND suggestion_id = $1 AND target_id = $2`, id, mergeInto).
		Scan(&src, &snap); err != nil || src != "merge" || snap != "old" {
		t.Fatalf("merge audit src=%q snapshot slug=%q err=%v", src, snap, err)
	}
	assertStatus(t, db, id, "applied", 3)

	var status, reason string
	if err := db.QueryRowContext(context.Background(), `SELECT status::text, reject_reason::text FROM suggestions
WHERE id = $1`, open).Scan(&status, &reason); err != nil || status != "expired" || reason != "target_gone" {
		t.Fatalf("open suggestion on source %s/%s %v", status, reason, err)
	}

	// target_gone does not count against the other author.
	var rejected int
	_ = db.QueryRowContext(context.Background(), `SELECT COALESCE((SELECT rejected_total FROM contributor_stats
WHERE user_id = $1), 0)`, other).Scan(&rejected)
	if rejected != 0 {
		t.Fatalf("target_gone counted: %d", rejected)
	}

	var deleted string
	if err := db.QueryRowContext(context.Background(), `SELECT merged_into::text FROM api_changes
WHERE item_id = $1 AND action = 'deleted'`, applyItem).Scan(&deleted); err != nil || deleted != mergeInto {
		t.Fatalf("api_changes merged_into %q %v", deleted, err)
	}
	if n := len(queue.Dispatched[CatalogChanged](fake)); n != 1 {
		t.Fatalf("catalog.changed jobs = %d", n)
	}
}

func TestApplyCreateConcurrentQIDIsDuplicate(t *testing.T) {
	db := openCommunityDB(t)
	const id = "00000000-0000-7000-8000-00000000b009"
	if _, err := db.ExecContext(
		context.Background(),
		`INSERT INTO suggestions
(id, kind, target_type, target_domain, locale, changes, status, version, submitted_at)
VALUES ($1, 'create', 'item', 'movie', 'en', $2::jsonb, 'in_review', 1, now())`,
		id,
		`{"title":{"before":null,"after":"Racing Film"},"external_ids.wikidata":{"before":null,"after":"Q4242"}}`,
	); err != nil {
		t.Fatal(err)
	}

	a, _ := newApplier(t, db)
	done := make(chan error, 1)

	// An ingest inserts the same Q-id and commits only after the apply has passed its lookup.
	if err := db.Tx(context.Background(), func(tx *database.Tx) error {
		if _, err := tx.ExecContext(context.Background(), `INSERT INTO items (id, type, slug, wikidata_qid)
VALUES ('00000000-0000-7000-8000-00000000b0aa', 'movie', 'racing-film-ingest', 'Q4242')`); err != nil {
			return err
		}

		go func() { done <- a.Approve(context.Background(), id, 1, ApproveOptions{}, adminActor) }()
		time.Sleep(500 * time.Millisecond)

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	var p *Problem
	if err := <-done; !errors.As(err, &p) || p.Code != "suggestion-duplicate" ||
		p.ExistingID != "00000000-0000-7000-8000-00000000b0aa" {
		t.Fatalf("approve err = %v, want 409 suggestion-duplicate with the ingested id", err)
	}
}
