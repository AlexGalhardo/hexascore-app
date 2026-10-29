package tests

import (
	"net/url"
	"strconv"
	"testing"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/events"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

const adminEmail = "admin@example.com"

func newAdminApp(t *testing.T, fake *events.Fake) *vukemtest.App {
	t.Helper()

	opts := vukemtest.Options{}
	if fake != nil {
		opts.AppOptions = []vukem.Option{vukem.WithEventsFake(fake)}
	}

	return vukemtest.New(t, bootstrap.New, opts)
}

// makeUser inserts a user with the given role and email verification state.
func makeUser(t *testing.T, app *vukemtest.App, email, role string, verified bool) models.User {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	u := models.User{Name: "Test", Email: email, Password: "unused", Role: role}
	u.EmailVerifiedAt.V, u.EmailVerifiedAt.Valid = time.Now().UTC(), verified

	if err := orm.Create(t.Context(), db, &u); err != nil {
		t.Fatal(err)
	}

	// Admins need confirmed 2FA to reach /admin (SDD 13 FR-13-5, AdminGuard).
	if role == models.RoleAdmin {
		ph := db.Dialect().Placeholder
		if _, err := db.ExecContext(
			t.Context(),
			"UPDATE users SET two_factor_confirmed_at = "+ph(
				1,
			)+" WHERE id = "+ph(
				2,
			),
			time.Now().UTC(),
			u.ID,
		); err != nil {
			t.Fatal(err)
		}
	}

	return u
}

// browserFor returns a same-origin browser client logged in as u.
func browserFor(t *testing.T, app *vukemtest.App, u models.User) *vukemtest.Client {
	t.Helper()

	return app.Client(t).WithHeader("Sec-Fetch-Site", "same-origin").ActingAs("web", u)
}

func reload(t *testing.T, app *vukemtest.App, id int64) models.User {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	u, err := orm.QueryOf[models.User](db).Find(t.Context(), id)
	if err != nil {
		t.Fatal(err)
	}

	return u
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

func TestAdminPanelAccess(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t, nil)

	app.Client(t).Get("/admin").AssertRedirect("/login")

	plain := makeUser(t, app, "user@example.com", "user", true)
	browserFor(t, app, plain).Get("/admin").AssertNotFound()

	unverified := makeUser(t, app, "unverified@example.com", "admin", false)
	browserFor(t, app, unverified).Get("/admin").AssertNotFound()

	admin := makeUser(t, app, adminEmail, "admin", true)
	c := browserFor(t, app, admin)
	c.Get("/admin").AssertOK().AssertSee("Dashboard")
	c.Get("/admin/users").AssertOK().AssertSee("user@example.com")
}

func TestSecurity_A01_AdminRoleNotEditableInPanel(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t, nil)
	admin := makeUser(t, app, adminEmail, "admin", true)
	target := makeUser(t, app, "target@example.com", "user", true)
	c := browserFor(t, app, admin)

	id := itoa(target.ID)
	c.PostForm("/admin/users/"+id, url.Values{
		"role": {"admin"}, "email": {"evil@example.com"}, "password": {"x"}, "banned_at": {"2026-01-01T00:00"},
	})

	got := reload(t, app, target.ID)
	if got.Role != "user" || got.Email != "target@example.com" || got.Password != "unused" || got.BannedAt.Valid {
		t.Fatalf("crafted POST changed the user: %+v", got)
	}

	// Reason and expiry only make sense on a ban.
	c.PostForm("/admin/users/"+id, url.Values{"banReason": {"spam"}})

	if reload(t, app, target.ID).BanReason != nil {
		t.Fatal("ban reason stored on a user who is not banned")
	}

	// Panel creation and deletion are denied.
	c.PostForm("/admin/users", url.Values{"banReason": {"x"}}).AssertForbidden()
	c.PostForm("/admin/users/"+id+"/delete", url.Values{}).AssertForbidden()

	// A non-admin never reaches the panel or the ban action.
	plain := browserFor(t, app, target)
	plain.PostForm("/admin/users/"+id, url.Values{"banReason": {"x"}}).AssertNotFound()
	plain.PostForm("/admin/users/"+itoa(admin.ID)+"/ban", url.Values{}).AssertNotFound()

	if reload(t, app, admin.ID).BannedAt.Valid {
		t.Fatal("non-admin banned an admin")
	}
}

