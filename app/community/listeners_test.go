//go:build integration

package community

import (
	"context"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"
)

func insertOwned(t *testing.T, db *database.DB, id string, author int64, status string) {
	t.Helper()
	var a any
	if author != 0 {
		a = author
	}
	if _, err := db.ExecContext(context.Background(), `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, status, version, submitted_at)
VALUES ($1, $2, 'edit', 'item', $3, 'movie', 'en', '{"title":{"before":"Old","after":"New"}}', $4, 1, now())`,
		id, a, applyItem, status); err != nil {
		t.Fatal(err)
	}
}

func systemEvents(t *testing.T, db *database.DB, id, reason string) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(context.Background(), `SELECT count(*) FROM suggestion_events
WHERE suggestion_id = $1 AND actor_role = 'system' AND actor_id IS NULL AND reason = $2`, id, reason).
		Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

func TestBanWithdraws(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	author := insertAuthor(t, db, "ban-author@example.com", mustDaysAgo(30), true, false)
	const draft, open, review, done = "00000000-0000-7000-8000-00000000e001", "00000000-0000-7000-8000-00000000e002",
		"00000000-0000-7000-8000-00000000e003", "00000000-0000-7000-8000-00000000e004"
	insertOwned(t, db, draft, author, "draft")
	insertOwned(t, db, open, author, "submitted")
	insertOwned(t, db, review, author, "in_review")
	insertOwned(t, db, done, author, "applied")

	for range 2 { // a replayed event changes nothing more
		if err := WithdrawAuthor(context.Background(), db, author, "author_banned"); err != nil {
			t.Fatal(err)
		}
	}

	for _, id := range []string{draft, open, review} {
		assertStatus(t, db, id, "withdrawn", 2)
		if n := systemEvents(t, db, id, "author_banned"); n != 1 {
			t.Fatalf("%s system events = %d", id, n)
		}
	}
	assertStatus(t, db, done, "applied", 1)
}

func TestDeletionNullsActors(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	gone := insertAuthor(t, db, "gone@example.com", mustDaysAgo(30), true, false)
	const mine, judged = "00000000-0000-7000-8000-00000000e011", "00000000-0000-7000-8000-00000000e012"
	insertOwned(t, db, mine, gone, "submitted")
	insertOwned(t, db, judged, 0, "in_review")
	for _, q := range []string{
		`UPDATE suggestions SET author_id = NULL, assigned_to = $1, decided_by = $1 WHERE id = '` + judged + `'`,
		`INSERT INTO suggestion_events (suggestion_id, from_status, to_status, actor_id, actor_role)
VALUES ('` + judged + `', 'submitted', 'in_review', $1, 'moderator')`,
		`INSERT INTO suggestion_reviews (suggestion_id, reviewer_id, recommendation, suggestion_version)
VALUES ('` + judged + `', $1, 'approve', 1)`,
		`INSERT INTO catalog_locks (id, target_type, target_id, reason, locked_by)
VALUES ('00000000-0000-7000-8000-00000000e0ff', 'item', '` + applyItem + `', 'vandalism', $1)`,
		`INSERT INTO catalog_audit (actor_id, actor_role, source, target_type, target_id, field, before, after)
VALUES ($1, 'admin', 'admin', 'item', '` + applyItem + `', 'title', '"a"', '"b"')`,
	} {
		if _, err := db.ExecContext(context.Background(), q, gone); err != nil {
			t.Fatal(err)
		}
	}

	if err := db.Tx(context.Background(), func(tx *database.Tx) error {
		return ForgetUser(context.Background(), tx, gone)
	}); err != nil {
		t.Fatal(err)
	}

	assertStatus(t, db, mine, "withdrawn", 2)
	if n := systemEvents(t, db, mine, "author_deleted"); n != 1 {
		t.Fatalf("author_deleted events = %d", n)
	}

	var refs, audit int
	if err := db.QueryRowContext(context.Background(), `SELECT
  (SELECT count(*) FROM suggestions WHERE $1 IN (author_id, assigned_to, decided_by))
+ (SELECT count(*) FROM suggestion_events WHERE actor_id = $1)
+ (SELECT count(*) FROM suggestion_reviews WHERE reviewer_id = $1)
+ (SELECT count(*) FROM catalog_locks WHERE locked_by = $1),
  (SELECT count(*) FROM catalog_audit WHERE actor_id = $1)`, gone).Scan(&refs, &audit); err != nil {
		t.Fatal(err)
	}
	if refs != 0 {
		t.Fatalf("%d community rows still name the deleted user", refs)
	}
	// catalog_audit is append-only: its rows stay and point at the anonymized account ("deleted user").
	if audit != 1 {
		t.Fatalf("catalog_audit rows = %d, want the row kept", audit)
	}
}

func TestTargetGoneExpires(t *testing.T) {
	db := openCommunityDB(t)
	seedApplyItem(t, db)
	author := insertAuthor(t, db, "gone-target@example.com", mustDaysAgo(30), true, false)
	const open, done = "00000000-0000-7000-8000-00000000e021", "00000000-0000-7000-8000-00000000e022"
	insertOwned(t, db, open, author, "submitted")
	insertOwned(t, db, done, author, "applied")
	if _, err := db.ExecContext(context.Background(), `INSERT INTO contributor_stats (user_id, approved_total)
VALUES ($1, 4)`, author); err != nil {
		t.Fatal(err)
	}

	a, fake := newApplier(t, db)
	if err := ExpireTargetGone(context.Background(), db, a.Queue, applyItem); err != nil {
		t.Fatal(err)
	}

	assertStatus(t, db, open, "expired", 2)
	assertStatus(t, db, done, "applied", 1)
	if n := systemEvents(t, db, open, "target_gone"); n != 1 {
		t.Fatalf("target_gone events = %d", n)
	}
	var stats string
	if err := db.QueryRowContext(context.Background(), `SELECT row_to_json(c)::jsonb - 'updated_at' - 'user_id'
FROM contributor_stats c WHERE user_id = $1`, author).Scan(&stats); err != nil {
		t.Fatal(err)
	}
	if n := len(queue.Dispatched[Notify](fake)); n != 1 {
		t.Fatalf("expired notices = %d", n)
	}
	var approved int
	if err := db.QueryRowContext(context.Background(), `SELECT approved_total FROM contributor_stats WHERE user_id = $1`,
		author).
		Scan(&approved); err != nil ||
		approved != 4 {
		t.Fatalf("contributor_stats changed: approved=%d %v (%s)", approved, err, stats)
	}
}
