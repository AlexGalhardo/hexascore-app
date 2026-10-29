//go:build integration

package tests

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

var exportLink = regexp.MustCompile(`href="([^"]*/exports/[^"]+)"`)

func TestExportZip(t *testing.T) {
	clock := vukemtest.NewFakeClock(time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC))
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env:   map[string]string{"DATABASE_URL": pgtest.Database(t, ""), "EXPORTS_DIR": t.TempDir()},
		Clock: clock.Now,
	})

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	u := makeUser(t, app, "exporter@example.com", "user", true)
	other := makeUser(t, app, "other@example.com", "user", true)
	plain, _, err := acc.Tokens.Create(t.Context(), u.AuthID(), "phone", []string{"profile:write"}, 30*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	exec := func(q string, args ...any) {
		t.Helper()

		if _, err := acc.DB.ExecContext(t.Context(), q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	exec(`INSERT INTO data_sources (code, name, licence, url) VALUES ('hexascore', 'x', 'proprietary', 'https://x')`)

	item := uuid.NewV7().String()
	exec(`INSERT INTO items (id, type, slug, wikidata_qid) VALUES ($1, 'game', 'formula', 'Q1')`, item)
	exec(
		`INSERT INTO item_translations (item_id, locale, title, title_source) VALUES ($1, 'en', '=HYPERLINK("x")', 'hexascore')`,
		item,
	)
	exec(
		`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6) VALUES ($1, $2, 80, 80, 80, 80, 80, 85)`,
		u.ID,
		item,
	)
	exec(
		`INSERT INTO ratings (user_id, item_id, c1, c2, c3, c4, c5, c6) VALUES ($1, $2, 10, 10, 10, 10, 10, 10)`,
		other.ID,
		item,
	)

	list := uuid.NewV7().String()
	exec(
		`INSERT INTO lists (id, user_id, type, title, is_public) VALUES ($1, $2, 'game', '@cmd list', true)`,
		list,
		u.ID,
	)
	exec(`INSERT INTO list_items (list_id, item_id, position, note) VALUES ($1, $2, 1, '+SUM(A1)')`, list, item)

	c := app.Client(t).WithToken(plain)
	c.PostJSON("/app/v1/me/export", map[string]any{}).AssertStatus(http.StatusAccepted).AssertSee(`"status":"queued"`)

	sent := app.Mails().SentTo(u.Email)
	if len(sent) != 1 {
		t.Fatalf("%d export mails, want 1", len(sent))
	}

	m := exportLink.FindStringSubmatch(sent[0].HTML)
	if m == nil {
		t.Fatalf("no export link in mail: %s", sent[0].HTML)
	}

	link := html.UnescapeString(m[1])
	link = link[strings.Index(link, "/exports/"):]

	body := app.Client(t).Get(link).AssertOK().Body()

	zr, err := zip.NewReader(bytes.NewReader([]byte(body)), int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}

	files := map[string]string{}
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}

		var b bytes.Buffer
		_, _ = b.ReadFrom(rc)
		rc.Close()
		files[f.Name] = b.String()
	}

	want := []string{
		"README.md", "account.json", "editions.json", "lists.csv", "lists.json", "notifications.json",
		"profile.json", "ratings.csv", "ratings.json", "statuses.json", "suggestions.json", "time_to_beat.json",
	}
	got := make([]string, 0, len(files))
	for name := range files {
		got = append(got, name)
	}
	slices.Sort(got)

	if !slices.Equal(got, want) {
		t.Fatalf("zip files = %v, want %v", got, want)
	}

	ratings, err := csv.NewReader(strings.NewReader(files["ratings.csv"])).ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(ratings) != 2 || ratings[1][1] != `'=HYPERLINK("x")` || ratings[1][5] != "8.0" || ratings[1][10] != "8.5" {
		t.Errorf("ratings.csv = %q, want only the user's rating with the formula guard", ratings)
	}

	lists, err := csv.NewReader(strings.NewReader(files["lists.csv"])).ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if len(lists) != 2 || lists[1][0] != "'@cmd list" || lists[1][5] != "'+SUM(A1)" {
		t.Errorf("lists.csv = %q, want guarded list title and note", lists)
	}

	for _, secret := range []string{"unused", `"password"`, "hash", other.Email} {
		for name, content := range files {
			if name != "README.md" && strings.Contains(content, secret) {
				t.Errorf("%s contains %q", name, secret)
			}
		}
	}

	if !strings.Contains(files["account.json"], `"phone"`) {
		t.Errorf("account.json = %s, want the token name", files["account.json"])
	}

	var events int
	if err := acc.DB.QueryRowContext(
		t.Context(),
		`SELECT count(*) FROM account_events WHERE user_id = $1 AND kind = 'exported'`,
		u.ID,
	).Scan(&events); err != nil {
		t.Fatal(err)
	}

	if events != 1 {
		t.Errorf("%d exported account events, want 1", events)
	}

	// A second request inside 7 days is refused with Retry-After; the web route shares the rule.
	clock.Advance(6 * 24 * time.Hour)
	again := c.PostJSON("/app/v1/me/export", map[string]any{}).AssertStatus(http.StatusTooManyRequests).
		AssertSee("export-cooldown")
	if again.Header("Retry-After") == "" {
		t.Error("cooldown answer has no Retry-After")
	}

	browserFor(t, app, u).PostForm("/profile/export", url.Values{}).AssertStatus(http.StatusTooManyRequests)

	// The link dies after 24 h (it is already 6 days old here); a tampered path never verifies.
	app.Client(t).Get(link).AssertStatus(http.StatusForbidden)
	app.Client(t).Get(strings.Replace(link, ".zip", "x.zip", 1)).AssertStatus(http.StatusForbidden)

	clock.Advance(24 * time.Hour)
	c.PostJSON("/app/v1/me/export", map[string]any{}).AssertStatus(http.StatusAccepted)
}

