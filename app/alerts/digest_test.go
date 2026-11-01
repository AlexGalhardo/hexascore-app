//go:build integration

package alerts

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/schema"

	"github.com/AlexGalhardo/hexascore-app/database/migrations"
)

func openDigestDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.Open(t.Context(), pgtest.Database(t, ""), database.WithDriver(postgres.Driver()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	m, err := schema.NewMigrator(db, migrations.All())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(t.Context()); err != nil {
		t.Fatal(err)
	}

	return db
}

// TestDigestFakeClock records errors through the log handler, flushes them, and builds the 08:00 digest: top 20
// fingerprints of the last 24 h with one request id each, alert codes, failed runs; rows over 90 days are pruned.
func TestDigestFakeClock(t *testing.T) {
	db := openDigestDB(t)
	ctx := t.Context()
	sp, _ := time.LoadLocation("America/Sao_Paulo")
	digestAt := time.Date(2026, 11, 2, 8, 0, 0, 0, sp)

	now := digestAt.Add(-2 * time.Hour)
	rec := NewRecorder(db, func() time.Time { return now })
	log := slog.New(rec)

	// 25 fingerprints; fingerprint i is logged i+1 times, so 24..5 are the top 20.
	for i := range 25 {
		for range i + 1 {
			log.Error(fmt.Sprintf("failure %02d", i), "request_id", fmt.Sprintf("req-%02d", i))
		}
	}
	log.Warn("TMDB records older than 170 days", "alert", "A7", "check", "tmdb_age")
	log.Warn("plain warning, not recorded")
	if err := rec.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	// A second flush in the same hour adds to the same row.
	log.Error("failure 24", "request_id", "req-24b")
	if err := rec.Flush(ctx); err != nil {
		t.Fatal(err)
	}

	// Outside the 24 h window, and one past the 90-day retention.
	now = digestAt.Add(-30 * time.Hour)
	log.Error("old failure")
	now = digestAt.AddDate(0, 0, -91)
	log.Error("ancient failure")
	if err := rec.Flush(ctx); err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO job_runs (job, mode, status, triggered_by, started_at, finished_at)
VALUES ('fetch.wikidata', 'incremental', 'failed', 'schedule', $1, $1)`, digestAt.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	text, err := Digest(ctx, db, digestAt)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"failure 24", "26×", "failure 05", "A7: 1", "Failed job runs: 1"} {
		if !strings.Contains(text, want) {
			t.Errorf("digest lacks %q:\n%s", want, text)
		}
	}
	for _, absent := range []string{"failure 04", "old failure", "ancient failure", "plain warning"} {
		if strings.Contains(text, absent) {
			t.Errorf("digest has %q:\n%s", absent, text)
		}
	}
	if strings.Count(text, "req-") != 20 {
		t.Errorf("want one request id per listed fingerprint:\n%s", text)
	}
	if len(text) > 4096 {
		t.Errorf("digest is %d bytes, over the Telegram limit", len(text))
	}

	f := newFakeTelegram(t)
	n := newNotifier(t, f, errorChat, nil, digestAt)
	if err := SendDigest(ctx, db, n, digestAt); err != nil {
		t.Fatal(err)
	}
	drain(t, n)
	got := f.messages()
	if len(got) != 1 || got[0].ChatID != errorChat || !got[0].Silent || !strings.Contains(got[0].Text, "failure 24") {
		t.Fatalf("digest must go once to the silent chat: %+v", got)
	}

	var old int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM error_digest_daily WHERE message = 'ancient failure'`).
		Scan(&old); err != nil {
		t.Fatal(err)
	}
	if old != 0 {
		t.Error("rows older than 90 days must be pruned")
	}

	// Off when unset: no notifier, the prune still runs and nothing fails.
	if err := SendDigest(ctx, db, nil, digestAt); err != nil {
		t.Fatal(err)
	}
}

func TestDigest_Empty(t *testing.T) {
	db := openDigestDB(t)
	text, err := Digest(context.Background(), db, time.Date(2026, 11, 2, 11, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(text, "No errors") {
		t.Fatalf("empty digest: %s", text)
	}
}
