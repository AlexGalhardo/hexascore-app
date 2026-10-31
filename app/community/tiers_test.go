package community

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/schema"
	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/database/migrations"
)

func TestAccuracy(t *testing.T) {
	db := openTierDB(t)
	ctx := context.Background()
	userID := insertTierAuthor(t, db, "accuracy@example.test")

	setTotals(t, db, userID, 0, 0, 0)
	assertAccuracy(t, db, userID, "0.500")

	setTotals(t, db, userID, 8, 2, 0)
	assertAccuracy(t, db, userID, "0.750")

	setTotals(t, db, userID, 0, 20, 0)
	assertAccuracy(t, db, userID, "0.045")

	if err := RebuildStats(ctx, db, time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	assertAccuracy(t, db, userID, "0.500")
}

func TestTier(t *testing.T) {
	db := openTierDB(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

	t.Run("blocked", func(t *testing.T) {
		userID := insertTierAuthor(t, db, "blocked-tier@example.test")
		insertDecisions(t, db, userID, 20, "rejected", "incorrect", now.Add(-40*24*time.Hour))
		if err := RebuildStats(ctx, db, now); err != nil {
			t.Fatal(err)
		}
		got := readTier(t, db, userID)
		if got.tier != "blocked" || got.accuracy != "0.045" || !got.blockedUntil.Equal(now.Add(blockFor)) {
			t.Fatalf("blocked = %+v", got)
		}
	})

	t.Run("nineteen stays new", func(t *testing.T) {
		userID := insertTierAuthor(t, db, "nineteen@example.test")
		insertDecisions(t, db, userID, 19, "rejected", "incorrect", now.Add(-40*24*time.Hour))
		if err := RebuildStats(ctx, db, now); err != nil {
			t.Fatal(err)
		}
		if got := readTier(t, db, userID); got.tier != "new" {
			t.Fatalf("tier = %s", got.tier)
		}
	})

	t.Run("low and standard boundary", func(t *testing.T) {
		lowID := insertTierAuthor(t, db, "low@example.test")
		insertDecisions(t, db, lowID, 1, "applied", "", now.Add(-40*24*time.Hour))
		insertDecisions(t, db, lowID, 9, "rejected", "incorrect", now.Add(-40*24*time.Hour))
		stdID := insertTierAuthor(t, db, "standard@example.test")
		insertDecisions(t, db, stdID, 1, "applied", "", now.Add(-40*24*time.Hour))
		insertDecisions(t, db, stdID, 8, "rejected", "incorrect", now.Add(-40*24*time.Hour))
		if err := RebuildStats(ctx, db, now); err != nil {
			t.Fatal(err)
		}
		if got := readTier(t, db, lowID); got.tier != "low" {
			t.Fatalf("low = %s accuracy %s", got.tier, got.accuracy)
		}
		if got := readTier(t, db, stdID); got.tier != "standard" {
			t.Fatalf("standard = %s accuracy %s", got.tier, got.accuracy)
		}
	})

	t.Run("admin block stays", func(t *testing.T) {
		userID := insertTierAuthor(t, db, "admin-block@example.test")
		insertDecisions(t, db, userID, 1, "applied", "", now.Add(-10*24*time.Hour))
		if err := RebuildStats(ctx, db, now); err != nil {
			t.Fatal(err)
		}
		until := now.Add(10 * 24 * time.Hour)
		if _, err := db.ExecContext(ctx, `UPDATE contributor_stats SET tier = 'blocked', blocked_until = $2, updated_at = $3 WHERE user_id = $1`, userID, until, now.Add(-time.Hour)); err != nil {
			t.Fatal(err)
		}
		if err := RebuildStats(ctx, db, now); err != nil {
			t.Fatal(err)
		}
		got := readTier(t, db, userID)
		if got.tier != "blocked" || !got.blockedUntil.Equal(until) {
			t.Fatalf("admin block = %+v", got)
		}
	})

	t.Run("trusted", func(t *testing.T) {
		cleanID := insertTierAuthor(t, db, "trusted@example.test")
		insertDecisions(t, db, cleanID, 50, "applied", "", now.Add(-10*24*time.Hour))
		spamID := insertTierAuthor(t, db, "spam-trust@example.test")
		insertDecisions(t, db, spamID, 50, "applied", "", now.Add(-10*24*time.Hour))
		insertDecisions(t, db, spamID, 1, "rejected", "spam", now.Add(-10*24*time.Hour))
		oldID := insertTierAuthor(t, db, "old-spam@example.test")
		insertDecisions(t, db, oldID, 50, "applied", "", now.Add(-10*24*time.Hour))
		insertDecisions(t, db, oldID, 1, "rejected", "spam", now.Add(-100*24*time.Hour))
		if err := RebuildStats(ctx, db, now); err != nil {
			t.Fatal(err)
		}
		if got := readTier(t, db, cleanID); !got.trusted || got.tier != "standard" {
			t.Fatalf("clean = %+v", got)
		}
		if got := readTier(t, db, spamID); got.trusted {
			t.Fatalf("recent spam is trusted: %+v", got)
		}
		if got := readTier(t, db, oldID); !got.trusted {
			t.Fatalf("old spam = %+v", got)
		}
	})

	t.Run("hourly", func(t *testing.T) {
		userID := insertTierAuthor(t, db, "hourly@example.test")
		disk := storage.NewMemory("")
		oldCover := "community/covers/old.webp"
		if err := storage.PutBytes(ctx, disk, oldCover, []byte("webp")); err != nil {
			t.Fatal(err)
		}
		if err := storage.PutBytes(ctx, disk, "keep.txt", []byte("x")); err != nil {
			t.Fatal(err)
		}
		oldDraft := insertTierDraft(t, db, userID, 1, "draft", fmt.Sprintf(`{"cover":{"after":%q}}`, oldCover), now.Add(-15*24*time.Hour))
		youngDraft := insertTierDraft(t, db, userID, 2, "draft", `{"cover":{"after":"https://cdn.example/a.jpg"}}`, now.Add(-13*24*time.Hour))
		idle := insertClaim(t, db, userID, 3, now.Add(-3*time.Hour))
		fresh := insertClaim(t, db, userID, 4, now.Add(-30*time.Minute))
		if err := Hourly(ctx, db, now, disk); err != nil {
			t.Fatal(err)
		}
		assertTierStatus(t, db, oldDraft, "expired")
		assertTierStatus(t, db, youngDraft, "draft")
		assertTierStatus(t, db, idle, "submitted")
		assertTierStatus(t, db, fresh, "in_review")
		if claimCleared(t, db, idle) == false {
			t.Fatal("idle claim kept its assignee")
		}
		exists, err := disk.Exists(ctx, oldCover)
		if err != nil || exists {
			t.Fatalf("cover exists %v err %v", exists, err)
		}
		kept, err := disk.Exists(ctx, "keep.txt")
		if err != nil || !kept {
			t.Fatal("unrelated file deleted")
		}
	})

	t.Run("rebuild matches live stats", func(t *testing.T) {
		userID := insertTierAuthor(t, db, "rebuild@example.test")
		insertDecisions(t, db, userID, 1, "applied", "", now.Add(-5*24*time.Hour))
		sentinel := now.Add(-2 * time.Hour)
		if _, err := db.ExecContext(ctx, `INSERT INTO contributor_stats
(user_id, submitted_total, approved_total, rejected_total, reverted_total, withdrawn_total, tier, trusted_eligible, updated_at)
VALUES ($1, 1, 1, 0, 0, 0, 'standard', false, $2)`, userID, sentinel); err != nil {
			t.Fatal(err)
		}
		before := readTier(t, db, userID)
		if err := RebuildStats(ctx, db, now); err != nil {
			t.Fatal(err)
		}
		after := readTier(t, db, userID)
		if before != after {
			t.Fatalf("rebuild changed live stats\nbefore %+v\nafter  %+v", before, after)
		}
	})
}

func TestCooldown(t *testing.T) {
	db := openTierDB(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)

	t.Run("three in 24h", func(t *testing.T) {
		userID := insertTierAuthor(t, db, "cool@example.test")
		third := now.Add(-30 * time.Minute)
		insertDecisions(t, db, userID, 1, "rejected", "incorrect", now.Add(-2*time.Hour))
		insertDecisions(t, db, userID, 1, "rejected", "no_source", now.Add(-time.Hour))
		insertDecisions(t, db, userID, 1, "rejected", "bad_source", third)
		if err := RebuildStats(ctx, db, now); err != nil {
			t.Fatal(err)
		}
		want := third.Add(48 * time.Hour)
		got := readTier(t, db, userID)
		if !got.cooldown.Equal(want) {
			t.Fatalf("cooldown = %s, want %s", got.cooldown, want)
		}
		draft := insertTierDraft(t, db, userID, 1, "draft", `{}`, now)
		status, err := Submit(ctx, db, draft, userID, 1, now)
		var problem *Problem
		if status != http.StatusTooManyRequests || !errors.As(err, &problem) || problem.Code != "suggestion-cooldown" || problem.RetryAfter <= 0 || problem.CooldownUntil == nil || !problem.CooldownUntil.Equal(want) {
			t.Fatalf("submit status %d err %#v", status, err)
		}
		assertTierStatus(t, db, draft, "draft")
	})

	t.Run("spam is 7 days", func(t *testing.T) {
		userID := insertTierAuthor(t, db, "spam-cool@example.test")
		when := now.Add(-time.Hour)
		insertDecisions(t, db, userID, 1, "rejected", "spam", when)
		if err := RebuildStats(ctx, db, now); err != nil {
			t.Fatal(err)
		}
		want := when.Add(7 * 24 * time.Hour)
		if got := readTier(t, db, userID); !got.cooldown.Equal(want) {
			t.Fatalf("cooldown = %s, want %s", got.cooldown, want)
		}
	})

	t.Run("duplicate does not cool down", func(t *testing.T) {
		userID := insertTierAuthor(t, db, "dup-cool@example.test")
		insertDecisions(t, db, userID, 3, "rejected", "duplicate", now.Add(-48*time.Hour))
		if err := RebuildStats(ctx, db, now); err != nil {
			t.Fatal(err)
		}
		if got := readTier(t, db, userID); !got.cooldown.IsZero() {
			t.Fatalf("cooldown = %s", got.cooldown)
		}
		draft := insertTierDraft(t, db, userID, 1, "draft", `{}`, now)
		status, err := Submit(ctx, db, draft, userID, 1, now)
		if err != nil || status != http.StatusSeeOther {
			t.Fatalf("submit status %d err %v", status, err)
		}
	})
}

func openTierDB(t *testing.T) *database.DB {
	t.Helper()
	if os.Getenv("TEST_DATABASE_URL") == "" && os.Getenv("POSTGRES_URL") == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
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

	return db
}

func insertTierAuthor(t *testing.T, db *database.DB, email string) int64 {
	t.Helper()
	created := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	var id int64
	err := db.QueryRowContext(context.Background(), `INSERT INTO users
(name, email, password, email_verified_at, created_at, updated_at)
VALUES ('Author', $1, 'x', $2, $2, $2) RETURNING id`, email, created).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func setTotals(t *testing.T, db *database.DB, userID int64, approved, rejected, reverted int) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `INSERT INTO contributor_stats (user_id, approved_total, rejected_total, reverted_total)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id) DO UPDATE SET approved_total = $2, rejected_total = $3, reverted_total = $4`, userID, approved, rejected, reverted)
	if err != nil {
		t.Fatal(err)
	}
}

func assertAccuracy(t *testing.T, db *database.DB, userID int64, want string) {
	t.Helper()
	var got string
	if err := db.QueryRowContext(context.Background(), `SELECT accuracy::text FROM contributor_stats WHERE user_id = $1`, userID).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("accuracy = %s, want %s", got, want)
	}
}

