//go:build integration

package community

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// applyTitle approves a title edit on the seeded item and returns its audit row id.
func applyTitle(t *testing.T, a *Applier, db *database.DB, id string, author int64, before, after string) int64 {
	t.Helper()
	seedApplySuggestion(t, db, id, author, "in_review",
		`{"title":{"before":"`+before+`","after":"`+after+`"}}`)
	if err := a.Approve(context.Background(), id, 1, ApproveOptions{}, adminActor); err != nil {
		t.Fatal(err)
	}

	return auditOf(t, db, id)
}

func auditOf(t *testing.T, db *database.DB, suggestion string) int64 {
	t.Helper()
	var id int64
	if err := db.QueryRowContext(context.Background(),
		`SELECT id FROM catalog_audit WHERE suggestion_id = $1 AND source <> 'revert' ORDER BY id LIMIT 1`, suggestion).
		Scan(&id); err != nil {
		t.Fatal(err)
	}

	return id
}

func revertRows(t *testing.T, db *database.DB) []int64 {
	t.Helper()
	var raw []int64
	rows, err := db.QueryContext(context.Background(),
		`SELECT reverts_audit_id FROM catalog_audit WHERE source = 'revert' ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		raw = append(raw, id)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	return raw
}

func revertedTotal(t *testing.T, db *database.DB, user int64) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(
		context.Background(),
		`SELECT COALESCE((SELECT reverted_total FROM contributor_stats WHERE user_id = $1), 0)`,
		user,
	).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

func TestRevertSingle(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	author := insertAuthor(t, db, "revert-author@example.com", time.Now().AddDate(0, 0, -30), true, false)
	a, fake := newApplier(t, db)
	const s1 = "00000000-0000-7000-8000-00000000b001"
	audit := applyTitle(t, a, db, s1, author, "Old", "New")
	jobs := len(fake.Messages())

	if err := a.Revert(context.Background(), audit, RevertSingle, "incorrect", adminActor); err != nil {
		t.Fatal(err)
	}

	if title, prov := originalTitle(t, db); title != "Old" || prov != "wikidata" {
		t.Fatalf("after revert: %s/%s, want Old/wikidata (exact provenance restored)", title, prov)
	}
	assertStatus(t, db, s1, "reverted", 4)
	var before, after, prev string
	if err := db.QueryRowContext(context.Background(), `SELECT before::text, after::text, prev_provenance::text
FROM catalog_audit WHERE reverts_audit_id = $1`, audit).Scan(&before, &after, &prev); err != nil {
		t.Fatal(err)
	}
	if before != `"New"` || after != `"Old"` || prev != `"community"` {
		t.Fatalf("revert row before=%s after=%s prev=%s", before, after, prev)
	}
	if n := revertedTotal(t, db, author); n != 1 {
		t.Fatalf("reverted_total = %d, want 1 for reason incorrect", n)
	}
	var action string
	if err := db.QueryRowContext(
		context.Background(),
		`SELECT action FROM api_changes WHERE item_id = $1 ORDER BY id DESC LIMIT 1`,
		applyItem,
	).Scan(&action); err != nil ||
		action != "updated" {
		t.Fatalf("api_changes %q %v", action, err)
	}
	if n := len(fake.Messages()) - jobs; n != 2 {
		t.Fatalf("revert dispatched %d jobs, want catalog.changed and community.notify", n)
	}

	// A row reverts once (unique index) and a revert row is never reverted.
	var ce *TransitionError
	if err := a.Revert(context.Background(), audit, RevertSingle, "incorrect", adminActor); !errors.As(err, &ce) ||
		ce.Code != "already-reverted" {
		t.Fatalf("second revert = %v", err)
	}
	var r int64
	if err := db.QueryRowContext(context.Background(), `SELECT id FROM catalog_audit WHERE source = 'revert'`).
		Scan(&r); err != nil {
		t.Fatal(err)
	}
	if err := a.Revert(context.Background(), r, RevertForce, "other", adminActor); !errors.As(err, &ce) ||
		ce.Code != "not-revertible" {
		t.Fatalf("revert of a revert = %v", err)
	}
}

func TestRevertReasonOtherKeepsAccuracy(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	author := insertAuthor(t, db, "revert-other@example.com", time.Now().AddDate(0, 0, -30), true, false)
	a, _ := newApplier(t, db)
	audit := applyTitle(t, a, db, "00000000-0000-7000-8000-00000000b011", author, "Old", "New")

	if err := a.Revert(context.Background(), audit, RevertSingle, "policy", adminActor); err != nil {
		t.Fatal(err)
	}
	if n := revertedTotal(t, db, author); n != 0 {
		t.Fatalf("reverted_total = %d, only incorrect counts", n)
	}
	if err := a.Revert(context.Background(), audit, "sideways", "policy", adminActor); err == nil {
		t.Fatal("unknown mode accepted")
	}
}

func TestRevertConflictCascadeForce(t *testing.T) {
	for _, mode := range []string{RevertSingle, RevertCascade, RevertForce} {
		t.Run(mode, func(t *testing.T) {
			db := openCommunityDB(t)
			seedApplyItem(t, db)
			a, _ := newApplier(t, db)
			const s1, s2 = "00000000-0000-7000-8000-00000000b021", "00000000-0000-7000-8000-00000000b022"
			first := applyTitle(t, a, db, s1, 0, "Old", "A")
			later := applyTitle(t, a, db, s2, 0, "A", "B")

			err := a.Revert(context.Background(), first, mode, "incorrect", adminActor)
			switch mode {
			case RevertSingle:
				var rc *RevertConflictError
				if !errors.As(err, &rc) || !slices.Equal(rc.LaterAuditIDs, []int64{later}) {
					t.Fatalf("single over a later change = %v, want revert-conflict [%d]", err, later)
				}
				if title, _ := originalTitle(t, db); title != "B" {
					t.Fatalf("conflict wrote the catalogue: %s", title)
				}
			case RevertCascade:
				if err != nil {
					t.Fatal(err)
				}
				if got := revertRows(t, db); !slices.Equal(got, []int64{later, first}) {
					t.Fatalf("cascade order = %v, want newest first", got)
				}
				assertStatus(t, db, s1, "reverted", 4)
				assertStatus(t, db, s2, "reverted", 4)
			case RevertForce:
				if err != nil {
					t.Fatal(err)
				}
				assertStatus(t, db, s2, "applied", 3)
				var before string
				if err := db.QueryRowContext(
					context.Background(),
					`SELECT before::text FROM catalog_audit WHERE reverts_audit_id = $1`,
					first,
				).Scan(&before); err != nil ||
					before != `"B"` {
					t.Fatalf("force revert before = %s, want the current value", before)
				}
			}
			if mode != RevertSingle {
				if title, prov := originalTitle(t, db); title != "Old" || prov != "wikidata" {
					t.Fatalf("%s: %s/%s", mode, title, prov)
				}
			}
		})
	}
}

func TestRevertRelationRestoresSource(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	a, _ := newApplier(t, db)
	const s = "00000000-0000-7000-8000-00000000b031"
	seedApplySuggestion(t, db, s, 0, "in_review", `{"genres":{"add":[{"id":"00000000-0000-0000-0000-0000000000e2"}],
"remove":[{"id":"00000000-0000-0000-0000-0000000000e1"}]}}`)
	if err := a.Approve(context.Background(), s, 1, ApproveOptions{}, adminActor); err != nil {
		t.Fatal(err)
	}

	if err := a.Revert(context.Background(), auditOf(t, db, s), RevertSingle, "incorrect", adminActor); err != nil {
		t.Fatal(err)
	}

	var got string
	if err := db.QueryRowContext(context.Background(), `SELECT string_agg(entity_id::text || ':' || source, ',')
FROM item_entities WHERE item_id = $1`, applyItem).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != "00000000-0000-0000-0000-0000000000e1:wikidata" {
		t.Fatalf("relations after revert = %s", got)
	}
}

func TestRevertRefusesMergeAndNonAdmins(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	a, _ := newApplier(t, db)
	var merge int64
	if err := db.QueryRowContext(context.Background(), `INSERT INTO catalog_audit
(actor_role, source, target_type, target_id, field, before, after) VALUES ('admin','merge','item',$1,'*merge*','{}','{}')
RETURNING id`, applyItem).Scan(&merge); err != nil {
		t.Fatal(err)
	}

	var ce *TransitionError
	if err := a.Revert(context.Background(), merge, RevertForce, "incorrect", adminActor); !errors.As(err, &ce) ||
		ce.Code != "not-revertible" {
		t.Fatalf("merge revert = %v", err)
	}

	audit := applyTitle(t, a, db, "00000000-0000-7000-8000-00000000b041", 0, "Old", "New")
	if err := a.Revert(context.Background(), audit, RevertSingle, "incorrect", Actor{Role: "moderator"}); !errors.Is(
		err, ErrAdminOnly) {
		t.Fatalf("moderator revert = %v", err)
	}
	if err := a.Revert(
		context.Background(),
		999999,
		RevertSingle,
		"incorrect",
		adminActor,
	); !errors.Is(
		err,
		ErrAuditNotFound,
	) {
		t.Fatalf("missing row = %v", err)
	}
}
