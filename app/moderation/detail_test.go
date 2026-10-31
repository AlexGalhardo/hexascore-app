//go:build integration

package moderation

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/AlexGalhardo/hexascore-app/app/community"
)

var adminActor = community.Actor{ID: sql.NullInt64{Int64: 1, Valid: true}, Role: "admin"}

// fakeApplier stands in for the P4.B.2.1 service: the first approval of a target wins, later ones are stale.
type fakeApplier struct {
	q       Queue
	applied map[string]bool
}

func (f *fakeApplier) Approve(ctx context.Context, id string, version int, _ bool, actor community.Actor) error {
	var target string
	if err := f.q.DB.QueryRowContext(ctx, `SELECT target_id::text FROM suggestions WHERE id=$1`, id).
		Scan(&target); err != nil {
		return err
	}
	if f.applied[target] {
		return &community.TransitionError{Status: 409, Code: "stale-before", Field: "title"}
	}
	f.applied[target] = true

	return community.Approve(ctx, f.q.DB, id, version, false, actor)
}

func TestDetailLoadsDiff(t *testing.T) {
	db := openDB(t)
	q := Queue{DB: db, Now: func() time.Time { return base }}
	id := insert(t, db, seed{changes: `{"title.en":{"before":"a","after":"<script>x</script>"}}`})

	d, err := q.Detail(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if d.Version != 1 || len(d.Diffs) != 1 || d.Diffs[0].After != "<script>x</script>" {
		t.Fatalf("detail: %+v", d)
	}
	if _, err := q.Detail(context.Background(), "00000000-0000-7000-8000-00000000ffff"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing id: %v", err)
	}
}

func TestReject(t *testing.T) {
	db := openDB(t)
	q := Queue{DB: db, Now: func() time.Time { return base }}
	ctx := context.Background()
	id := insert(t, db, seed{})

	var ie *InvalidError
	if err := q.Reject(ctx, id, 1, "", "", adminActor); !errors.As(err, &ie) || ie.Field != "reason" {
		t.Fatalf("reject without reason: %v, want InvalidError(reason)", err)
	}
	if err := q.Reject(ctx, id, 1, "nonsense", "", adminActor); !errors.As(err, &ie) {
		t.Fatalf("unknown reason: %v", err)
	}
	if err := q.Reject(ctx, id, 1, "no_source", "add a link", adminActor); err != nil {
		t.Fatal(err)
	}

	var status, reason, note string
	if err := db.QueryRowContext(ctx, `SELECT status, reject_reason, decision_note FROM suggestions WHERE id=$1`, id).
		Scan(&status, &reason, &note); err != nil {
		t.Fatal(err)
	}
	if status != "rejected" || reason != "no_source" || note != "add a link" {
		t.Fatalf("after reject: %s %s %s", status, reason, note)
	}
}

func TestApproveWithoutApplier(t *testing.T) {
	db := openDB(t)
	q := Queue{DB: db}
	id := insert(t, db, seed{})
	if err := q.Approve(context.Background(), id, 1, false, adminActor); !errors.Is(err, ErrApplyUnavailable) {
		t.Fatalf("approve without applier: %v", err)
	}
}

func TestBulk(t *testing.T) {
	db := openDB(t)
	q := Queue{DB: db}
	q.Applier = &fakeApplier{q: q, applied: map[string]bool{}}
	ctx := context.Background()

	tooMany := make([]string, 51)
	for i := range tooMany {
		tooMany[i] = "00000000-0000-7000-8000-00000000ffff"
	}
	if _, err := q.Bulk(ctx, BulkApprove, tooMany, "", "", false, adminActor); !errors.Is(err, ErrBulkSize) {
		t.Fatalf("51 ids: %v, want ErrBulkSize", err)
	}

	a := insert(t, db, seed{})
	b := insert(t, db, seed{})
	res, err := q.Bulk(ctx, BulkApprove, []string{a, b}, "", "", false, adminActor)
	if err != nil {
		t.Fatal(err)
	}
	if !res[0].OK || res[1].OK || res[1].Code != "stale-before" {
		t.Fatalf("competing bulk approve: %+v, want first ok then stale-before", res)
	}

	c := insert(t, db, seed{target: "00000000-0000-0000-0000-0000000000bb"})
	res, err = q.Bulk(
		ctx,
		BulkReject,
		[]string{c, "00000000-0000-7000-8000-00000000ffff", "not-a-uuid"},
		"spam",
		"",
		false,
		adminActor,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !res[0].OK || res[1].OK || res[1].Code != "not-found" || res[2].Code != "not-found" {
		t.Fatalf("bulk reject: %+v", res)
	}
	if _, err := q.Bulk(ctx, BulkReject, []string{a}, "", "", false, adminActor); err == nil {
		t.Fatal("bulk reject without reason must fail as a whole")
	}
}

func TestCloneAsDraft(t *testing.T) {
	db := openDB(t)
	q := Queue{DB: db}
	ctx := context.Background()

	live := insert(t, db, seed{author: 2})
	if _, err := q.CloneAsDraft(ctx, live, adminActor); err == nil {
		t.Fatal("cloning an open suggestion must fail")
	}

	done := insert(t, db, seed{author: 2, status: "rejected"})
	id, err := q.CloneAsDraft(ctx, done, adminActor)
	if err != nil {
		t.Fatal(err)
	}
	var status string
	var author int64
	var submitted sql.NullTime
	if err := db.QueryRowContext(ctx, `SELECT status, author_id, submitted_at FROM suggestions WHERE id=$1`, id).
		Scan(&status, &author, &submitted); err != nil {
		t.Fatal(err)
	}
	if status != "draft" || author != 2 || submitted.Valid {
		t.Fatalf("clone: %s %d %v", status, author, submitted)
	}

	orphan := insert(t, db, seed{status: "expired"})
	if _, err := q.CloneAsDraft(ctx, orphan, adminActor); !errors.Is(err, ErrNoAuthor) {
		t.Fatalf("clone without author: %v", err)
	}
}
