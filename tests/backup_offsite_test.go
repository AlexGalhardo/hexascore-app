//go:build integration

package tests

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/signed"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/backup"
	hmail "github.com/AlexGalhardo/hexascore-app/app/mail"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestBackupOffsiteLinkMailed covers SDD 11 §7.3 step 6: the owner gets a 7-day signed link to this month's
// encrypted monthly dump, the link serves only monthly dumps, and backup_runs records the mailing once.
func TestBackupOffsiteLinkMailed(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	ctx := t.Context()
	day1 := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	now := day1
	clock := func() time.Time { return now }

	disk := storage.NewMemory("")
	monthly := backup.ObjectKey("monthly", "test", day1.Add(-time.Hour))
	daily := backup.ObjectKey("daily", "test", day1.Add(-time.Hour))

	for _, k := range []string{monthly, daily} {
		if err := disk.Put(ctx, k, strings.NewReader("sealed frames")); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO backup_runs (id, kind, object_key, bytes, sha256, status, started_at)
VALUES ('0190f0f0-0000-7000-8000-000000000001', 'monthly', $1, 13, 'abc123', 'uploaded', $2)`,
		monthly, day1.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	signer := signed.New(bytes.Repeat([]byte{9}, 32)).WithClock(clock)

	var sent []backup.OffsiteMail

	o := &backup.Offsite{
		DB: db, Disk: disk, Signer: signer, URL: "https://hexascore.test",
		To: []string{"owner@example.test"}, Now: clock,
		Send: func(_ context.Context, m backup.OffsiteMail) error {
			sent = append(sent, m)
			return nil
		},
	}

	if err := o.Mail(ctx); err != nil {
		t.Fatalf("mail: %v", err)
	}

	if len(sent) != 1 {
		t.Fatalf("sent %d mails, want 1", len(sent))
	}

	m := sent[0]
	if len(m.To) != 1 || m.To[0] != "owner@example.test" {
		t.Fatalf("mail to %v, want ADMIN_EMAILS only", m.To)
	}

	if m.Key != monthly || m.SHA256 != "abc123" || !m.Expires.Equal(day1.Add(backup.OffsiteTTL)) {
		t.Fatalf("mail = %+v", m)
	}

	link, err := url.Parse(m.Link)
	if err != nil || link.Host != "hexascore.test" || link.Path != backup.OffsitePrefix+"/"+monthly {
		t.Fatalf("link %q (%v)", m.Link, err)
	}

	if exp, _ := strconv.ParseInt(link.Query().Get("expires"), 10, 64); exp != day1.Add(7*24*time.Hour).Unix() {
		t.Fatalf("expires = %d, want 7 days", exp)
	}

	var status string
	if err := db.QueryRowContext(ctx, `SELECT status FROM backup_runs WHERE kind = 'offsite' AND object_key = $1`,
		monthly).Scan(&status); err != nil || status != "uploaded" {
		t.Fatalf("offsite row status = %q (%v)", status, err)
	}

	mailer, err := app.Mailer()
	if err != nil {
		t.Fatal(err)
	}

	msg, err := mailer.Render(ctx, hmail.BackupOffsite{
		To: m.To, Link: m.Link, Key: m.Key, SHA256: m.SHA256, Bytes: m.Bytes, Expires: m.Expires.Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	html := strings.ReplaceAll(m.Link, "&", "&amp;")
	if !strings.Contains(msg.HTML, html) || !strings.Contains(msg.From, "<noreply@") ||
		len(msg.ReplyTo) != 1 || !strings.Contains(msg.ReplyTo[0], "support@") {
		t.Fatalf("rendered mail: from %q reply-to %v html %q", msg.From, msg.ReplyTo, msg.HTML)
	}

	// A retry the same month does not mail again.
	if err := o.Mail(ctx); err != nil || len(sent) != 1 {
		t.Fatalf("second run: %v, %d mails", err, len(sent))
	}

	mux := http.NewServeMux()
	mux.Handle("GET "+backup.OffsitePrefix+"/",
		http.StripPrefix(backup.OffsitePrefix, backup.OffsiteHandler(disk, signer)))

	get := func(target string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))

		return rec
	}

	rel := link.RequestURI()
	if rec := get(rel); rec.Code != http.StatusOK {
		t.Fatalf("signed link: %d", rec.Code)
	} else if body, _ := io.ReadAll(rec.Body); string(body) != "sealed frames" {
		t.Fatalf("body %q", body)
	}

	if rec := get(strings.Replace(rel, "signature=", "signature=0", 1)); rec.Code != http.StatusForbidden {
		t.Fatalf("tampered link: %d, want 403", rec.Code)
	}

	// The signer serves any signed path of the bucket; the handler still refuses everything but monthly dumps.
	dailyLink, err := storage.TemporaryURL(signer, backup.OffsitePrefix, daily, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	if rec := get(dailyLink); rec.Code != http.StatusNotFound {
		t.Fatalf("daily dump link: %d, want 404", rec.Code)
	}

	now = day1.Add(backup.OffsiteTTL + time.Minute)
	if rec := get(rel); rec.Code != http.StatusForbidden {
		t.Fatalf("expired link: %d, want 403", rec.Code)
	}

	// Next month without a monthly dump: an error, no mail.
	now = time.Date(2026, 12, 1, 10, 0, 0, 0, time.UTC)
	if err := o.Mail(ctx); !errors.Is(err, backup.ErrNoMonthlyDump) || len(sent) != 1 {
		t.Fatalf("no monthly dump: %v, %d mails", err, len(sent))
	}

	o.To = nil
	if err := o.Mail(ctx); err == nil {
		t.Fatal("empty ADMIN_EMAILS must fail closed")
	}
}