var tierSeq int

func insertDecisions(t *testing.T, db *database.DB, userID int64, n int, status, reason string, when time.Time) {
	t.Helper()
	for range n {
		tierSeq++
		id := fmt.Sprintf("00000000-0000-7000-8000-%012d", tierSeq)
		var reasonArg any
		if reason != "" {
			reasonArg = reason
		}
		_, err := db.ExecContext(context.Background(), `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, status, version, submitted_at, decided_at, reject_reason)
VALUES ($1, $2, 'edit', 'item', '00000000-0000-0000-0000-0000000000aa', 'movie', 'en', '{}'::jsonb, $3, 1, $4, $4, $5)`,
			id, userID, status, when.UTC(), reasonArg)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func insertTierDraft(t *testing.T, db *database.DB, userID int64, n int, status, changes string, updated time.Time) string {
	t.Helper()
	id := fmt.Sprintf("00000000-0000-7000-8100-%012d", n+int(userID))
	_, err := db.ExecContext(context.Background(), `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, status, version, updated_at)
VALUES ($1, $2, 'edit', 'item', '00000000-0000-0000-0000-0000000000aa', 'movie', 'en', $3::jsonb, $4, 1, $5)`,
		id, userID, changes, status, updated.UTC())
	if err != nil {
		t.Fatal(err)
	}

	return id
}

func insertClaim(t *testing.T, db *database.DB, userID int64, n int, claimed time.Time) string {
	t.Helper()
	id := fmt.Sprintf("00000000-0000-7000-8200-%012d", n+int(userID))
	_, err := db.ExecContext(context.Background(), `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, status, version, assigned_to, claimed_at)
VALUES ($1, $2, 'edit', 'item', '00000000-0000-0000-0000-0000000000aa', 'movie', 'en', '{}'::jsonb, 'in_review', 1, $2, $3)`,
		id, userID, claimed.UTC())
	if err != nil {
		t.Fatal(err)
	}

	return id
}

type tierRow struct {
	tier         string
	accuracy     string
	trusted      bool
	blockedUntil time.Time
	cooldown     time.Time
	updatedAt    time.Time
}

func readTier(t *testing.T, db *database.DB, userID int64) tierRow {
	t.Helper()
	var row tierRow
	var blocked, cool sql.NullTime
	err := db.QueryRowContext(context.Background(), `SELECT tier::text, accuracy::text, trusted_eligible, blocked_until, cooldown_until, updated_at
FROM contributor_stats WHERE user_id = $1`, userID).Scan(&row.tier, &row.accuracy, &row.trusted, &blocked, &cool, &row.updatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if blocked.Valid {
		row.blockedUntil = blocked.Time.UTC()
	}
	if cool.Valid {
		row.cooldown = cool.Time.UTC()
	}
	row.updatedAt = row.updatedAt.UTC()

	return row
}

func assertTierStatus(t *testing.T, db *database.DB, id, status string) {
	t.Helper()
	var got string
	if err := db.QueryRowContext(context.Background(), `SELECT status::text FROM suggestions WHERE id = $1`, id).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != status {
		t.Fatalf("suggestion %s status %s, want %s", id, got, status)
	}
}

func claimCleared(t *testing.T, db *database.DB, id string) bool {
	t.Helper()
	var assigned sql.NullInt64
	var claimed sql.NullTime
	if err := db.QueryRowContext(context.Background(), `SELECT assigned_to, claimed_at FROM suggestions WHERE id = $1`, id).Scan(&assigned, &claimed); err != nil {
		t.Fatal(err)
	}

	return !assigned.Valid && !claimed.Valid
}
