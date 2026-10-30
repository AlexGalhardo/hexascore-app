//go:build integration

package tests

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	"github.com/AlexGalhardo/hexascore-app/app/jobs"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestAppTelemetry covers anonymous client-errors and client-metrics:
// 204, unknown field 400, the 11th error in a minute 429, oversized bodies
// 413, and prune at day 8 on the fake clock.
func TestAppTelemetry(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	clock := vukemtest.NewFakeClock(start)
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env:   map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
		Clock: clock.Now,
	})
	c := app.Client(t).WithHeader("Origin", "https://evil.example")

	c.PostJSON("/app/v1/client-errors", map[string]any{
		"message": "boom", "email": "a@b.test",
	}).AssertStatus(http.StatusBadRequest)
	if n := reportCount(t, app); n != 0 {
		t.Fatalf("unknown field stored %d rows", n)
	}

	c.PostJSON("/app/v1/client-errors", map[string]any{
		"appVersion":  "1.2.3",
		"versionCode": 45,
		"androidApi":  34,
		"screen":      "ItemDetail",
		"message":     "boom",
		"stack":       "main",
	}).AssertNoContent()

	assertReportShape(t, app, start)

	c.PostJSON("/app/v1/client-errors", map[string]any{
		"message": "x", "screen": "https://hexascore.app/secret",
	}).AssertStatus(http.StatusBadRequest)

	clock.Advance(6 * 24 * time.Hour)
	dispatchRetain(t, app)
	if n := reportCount(t, app); n != 1 {
		t.Fatalf("reports after 6 days = %d, want 1", n)
	}

	clock.Advance(2 * 24 * time.Hour)
	dispatchRetain(t, app)
	if n := reportCount(t, app); n != 0 {
		t.Fatalf("reports at day 8 = %d, want 0", n)
	}

	c.PostJSON("/app/v1/client-metrics", map[string]any{
		"name": "LCP", "value": 1.25, "page": "/games/{slug}", "locale": "pt-BR",
	}).AssertNoContent()
	if n := reportCount(t, app); n != 1 {
		t.Fatalf("fresh vital stored %d rows, want 1", n)
	}

	dispatchRetain(t, app)
	if n := reportCount(t, app); n != 1 {
		t.Fatalf("fresh vital after retain = %d, want 1", n)
	}

	limited := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
	})
	lc := limited.Client(t)
	body := map[string]any{"message": "boom", "screen": "Home"}
	for range 10 {
		lc.PostJSON("/app/v1/client-errors", body).AssertNoContent()
	}

	lc.PostJSON("/app/v1/client-errors", body).AssertStatus(http.StatusTooManyRequests)
	if n := reportCount(t, limited); n != 10 {
		t.Fatalf("stored errors = %d, want 10", n)
	}

	sized := vukemtest.New(t, bootstrap.New, vukemtest.Options{
		Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")},
	})
	sc := sized.Client(t)
	overErr := bytes.Repeat([]byte("x"), int(limits.ClientErrorsBody)+1)
	sc.Do(http.MethodPost, "/app/v1/client-errors", bytes.NewReader(overErr), "application/json").
		AssertStatus(http.StatusRequestEntityTooLarge)
	overMet := bytes.Repeat([]byte("x"), int(limits.ClientMetricsBody)+1)
	sc.Do(http.MethodPost, "/app/v1/client-metrics", bytes.NewReader(overMet), "application/json").
		AssertStatus(http.StatusRequestEntityTooLarge)
	if n := reportCount(t, sized); n != 0 {
		t.Fatalf("oversized bodies stored %d rows", n)
	}
}

func dispatchRetain(t *testing.T, app *vukemtest.App) {
	t.Helper()

	q, err := app.Queue()
	if err != nil {
		t.Fatal(err)
	}

	if err := q.Dispatch(t.Context(), jobs.RetainPersonalData{}); err != nil {
		t.Fatal(err)
	}
}

func reportCount(t *testing.T, app *vukemtest.App) int {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var n int
	if err := db.QueryRowContext(t.Context(), `SELECT count(*) FROM client_reports`).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

func assertReportShape(t *testing.T, app *vukemtest.App, received time.Time) {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	rows, err := db.QueryContext(t.Context(),
		`SELECT column_name FROM information_schema.columns WHERE table_name = 'client_reports'`)
	if err != nil {
		t.Fatal(err)
	}

	defer rows.Close()

	got := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}

		got[name] = true
	}

	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}

	want := []string{"id", "kind", "platform", "app_version", "route", "name", "value", "message", "stack", "received_at"}
	if len(got) != len(want) {
		t.Fatalf("columns = %v", got)
	}

	for _, name := range want {
		if !got[name] {
			t.Fatalf("missing column %s in %v", name, got)
		}
	}

	var kind, platform, route, message, version string
	var at time.Time
	err = db.QueryRowContext(t.Context(),
		`SELECT kind, platform, route, message, app_version, received_at FROM client_reports`).
		Scan(&kind, &platform, &route, &message, &version, &at)
	if err != nil {
		t.Fatal(err)
	}

	if kind != "error" || platform != "android" || route != "ItemDetail" || message != "boom" || version != "1.2.3" {
		t.Fatalf("row = %s %s %s %s %s", kind, platform, route, message, version)
	}

	if !at.Equal(received) {
		t.Fatalf("received_at = %s, want %s", at, received)
	}
}
