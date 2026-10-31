//go:build integration

package tests

import (
	"encoding/json"
	"html"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestLGPD_ExportDeleteOptOut(t *testing.T) {
	assertPolicyCatalogues(t)

	clock := vukemtest.NewFakeClock(time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC))
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{
			"DATABASE_URL":   pgtest.Database(t, ""),
			"EXPORTS_DIR":    t.TempDir(),
			"SESSION_DRIVER": "database",
		},
		Clock: clock.Now,
	})

	for _, prefix := range []string{"", "/pt-br", "/es"} {
		page := app.Client(t).Get(prefix + "/policy").AssertOK()
		page.AssertSee("Railway").AssertSee("Grafana Cloud").AssertSee("privacy@hexascore.app")
		page.AssertInertia("Legal/Page").PropEquals("slug", "policy")
	}

	c := app.Client(t)
	register(t, app, c, "lgpd@example.com")
	c.WithHeader("Sec-Fetch-Site", "same-origin").PostForm("/profile/export", url.Values{}).AssertRedirect("/profile")

	link := exportHref(t, app, "lgpd@example.com")
	app.Client(t).Get(link).AssertOK()
	app.Client(t).Get(tamperSignature(link)).AssertStatus(http.StatusForbidden)

	c.PatchJSON("/profile/privacy", map[string]any{"profilePrivate": true, "personalizationOptOut": true}).
		AssertRedirect("/profile")
	vukemtest.AssertDatabaseHas(t, app, "users", map[string]any{
		"email": "lgpd@example.com", "profile_private": true, "personalization_opt_out": true,
	})

	c.PostJSON("/profile/deletion", map[string]any{"confirm": true}).AssertRedirect("/login")

	// The signed link dies on the app clock. Advance after the session calls: the default
	// session lifetime is 24h, so moving the clock first would log the user out.
	clock.Advance(24*time.Hour + time.Second)
	app.Client(t).Get(link).AssertStatus(http.StatusForbidden)

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if n := count(
		t, db,
		`SELECT count(*) FROM users WHERE email = $1 AND deletion_scheduled_at IS NOT NULL`,
		"lgpd@example.com",
	); n != 1 {
		t.Fatalf("scheduled deletions = %d", n)
	}
}

func TestRetentionJobs(t *testing.T) {
	clock := vukemtest.NewFakeClock(time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC))
	dir := t.TempDir()
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{
			"DATABASE_URL":   pgtest.Database(t, ""),
			"EXPORTS_DIR":    dir,
			"SESSION_DRIVER": "database",
		},
		Clock: clock.Now,
	})

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	now := clock.Now()
	putSession(t, db, "old", now.Add(-31*24*time.Hour).Unix(), "203.0.113.0")
	putSession(t, db, "young", now.Add(-24*time.Hour).Unix(), "198.51.100.0")
	putSession(t, db, "undated", 0, "192.0.2.0")

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	svc, err := library.NewExports(app.App, acc)
	if err != nil {
		t.Fatal(err)
	}

	u := makeUser(t, app, "retain@example.com", "user", true)

	id, err := svc.Request(t.Context(), u.ID)
	if err != nil {
		t.Fatal(err)
	}

	zipPath := filepath.Join(dir, strconv.FormatInt(u.ID, 10), id+".zip")
	if _, err := os.Stat(zipPath); err != nil {
		t.Fatalf("export file: %v", err)
	}

	q, err := app.Queue()
	if err != nil {
		t.Fatal(err)
	}

	if err := q.Dispatch(t.Context(), jobs.RetainPersonalData{}); err != nil {
		t.Fatal(err)
	}

	if ip, ok := sessionField(t, db, "old", "_client.ip"); ok {
		t.Fatalf("old session still has ip %v", ip)
	}

	if got := sessionFieldMust(t, db, "old", "_token"); got != "keep" {
		t.Fatalf("old session _token = %v", got)
	}

	if got := sessionFieldMust(t, db, "young", "_client.ip"); got != "198.51.100.0" {
		t.Fatalf("young ip = %v", got)
	}

	if _, ok := sessionField(t, db, "undated", "_client.ua"); ok {
		t.Fatal("undated session kept the user agent")
	}

	var status string
	if err := db.QueryRowContext(t.Context(), `SELECT status FROM data_exports WHERE id = $1`, id).
		Scan(&status); err != nil ||
		status != "ready" {
		t.Fatalf("status before 24 h = %q, %v", status, err)
	}

	clock.Advance(24 * time.Hour)

	if _, err := db.ExecContext(t.Context(),
		`INSERT INTO client_reports (kind, platform, received_at) VALUES
		('error', 'android', $1), ('error', 'android', $2)`,
		clock.Now().Add(-8*24*time.Hour), clock.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}

	if err := q.Dispatch(t.Context(), jobs.RetainPersonalData{}); err != nil {
		t.Fatal(err)
	}

	if n := count(t, db, `SELECT count(*) FROM client_reports`); n != 1 {
		t.Fatalf("client_reports = %d, want 1", n)
	}

	if err := db.QueryRowContext(t.Context(), `SELECT status FROM data_exports WHERE id = $1`, id).
		Scan(&status); err != nil ||
		status != "expired" {
		t.Fatalf("status at 24 h = %q, %v", status, err)
	}

	if _, err := os.Stat(zipPath); !os.IsNotExist(err) {
		t.Fatalf("export file still present: %v", err)
	}

	if got := sessionFieldMust(t, db, "young", "_client.ip"); got != "198.51.100.0" {
		t.Fatalf("young ip after the second run = %v", got)
	}
}