func TestExportExpirySweep(t *testing.T) {
	clock := vukemtest.NewFakeClock(time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC))
	dir := t.TempDir()
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env:   map[string]string{"DATABASE_URL": pgtest.Database(t, ""), "EXPORTS_DIR": dir},
		Clock: clock.Now,
	})

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	svc, err := library.NewExports(app.App, acc)
	if err != nil {
		t.Fatal(err)
	}

	u := makeUser(t, app, "sweep@example.com", "user", true)
	id, err := svc.Request(t.Context(), u.ID)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, strconv.FormatInt(u.ID, 10), id+".zip")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("export file missing after the job: %v", err)
	}

	clock.Advance(23 * time.Hour)
	if n, err := svc.Expire(t.Context()); err != nil || n != 0 {
		t.Fatalf("Expire before 24 h = %d, %v; want 0", n, err)
	}

	clock.Advance(time.Hour)
	if n, err := svc.Expire(t.Context()); err != nil || n != 1 {
		t.Fatalf("Expire at 24 h = %d, %v; want 1", n, err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("file still on disk: %v", err)
	}

	var status string
	if err := acc.DB.QueryRowContext(t.Context(), `SELECT status FROM data_exports WHERE id = $1`, id).
		Scan(&status); err != nil ||
		status != "expired" {
		t.Errorf("status = %q, %v; want expired", status, err)
	}
}

func TestExportPurgeRemovesRowsAndFiles(t *testing.T) {
	dir := t.TempDir()
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{"DATABASE_URL": pgtest.Database(t, ""), "EXPORTS_DIR": dir},
	})

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	svc, err := library.NewExports(app.App, acc)
	if err != nil {
		t.Fatal(err)
	}

	u := makeUser(t, app, "purge@example.com", "user", true)
	other := makeUser(t, app, "keep@example.com", "user", true)

	id, err := svc.Request(t.Context(), u.ID)
	if err != nil {
		t.Fatal(err)
	}

	keep, err := svc.Request(t.Context(), other.ID)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, strconv.FormatInt(u.ID, 10), id+".zip")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("export file missing before purge: %v", err)
	}

	if err := svc.Purge(t.Context(), u.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("file survived the purge: %v", err)
	}

	var n int
	if err := acc.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM data_exports WHERE user_id = $1`, u.ID).
		Scan(&n); err != nil || n != 0 {
		t.Errorf("%d rows left for the purged user, %v", n, err)
	}

	if err := acc.DB.QueryRowContext(t.Context(), `SELECT count(*) FROM data_exports WHERE id = $1`, keep).
		Scan(&n); err != nil || n != 1 {
		t.Errorf("other user's export rows = %d, %v; want 1", n, err)
	}
}

// purgeOnPut runs purge just before the file is written, like an account purge landing mid-build.
type purgeOnPut struct {
	storage.Disk
	purge func()
}

func (d purgeOnPut) Put(ctx context.Context, p string, r io.Reader) error {
	d.purge()
	return d.Disk.Put(ctx, p, r)
}

func TestExportBuildAfterPurgeLeavesNoFile(t *testing.T) {
	dir := t.TempDir()
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{"DATABASE_URL": pgtest.Database(t, ""), "EXPORTS_DIR": dir},
	})

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	svc, err := library.NewExports(app.App, acc)
	if err != nil {
		t.Fatal(err)
	}

	u := makeUser(t, app, "race@example.com", "user", true)
	id := uuid.NewV7().String()

	if _, err := acc.DB.ExecContext(t.Context(),
		`INSERT INTO data_exports (id, user_id, status) VALUES ($1, $2, 'queued')`, id, u.ID); err != nil {
		t.Fatal(err)
	}

	svc.Disk = purgeOnPut{Disk: svc.Disk, purge: func() {
		if err := svc.Purge(t.Context(), u.ID); err != nil {
			t.Fatal(err)
		}
	}}

	if err := svc.Build(t.Context(), id); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(dir, strconv.FormatInt(u.ID, 10), id+".zip")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("export file survived a purge during the build: %v", err)
	}
}
