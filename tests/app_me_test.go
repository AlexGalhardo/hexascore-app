//go:build integration

package tests

import (
	"database/sql"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestAppMe covers GET/PATCH /app/v1/me, the export cooldown and deletion.
// A nickname someone else holds is 422 nickname-taken and does not change the
// profile. A second export inside 7 days is 429 export-cooldown with Retry-After.
func TestAppMe(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL": pgtest.Database(t, ""),
		"EXPORTS_DIR":  t.TempDir(),
	}})
	owner := makeUser(t, app, "nick-owner@example.com", "user", true)
	user := makeUser(t, app, "me-user@example.com", "user", true)
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := acc.DB.ExecContext(t.Context(),
		`UPDATE users SET username = 'taken', display_username = 'taken' WHERE id = $1`, owner.ID); err != nil {
		t.Fatal(err)
	}

	write, _, err := acc.Tokens.Create(
		t.Context(),
		user.AuthID(),
		"phone",
		[]string{"catalog:read", "profile:write"},
		time.Hour,
	)
	if err != nil {
		t.Fatal(err)
	}
	read, _, err := acc.Tokens.Create(t.Context(), user.AuthID(), "shelf", []string{"catalog:read"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	body := map[string]any{
		"name": "Changed", "locale": "en", "profilePrivate": false, "personalizationOptOut": false,
		"username": "taken",
	}
	app.Client(t).WithToken(read).PatchJSON("/app/v1/me", body).AssertForbidden()
	app.Client(t).WithToken(write).PatchJSON("/app/v1/me", body).AssertUnprocessable().AssertSee("nickname-taken")

	var name string
	var username sql.NullString
	if err := acc.DB.QueryRowContext(t.Context(), `SELECT name, username FROM users WHERE id = $1`, user.ID).
		Scan(&name, &username); err != nil {
		t.Fatal(err)
	}
	if name != "Test" || username.Valid {
		t.Fatalf("profile after taken nickname = name %q username %v", name, username.String)
	}

	// The offline queue binds its writes to this id (SDD 10 §7.3); a string, so no JS precision loss.
	app.Client(t).WithToken(write).GetJSON("/app/v1/me").AssertOK().AssertJSONPath("name", "Test").
		AssertJSONPath("id", strconv.FormatInt(user.ID, 10)).AssertJSONPath("deletionScheduledAt", nil)

	c := app.Client(t).WithToken(write)
	c.PostJSON("/app/v1/me/export", map[string]any{}).AssertStatus(http.StatusAccepted)
	again := c.PostJSON("/app/v1/me/export", map[string]any{}).AssertStatus(http.StatusTooManyRequests).
		AssertSee("export-cooldown")
	if again.Header("Retry-After") == "" {
		t.Fatal("export cooldown has no Retry-After")
	}

	c.PostJSON("/app/v1/me/deletion", map[string]any{"confirm": false}).AssertUnprocessable()
	c.PostJSON("/app/v1/me/deletion", map[string]any{"confirm": true}).AssertOK()
	if n := scheduled(t, acc, user.ID); n != 1 {
		t.Fatalf("scheduled deletions = %d, want 1", n)
	}

	fresh, _, err := acc.Tokens.Create(t.Context(), user.AuthID(), "phone-2", []string{"profile:write"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	app.Client(t).WithToken(fresh).Delete("/app/v1/me/deletion").AssertNoContent()
	if n := scheduled(t, acc, user.ID); n != 0 {
		t.Fatalf("scheduled deletions after cancel = %d, want 0", n)
	}
}

func scheduled(t *testing.T, acc *accounts.Accounts, userID int64) int {
	t.Helper()

	var n int
	if err := acc.DB.QueryRowContext(t.Context(),
		`SELECT count(*) FROM users WHERE id = $1 AND deletion_scheduled_at IS NOT NULL`, userID).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}