func TestSecurity_A07_Ban(t *testing.T) {
	t.Parallel()

	fake := events.NewFake()
	app := newAdminApp(t, fake)
	admin := makeUser(t, app, adminEmail, "admin", true)
	adminC := browserFor(t, app, admin)

	// A real account with a password and a live session.
	victim := app.Client(t).Inertia("")
	register(t, app, victim, "victim@example.com")
	victim.Get("/dashboard").AssertOK()

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	v, err := acc.Users.ByCredentials(t.Context(), map[string]string{"email": "victim@example.com"})
	if err != nil {
		t.Fatal(err)
	}

	id := itoa(v.ID)
	adminC.PostForm("/admin/users/"+id+"/ban", url.Values{"reason": {"spam"}}).AssertRedirect("/admin/users/" + id)

	got := reload(t, app, v.ID)
	if !got.BannedAt.Valid || got.BanReason == nil || *got.BanReason != "spam" {
		t.Fatalf("ban not stored: %+v", got)
	}

	if n := len(events.Dispatched[accounts.UserBanned](fake)); n != 1 {
		t.Fatalf("user.banned dispatched %d times, want 1", n)
	}

	// The ban ended the live session ...
	victim.Get("/dashboard").AssertRedirect("/login")

	// ... and the correct password no longer logs in.
	victim.PostJSON("/login", map[string]any{"identifier": "victim@example.com", "password": testPassword}).
		AssertStatus(403).AssertSee("account-banned").AssertSee("spam")

	// Wrong passwords still reveal nothing about the ban.
	victim.PostJSON("/login", map[string]any{"identifier": "victim@example.com", "password": "wrong"}).
		AssertDontSee("account-banned")

	// Banning twice, banning yourself and banning an admin are refused.
	adminC.PostForm("/admin/users/"+id+"/ban", url.Values{}).AssertStatus(409)
	adminC.PostForm("/admin/users/"+itoa(admin.ID)+"/ban", url.Values{}).AssertForbidden()

	other := makeUser(t, app, "admin2@example.com", "admin", true)
	adminC.PostForm("/admin/users/"+itoa(other.ID)+"/ban", url.Values{}).AssertForbidden()

	// The ban is in the user's history, with the actor.
	assertEvent(t, app, v.ID, "banned", admin.ID)

	adminC.PostForm("/admin/users/"+id+"/unban", url.Values{}).AssertRedirect("/admin/users/" + id)

	if n := len(events.Dispatched[accounts.UserUnbanned](fake)); n != 1 {
		t.Fatalf("user.unbanned dispatched %d times, want 1", n)
	}

	adminC.PostForm("/admin/users/"+id+"/unban", url.Values{}).AssertStatus(409)
	victim.PostJSON("/login", map[string]any{"identifier": "victim@example.com", "password": testPassword}).
		AssertRedirect("/onboarding")
}

// A session that survives a ban (here the ban was written by plain SQL, so
// there was no session version bump) is stopped by the NotBanned middleware.
func TestSecurity_A07_BanSurvivorsAnswer403(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t, nil)
	u := makeUser(t, app, "survivor@example.com", "user", true)
	c := browserFor(t, app, u)
	c.Get("/dashboard").AssertOK()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	set := orm.Set{"banned_at": time.Now().UTC().Add(-time.Minute), "ban_reason": "sql"}
	if _, err := orm.QueryOf[models.User](db).Where("id", u.ID).Update(t.Context(), set); err != nil {
		t.Fatal(err)
	}

	c.Get("/dashboard").AssertStatus(403).AssertSee("account-banned")
	c.Get("/").AssertStatus(403)

	// An expired ban no longer counts.
	expired := orm.Set{"ban_expires_at": time.Now().UTC().Add(-time.Second)}
	if _, err := orm.QueryOf[models.User](db).Where("id", u.ID).Update(t.Context(), expired); err != nil {
		t.Fatal(err)
	}

	c.Get("/dashboard").AssertOK()
}

