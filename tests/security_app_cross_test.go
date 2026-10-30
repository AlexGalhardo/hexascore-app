//go:build !integration

package tests

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestSecurity_A01_CrossSurfaceCredentials hits the shipped guards: an hx_app_
// token on /api/v1 is 401, an hx_live_ key on /app/v1 is 401, and the app token
// still lists its own tokens. /api/v1 is mounted only on Postgres (VF-033).
// The integration build keeps the same test in api_keys_test.go.
func TestSecurity_A01_CrossSurfaceCredentials(t *testing.T) {
	admin := postgresAdmin(t)
	if admin == "" {
		t.Skip("postgres is down: /api/v1 is not mounted; the integration build covers this test")
	}

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL": pgtest.Database(t, admin),
	}})
	email := "crossgate@example.com"
	session := app.Client(t).Inertia("")
	register(t, app, session, email)

	session.PostJSON("/profile/api/keys", map[string]any{"name": "ci"}).AssertRedirect("/profile/api")
	page := session.Get("/profile/api").AssertOK().AssertInertia("Developer/ApiKeys")
	flash, _ := page.Page()["flash"].(map[string]any)
	plain, _ := flash["newKey"].(string)
	if !api.WellFormed(plain) {
		t.Fatal("issued key is not hx_live_")
	}

	tok := issue(t, app.Client(t), email, "Pixel")
	app.Client(t).WithToken(tok.Token).GetJSON("/api/v1/search").AssertUnauthorized()
	app.Client(t).WithToken(plain).GetJSON("/app/v1/tokens").AssertUnauthorized()
	app.Client(t).WithToken(tok.Token).GetJSON("/app/v1/tokens").AssertOK()
}

// postgresAdmin is TEST_DATABASE_URL, else POSTGRES_URL, else the compose
// server on 127.0.0.1:5433 when that port accepts a connection.
func postgresAdmin(t *testing.T) string {
	t.Helper()

	if s := os.Getenv("TEST_DATABASE_URL"); s != "" {
		return s
	}

	if s := os.Getenv("POSTGRES_URL"); s != "" {
		return s
	}

	conn, err := net.DialTimeout("tcp", "127.0.0.1:5433", 300*time.Millisecond)
	if err != nil {
		return ""
	}

	_ = conn.Close()

	return "postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable"
}
