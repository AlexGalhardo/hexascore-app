//go:build integration

package community

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/schema"

	"github.com/AlexGalhardo/hexascore-app/database/migrations"
)

// sddEdges is SDD 04 §7.1. The test owns this list so a widened machine fails.
var sddEdges = map[string]bool{
	"draft→submitted":     true,
	"draft→withdrawn":     true,
	"draft→expired":       true,
	"submitted→in_review": true,
	"in_review→submitted": true,
	"submitted→approved":  true,
	"in_review→approved":  true,
	"submitted→applied":   true,
	"in_review→applied":   true,
	"approved→applied":    true,
	"approved→in_review":  true,
	"submitted→rejected":  true,
	"in_review→rejected":  true,
	"submitted→withdrawn": true,
	"in_review→withdrawn": true,
	"submitted→expired":   true,
	"in_review→expired":   true,
	"applied→reverted":    true,
}

func TestTransitions(t *testing.T) {
	db := openCommunityDB(t)
	ctx := context.Background()
	actor := Actor{Role: "system"}

	statuses := []string{
		"draft", "submitted", "in_review", "approved", "applied",
		"rejected", "reverted", "withdrawn", "expired",
	}
	pairs := 0
	for _, from := range statuses {
		for _, to := range statuses {
			pairs++
			id := fmt.Sprintf("00000000-0000-7000-8000-%012d", pairs)
			insertSuggestion(t, db, id, from)
			err := Transition(ctx, db, id, 1, to, actor)
			if sddEdges[from+"→"+to] {
				if err != nil {
					t.Fatalf("%s → %s: %v", from, to, err)
				}
				assertStatus(t, db, id, to, 2)
				continue
			}
			var te *TransitionError
			if !errors.As(err, &te) || te.Status != 409 || te.Code != "invalid-transition" || te.From != from || te.To != to {
				t.Fatalf("%s → %s: err = %#v, want 409 invalid-transition", from, to, err)
			}
			assertStatus(t, db, id, from, 1)
		}
	}
	if pairs != 81 {
		t.Fatalf("pairs = %d, want 81", pairs)
	}
}

func TestTransitionWritesEvent(t *testing.T) {
	db := openCommunityDB(t)
	ctx := context.Background()
	actor := Actor{Role: "admin"}

	t.Run("event joins the status change", func(t *testing.T) {
		id := "00000000-0000-7000-8000-0000000000a1"
		insertSuggestion(t, db, id, "draft")
		err := db.Tx(ctx, func(tx *database.Tx) error {
			if err := TransitionTx(ctx, tx, id, 1, "submitted", actor); err != nil {
				return err
			}
			if got := eventCount(t, tx, id); got != 1 {
				t.Fatalf("events inside tx = %d, want 1", got)
			}
			var from, to string
			if err := tx.QueryRowContext(ctx, `SELECT from_status, to_status FROM suggestion_events WHERE suggestion_id = $1`, id).Scan(&from, &to); err != nil {
				t.Fatal(err)
			}
			if from != "draft" || to != "submitted" {
				t.Fatalf("event %s → %s", from, to)
			}

			return errors.New("rollback")
		})
		if err == nil || err.Error() == "" {
			t.Fatal("expected rollback")
		}
		assertStatus(t, db, id, "draft", 1)
		if got := eventCount(t, db, id); got != 0 {
			t.Fatalf("events after rollback = %d", got)
		}
	})

	t.Run("version mismatch", func(t *testing.T) {
		id := "00000000-0000-7000-8000-0000000000a2"
		insertSuggestion(t, db, id, "draft")
		err := Transition(ctx, db, id, 4, "submitted", actor)
		var te *TransitionError
		if !errors.As(err, &te) || te.Status != 409 || te.Code != "stale-before" || te.Field != "version" || te.Current != "1" {
			t.Fatalf("err = %#v, want 409 stale-before version=1", err)
		}
		assertStatus(t, db, id, "draft", 1)
		if got := eventCount(t, db, id); got != 0 {
			t.Fatalf("events = %d, want 0", got)
		}
	})

	t.Run("approve without deferred work", func(t *testing.T) {
		id := "00000000-0000-7000-8000-0000000000a3"
		insertSuggestion(t, db, id, "submitted")
		if err := Approve(ctx, db, id, 1, false, actor); err != nil {
			t.Fatal(err)
		}
		assertStatus(t, db, id, "applied", 3)
		if got := eventCount(t, db, id); got != 2 {
			t.Fatalf("events = %d, want 2", got)
		}
		rows, err := db.QueryContext(ctx, `SELECT from_status, to_status FROM suggestion_events WHERE suggestion_id = $1 ORDER BY id`, id)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		want := [][2]string{{"submitted", "approved"}, {"approved", "applied"}}
		i := 0
		for rows.Next() {
			var from, to string
			if err := rows.Scan(&from, &to); err != nil {
				t.Fatal(err)
			}
			if i >= len(want) || from != want[i][0] || to != want[i][1] {
				t.Fatalf("event %d = %s → %s", i, from, to)
			}
			i++
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		if i != 2 {
			t.Fatalf("scanned %d events", i)
		}
	})
}

func openCommunityDB(t *testing.T) *database.DB {
	t.Helper()
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

func insertSuggestion(t *testing.T, db *database.DB, id, status string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `INSERT INTO suggestions
(id, kind, target_type, target_id, target_domain, locale, changes, status, version)
VALUES ($1, 'edit', 'item', '00000000-0000-0000-0000-0000000000aa', 'movie', 'en', '{}'::jsonb, $2, 1)`, id, status)
	if err != nil {
		t.Fatal(err)
	}
}

func assertStatus(t *testing.T, db *database.DB, id, status string, version int) {
	t.Helper()
	var gotStatus string
	var gotVersion int
	err := db.QueryRowContext(context.Background(), `SELECT status, version FROM suggestions WHERE id = $1`, id).Scan(&gotStatus, &gotVersion)
	if err != nil {
		t.Fatal(err)
	}
	if gotStatus != status || gotVersion != version {
		t.Fatalf("suggestion %s status=%s version=%d, want %s/%d", id, gotStatus, gotVersion, status, version)
	}
}

func eventCount(t *testing.T, q rowQuery, id string) int {
	t.Helper()
	var n int
	if err := q.QueryRowContext(context.Background(), `SELECT count(*) FROM suggestion_events WHERE suggestion_id = $1`, id).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}