func assertEvent(t *testing.T, app *vukemtest.App, userID int64, kind string, actor int64) {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	var n int
	if err := db.QueryRowContext(t.Context(),
		"SELECT count(*) FROM account_events WHERE user_id = ? AND kind = ? AND actor_id = ?", userID, kind, actor,
	).Scan(&n); err != nil || n != 1 {
		t.Fatalf("account_events %s: count %d err %v", kind, n, err)
	}
}

func TestAdminCLI(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t, nil)
	u := makeUser(t, app, "grantee@example.com", "user", true)

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	if changed, err := acc.SetRole(t.Context(), "Grantee@example.com", "admin"); err != nil || !changed {
		t.Fatalf("grant: %v %v", changed, err)
	}

	if changed, _ := acc.SetRole(t.Context(), "grantee@example.com", "admin"); changed {
		t.Fatal("second grant reported a change")
	}

	granted := reload(t, app, u.ID)
	if granted.Role != "admin" {
		t.Fatal("role not stored")
	}

	// A role change ends the user's sessions (no Regenerate from the CLI).
	if granted.SessionVersion <= u.SessionVersion {
		t.Fatalf("session version %d not bumped from %d", granted.SessionVersion, u.SessionVersion)
	}

	// A fresh admin has no 2FA yet: the panel sends them to the setup first.
	browserFor(t, app, granted).Get("/admin").AssertRedirect("/user/two-factor")

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.ExecContext(t.Context(),
		"UPDATE users SET two_factor_confirmed_at = ? WHERE id = ?", time.Now().UTC(), u.ID); err != nil {
		t.Fatal(err)
	}

	browserFor(t, app, reload(t, app, u.ID)).Get("/admin").AssertOK()

	if _, err := acc.SetRole(t.Context(), "grantee@example.com", "user"); err != nil {
		t.Fatal(err)
	}

	browserFor(t, app, reload(t, app, u.ID)).Get("/admin").AssertNotFound()

	var n int
	if err := db.QueryRowContext(t.Context(),
		`SELECT count(*) FROM account_events WHERE user_id = ? AND actor_id IS NULL AND data LIKE '%"cli":true%'`, u.ID,
	).Scan(&n); err != nil || n != 2 {
		t.Fatalf("cli events: %d %v", n, err)
	}
}

func TestResetTwoFactorCLI(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t, nil)
	c := app.Client(t).Inertia("")
	register(t, app, c, "twofa@example.com")

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	u, err := acc.Users.ByCredentials(t.Context(), map[string]string{"email": "twofa@example.com"})
	if err != nil {
		t.Fatal(err)
	}

	if err := acc.EnableTwoFactor(t.Context(), u.ID); err != nil {
		t.Fatal(err)
	}

	if err := acc.ResetTwoFactor(t.Context(), "twofa@example.com", "id checked by phone"); err != nil {
		t.Fatal(err)
	}

	tf, err := acc.TwoFactor(t.Context(), u.ID)
	if err != nil || tf.Pending() || tf.Enabled() {
		t.Fatalf("2FA not cleared: %+v %v", tf, err)
	}

	c.Get("/dashboard").AssertRedirect("/login")

	if len(app.Mails().SentTo("twofa@example.com")) == 0 {
		t.Fatal("no mail sent")
	}

	if err := acc.ResetTwoFactor(t.Context(), "nobody@example.com", "x"); err == nil {
		t.Fatal("unknown email accepted")
	}
}

func TestAdminEditsBanReason(t *testing.T) {
	t.Parallel()

	app := newAdminApp(t, nil)
	admin := makeUser(t, app, adminEmail, "admin", true)
	target := makeUser(t, app, "target@example.com", "user", true)

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	if err := acc.Ban(t.Context(), admin.ID, target.ID, "first", nil); err != nil {
		t.Fatal(err)
	}

	c := browserFor(t, app, admin)
	c.PostForm("/admin/users/"+itoa(target.ID), url.Values{"banReason": {"second"}})

	got := reload(t, app, target.ID)
	if got.BanReason == nil || *got.BanReason != "second" || !got.BannedAt.Valid || got.BanExpiresAt.Valid {
		t.Fatalf("edit not applied cleanly: %+v", got)
	}
}
