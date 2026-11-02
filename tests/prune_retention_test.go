//go:build integration

package tests

import (
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/config"
)

// TestPruneRetention_DeletesOnlyPastTheWindow pins audit T-17: rows older than the window go, in batches, newer
// rows stay, and the ingest_failures window only applies to parked rows.
func TestPruneRetention_DeletesOnlyPastTheWindow(t *testing.T) {
	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	exec := func(q string, args ...any) {
		t.Helper()

		if _, err := db.ExecContext(t.Context(), q, args...); err != nil {
			t.Fatal(err)
		}
	}

	// 7 old and 2 recent contact messages; a batch of 3 needs three rounds plus the short last one.
	exec(`INSERT INTO contact_messages (id, name, email, subject, message, created_at)
SELECT gen_random_uuid(), 'n', 'e@example.test', 'question', repeat('m', 40), now() - interval '200 days' FROM generate_series(1, 7)`)
	exec(`INSERT INTO contact_messages (id, name, email, subject, message, created_at)
SELECT gen_random_uuid(), 'n', 'e@example.test', 'question', repeat('m', 40), now() - interval '10 days' FROM generate_series(1, 2)`)

	window := config.Retention{Table: "contact_messages", Column: "created_at", Days: 180}
	n, err := jobs.PruneTable(t.Context(), db, window, 3)
	if err != nil || n != 7 {
		t.Fatalf("pruned %d, %v; want 7 old rows", n, err)
	}

	var left int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM contact_messages`).
		Scan(&left); err != nil ||
		left != 2 {
		t.Fatalf("left %d, %v; want the 2 recent rows", left, err)
	}

	exec(`INSERT INTO ingest_failures (source, external_id, error_class, error, last_run_id, parked, last_seen)
VALUES ('tmdb', 'a', 'x', 'e', 1, true, now() - interval '100 days'),
       ('tmdb', 'b', 'x', 'e', 1, false, now() - interval '100 days')`)

	windows, err := config.RetentionWindows(func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}

	for _, w := range windows {
		if w.Table != "ingest_failures" {
			continue
		}

		if n, err := jobs.PruneTable(t.Context(), db, w, jobs.PruneBatch); err != nil || n != 1 {
			t.Fatalf("ingest_failures pruned %d, %v; want only the parked row", n, err)
		}
	}
}

func TestRetentionWindows_Overrides(t *testing.T) {
	got, err := config.RetentionWindows(func(k string) string {
		if k == "RETENTION_CONTACT_MESSAGES_DAYS" {
			return "30"
		}

		return ""
	})
	if err != nil || got[0].Table != "contact_messages" || got[0].Days != 30 {
		t.Fatalf("got %+v, %v; want contact_messages at 30 days", got[0], err)
	}

	if _, err := config.RetentionWindows(func(string) string { return "0" }); err == nil {
		t.Fatal("a zero window was accepted")
	}
}