func assertPolicyCatalogues(t *testing.T) {
	t.Helper()

	need := map[string][]string{
		"en.json": {
			"Railway", "Grafana Cloud", "Resend", "Stripe", "Google", "Telegram",
			"privacy@hexascore.app", "encarregado", "locale", "theme", "consent",
			"13 years", "30 days", "7 days", "24 hours", "IGDB", "TMDB", "youtube-nocookie",
		},
		"pt-BR.json": {
			"Railway", "Grafana Cloud", "Resend", "Stripe", "Google", "Telegram",
			"privacy@hexascore.app", "encarregado", "locale", "theme", "consent",
			"13 anos", "30 dias", "7 dias", "24 horas", "IGDB", "TMDB", "youtube-nocookie",
		},
		"es.json": {
			"Railway", "Grafana Cloud", "Resend", "Stripe", "Google", "Telegram",
			"privacy@hexascore.app", "encarregado", "locale", "theme", "consent",
			"13 años", "30 días", "7 días", "24 horas", "IGDB", "TMDB", "youtube-nocookie",
		},
	}
	ban := []string{"vercel", "neon", "upstash", "passkey", "abacatepay"}

	for file, wants := range need {
		text := policyText(t, file)
		low := strings.ToLower(text)

		for _, w := range wants {
			if !strings.Contains(text, w) {
				t.Errorf("%s policy missing %q", file, w)
			}
		}

		for _, b := range ban {
			if strings.Contains(low, b) {
				t.Errorf("%s policy contains %q", file, b)
			}
		}
	}
}

func policyText(t *testing.T, file string) string {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "resources", "lang", file))
	if err != nil {
		t.Fatal(err)
	}

	var doc map[string]json.RawMessage
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("%s: %v", file, err)
	}

	return string(doc["policy"])
}

func exportHref(t *testing.T, app *vukemtest.App, email string) string {
	t.Helper()

	for _, m := range app.Mails().SentTo(email) {
		sub := exportLink.FindStringSubmatch(m.HTML)
		if sub == nil {
			continue
		}

		link := html.UnescapeString(sub[1])
		i := strings.Index(link, "/exports/")
		if i < 0 {
			t.Fatalf("export link %q has no /exports/", link)
		}

		return link[i:]
	}

	t.Fatal("no export link in the mail")

	return ""
}

func tamperSignature(link string) string {
	const key = "signature="
	i := strings.LastIndex(link, key)
	if i < 0 {
		return link + "?signature=deadbeef"
	}

	return link[:i+len(key)] + "deadbeef"
}

func putSession(t *testing.T, db *database.DB, id string, created int64, ip string) {
	t.Helper()

	payload := map[string]any{"_client.ip": ip, "_client.ua": "Agent/1", "_token": "keep"}
	if created != 0 {
		payload["_created"] = created
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(),
		`INSERT INTO sessions (id, data, expires_at, user_id) VALUES ($1, $2, $3, '1')`,
		id, raw, time.Now().Add(40*24*time.Hour).Unix()); err != nil {
		t.Fatal(err)
	}
}

func sessionField(t *testing.T, db *database.DB, id, key string) (any, bool) {
	t.Helper()

	var raw string
	if err := db.QueryRowContext(t.Context(), `SELECT data FROM sessions WHERE id = $1`, id).Scan(&raw); err != nil {
		t.Fatal(err)
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatal(err)
	}

	v, ok := data[key]

	return v, ok
}

func sessionFieldMust(t *testing.T, db *database.DB, id, key string) any {
	t.Helper()

	v, ok := sessionField(t, db, id, key)
	if !ok {
		t.Fatalf("session %s has no %s", id, key)
	}

	return v
}
