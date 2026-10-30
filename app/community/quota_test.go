//go:build integration

package community

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

func TestQuota(t *testing.T) {
	db := openCommunityDB(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 30, 23, 0, 0, 0, time.UTC)
	userID := insertAuthor(t, db, "quota@example.test", now.Add(-30*24*time.Hour), true, false)

	for i := 1; i <= 3; i++ {
		status, err := Submit(ctx, db, insertDraft(t, db, userID, 100+i), userID, 1, now)
		if err != nil || status != http.StatusSeeOther {
			t.Fatalf("submit %d: status %d err %v", i, status, err)
		}
	}
	refused := insertDraft(t, db, userID, 104)
	status, err := Submit(ctx, db, refused, userID, 1, now)
	var problem *Problem
	if status != http.StatusTooManyRequests || !errors.As(err, &problem) || problem.Type != problemBase+"suggestion-quota" || problem.RetryAfter <= 0 {
		t.Fatalf("4th: status %d err %#v", status, err)
	}
	assertStatus(t, db, refused, "draft", 1)
	if got := submittedAt(t, db, refused); got.Valid {
		t.Fatal("refused submit stored submitted_at")
	}

	midnight := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)
	status, err = Submit(ctx, db, refused, userID, 1, midnight)
	if err != nil || status != http.StatusSeeOther {
		t.Fatalf("after midnight: status %d err %v", status, err)
	}
	assertStatus(t, db, refused, "submitted", 2)
}

func TestWithdrawStillCounts(t *testing.T) {
	db := openCommunityDB(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	userID := insertAuthor(t, db, "withdraw@example.test", now.Add(-30*24*time.Hour), true, false)

	first := insertDraft(t, db, userID, 201)
	if status, err := Submit(ctx, db, first, userID, 1, now); err != nil || status != http.StatusSeeOther {
		t.Fatalf("first: status %d err %v", status, err)
	}
	if err := Transition(ctx, db, first, 2, "withdrawn", Actor{Role: "author"}); err != nil {
		t.Fatal(err)
	}
	for i := 2; i <= 3; i++ {
		status, err := Submit(ctx, db, insertDraft(t, db, userID, 200+i), userID, 1, now)
		if err != nil || status != http.StatusSeeOther {
			t.Fatalf("submit %d: status %d err %v", i, status, err)
		}
	}
	status, err := Submit(ctx, db, insertDraft(t, db, userID, 204), userID, 1, now)
	var problem *Problem
	if status != http.StatusTooManyRequests || !errors.As(err, &problem) || problem.Code != "suggestion-quota" {
		t.Fatalf("4th: status %d err %#v", status, err)
	}
}

func TestEligibility(t *testing.T) {
	db := openCommunityDB(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

	t.Run("unverified", func(t *testing.T) {
		userID := insertAuthor(t, db, "unverified@example.test", now.Add(-30*24*time.Hour), false, false)
		assertNotEligible(t, db, userID, now, nil)
	})

	t.Run("younger than 7 days", func(t *testing.T) {
		created := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
		userID := insertAuthor(t, db, "young@example.test", created, true, false)
		want := created.Add(7 * 24 * time.Hour)
		assertNotEligible(t, db, userID, now, &want)
	})

	t.Run("banned", func(t *testing.T) {
		userID := insertAuthor(t, db, "banned@example.test", now.Add(-30*24*time.Hour), true, true)
		assertNotEligible(t, db, userID, now, nil)
	})

	t.Run("blocked", func(t *testing.T) {
		userID := insertAuthor(t, db, "blocked@example.test", now.Add(-30*24*time.Hour), true, false)
		until := now.Add(30 * 24 * time.Hour)
		if _, err := db.ExecContext(ctx, `INSERT INTO contributor_stats (user_id, tier, blocked_until) VALUES ($1, 'blocked', $2)`, userID, until); err != nil {
			t.Fatal(err)
		}
		assertNotEligible(t, db, userID, now, &until)
	})
}

func TestConcurrentSubmit(t *testing.T) {
	db := openCommunityDB(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 30, 15, 0, 0, 0, time.UTC)
	userID := insertAuthor(t, db, "race@example.test", now.Add(-30*24*time.Hour), true, false)
	ids := make([]string, 20)
	for i := range ids {
		ids[i] = insertDraft(t, db, userID, 300+i)
	}

	var wg sync.WaitGroup
	statuses := make([]int, len(ids))
	errs := make([]error, len(ids))
	for i := range ids {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			statuses[i], errs[i] = Submit(ctx, db, ids[i], userID, 1, now)
		}(i)
	}
	wg.Wait()

	ok := 0
	refused := 0
	for i := range ids {
		switch statuses[i] {
		case http.StatusSeeOther:
			if errs[i] != nil {
				t.Fatalf("submit %d: %v", i, errs[i])
			}
			ok++
		case http.StatusTooManyRequests:
			var problem *Problem
			if !errors.As(errs[i], &problem) || problem.Code != "suggestion-quota" {
				t.Fatalf("submit %d: status %d err %#v", i, statuses[i], errs[i])
			}
			refused++
		default:
			t.Fatalf("submit %d: status %d err %v", i, statuses[i], errs[i])
		}
	}
	if ok != 3 || refused != 17 {
		t.Fatalf("admitted %d refused %d, want 3 and 17", ok, refused)
	}
	if got := countStatus(t, db, userID, "submitted"); got != 3 {
		t.Fatalf("submitted rows = %d, want 3", got)
	}
}

func assertNotEligible(t *testing.T, db *database.DB, userID int64, now time.Time, want *time.Time) {
	t.Helper()
	id := insertDraft(t, db, userID, int(userID)+1000)
	status, err := Submit(context.Background(), db, id, userID, 1, now)
	var problem *Problem
	if status != http.StatusForbidden || !errors.As(err, &problem) || problem.Code != "suggestion-not-eligible" {
		t.Fatalf("status %d err %#v", status, err)
	}
	if want == nil {
		if problem.EligibleAt != nil {
			t.Fatalf("eligibleAt = %s, want none", problem.EligibleAt)
		}

		return
	}
	if problem.EligibleAt == nil || !problem.EligibleAt.Equal(want.UTC()) {
		t.Fatalf("eligibleAt = %v, want %s", problem.EligibleAt, want.UTC())
	}
	assertStatus(t, db, id, "draft", 1)
}

func insertAuthor(t *testing.T, db *database.DB, email string, created time.Time, verified, banned bool) int64 {
	t.Helper()
	var verifiedAt any
	if verified {
		verifiedAt = created
	}
	var bannedAt any
	if banned {
		bannedAt = created
	}
	var id int64
	err := db.QueryRowContext(context.Background(), `INSERT INTO users
(name, email, password, email_verified_at, banned_at, created_at, updated_at)
VALUES ('Author', $1, 'x', $2, $3, $4, $4) RETURNING id`, email, verifiedAt, bannedAt, created.UTC()).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func insertDraft(t *testing.T, db *database.DB, authorID int64, n int) string {
	t.Helper()
	id := fmt.Sprintf("00000000-0000-7000-8000-%012d", n)
	_, err := db.ExecContext(context.Background(), `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, status, version)
VALUES ($1, $2, 'edit', 'item', '00000000-0000-0000-0000-0000000000aa', 'movie', 'en', '{}'::jsonb, 'draft', 1)`, id, authorID)
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func submittedAt(t *testing.T, db *database.DB, id string) sql.NullTime {
	t.Helper()
	var at sql.NullTime
	if err := db.QueryRowContext(context.Background(), `SELECT submitted_at FROM suggestions WHERE id = $1`, id).Scan(&at); err != nil {
		t.Fatal(err)
	}

	return at
}

func countStatus(t *testing.T, db *database.DB, authorID int64, status string) int {
	t.Helper()
	var n int
	if err := db.QueryRowContext(context.Background(), `SELECT count(*) FROM suggestions WHERE author_id = $1 AND status = $2`, authorID, status).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}
