package tests

import (
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func profileUser(t *testing.T, email string) (*vukemtest.App, *vukemtest.Client) {
	t.Helper()

	app := vukemtest.New(t, bootstrap.New)
	c := app.Client(t).Inertia("")
	register(t, app, c, email)

	return app, c
}

func TestProfileDetails(t *testing.T) {
	t.Parallel()

	app, c := profileUser(t, "details@example.com")

	c.Get("/profile").AssertOK().AssertInertia("Profile/Index").
		PropEquals("details.name", "Ana Silva").PropEquals("details.locale", "en").
		PropEquals("passwordSet", true)

	c.PatchJSON("/profile/details", map[string]any{"name": "Ana Souza", "locale": "pt-BR"}).AssertRedirect("/profile")
	vukemtest.AssertDatabaseHas(
		t,
		app,
		"users",
		map[string]any{"email": "details@example.com", "name": "Ana Souza", "locale": "pt-BR"},
	)

	c.WithHeader("Referer", "/profile").PatchJSON("/profile/details", map[string]any{"name": "Ana", "locale": "fr"}).
		AssertRedirect("/profile")
	c.Get("/profile").AssertInertia("Profile/Index").PropEquals("details.locale", "pt-BR").HasProp("errors.locale")

	app.Client(t).Inertia("").PatchJSON("/profile/details", map[string]any{"name": "x", "locale": "en"}).
		AssertRedirect("/login")
}

func TestProfilePassword(t *testing.T) {
	t.Parallel()

	const next = "Another horse 7 staple!"

	app, c := profileUser(t, "pw@example.com")
	other := app.Client(t).Inertia("")
	other.PostJSON("/login", map[string]any{"identifier": "pw@example.com", "password": testPassword})
	other.Get("/profile").AssertOK()

	c.WithHeader("Referer", "/profile").
		PutJSON("/profile/password", map[string]any{"currentPassword": "wrong", "newPassword": next}).
		AssertRedirect("/profile")
	c.Get("/profile").AssertInertia("Profile/Index").HasProp("errors.currentPassword")
	c.WithHeader("Referer", "/profile").
		PutJSON("/profile/password", map[string]any{"currentPassword": testPassword, "newPassword": "weak"}).
		AssertRedirect("/profile")
	c.Get("/profile").AssertInertia("Profile/Index").HasProp("errors.newPassword")

	c.PutJSON("/profile/password", map[string]any{"currentPassword": testPassword, "newPassword": next}).
		AssertRedirect("/profile")
	vukemtest.AssertDatabaseHas(t, app, "account_events", map[string]any{"kind": "password_changed"})

	// This session survives; every other one ended.
	c.Get("/profile").AssertOK().AssertInertia("Profile/Index")
	other.Get("/profile").AssertRedirect("/login")

	fresh := app.Client(t).Inertia("")
	fresh.PostJSON("/login", map[string]any{"identifier": "pw@example.com", "password": next}).
		AssertRedirect("/onboarding")
}

func TestProfilePassword_SetFirstTime(t *testing.T) {
	t.Parallel()

	const next = "Another horse 7 staple!"

	app, c := profileUser(t, "google@example.com")
	clearPasswordSet(t, app, "google@example.com")
	linkGoogle(t, app, "google@example.com")

	c.Get("/profile").AssertInertia("Profile/Index").PropEquals("passwordSet", false)
	// No current password is asked when none was ever chosen.
	c.PutJSON("/profile/password", map[string]any{"newPassword": next}).AssertRedirect("/profile")
	c.Get("/profile").AssertInertia("Profile/Index").PropEquals("passwordSet", true)
}

// TestSecurity_A07_PasswordChangeNeedsCurrentUnlessSocialOnly: an account
// without password_set_at but also without a linked social account (legacy or
// imported) must still prove the current password.
func TestSecurity_A07_PasswordChangeNeedsCurrentUnlessSocialOnly(t *testing.T) {
	t.Parallel()

	app, c := profileUser(t, "legacy@example.com")
	clearPasswordSet(t, app, "legacy@example.com")

	c.Get("/profile").AssertInertia("Profile/Index").PropEquals("passwordSet", true)
	c.WithHeader("Referer", "/profile").
		PutJSON("/profile/password", map[string]any{"newPassword": "Another horse 7 staple!"}).
		AssertRedirect("/profile")
	c.Get("/profile").AssertInertia("Profile/Index").HasProp("errors.currentPassword")
}

func linkGoogle(t *testing.T, app *vukemtest.App, email string) {
	t.Helper()

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	p := acc.DB.Dialect().Placeholder
	if _, err := acc.DB.ExecContext(t.Context(),
		"INSERT INTO social_accounts (user_id, provider, provider_id, created_at, updated_at) SELECT id, 'google', "+
			p(1)+", "+p(2)+", "+p(3)+" FROM users WHERE email = "+p(4),
		"g-"+email, time.Now().UTC(), time.Now().UTC(), email); err != nil {
		t.Fatal(err)
	}
}

func TestProfilePrivacy(t *testing.T) {
	t.Parallel()

	app, c := profileUser(t, "privacy@example.com")

	c.PatchJSON("/profile/privacy", map[string]any{"profilePrivate": true, "personalizationOptOut": true}).
		AssertRedirect("/profile")
	vukemtest.AssertDatabaseHas(t, app, "users", map[string]any{
		"email": "privacy@example.com", "profile_private": true, "personalization_opt_out": true,
	})
	c.Get("/profile").AssertInertia("Profile/Index").
		PropEquals("privacy.profilePrivate", true).PropEquals("privacy.personalizationOptOut", true)
}

// The community mail switch (SDD 04 FR-04-14) has its own form, so saving the privacy switches, or the app's
// PATCH /app/v1/me, never resets it.
func TestProfileCommunityMailOptOut(t *testing.T) {
	t.Parallel()

	app, c := profileUser(t, "mail-optout@example.com")

	c.Get("/profile").AssertInertia("Profile/Index").PropEquals("notifications.communityMailOptOut", false)
	c.PatchJSON("/profile/notifications", map[string]any{"communityMailOptOut": true}).AssertRedirect("/profile")
	c.PatchJSON("/profile/privacy", map[string]any{"profilePrivate": true}).AssertRedirect("/profile")
	vukemtest.AssertDatabaseHas(t, app, "users", map[string]any{
		"email": "mail-optout@example.com", "community_mail_opt_out": true, "profile_private": true,
	})
	c.Get("/profile").AssertInertia("Profile/Index").PropEquals("notifications.communityMailOptOut", true)
}

func TestSessionsRevoke(t *testing.T) {
	t.Parallel()

	// Listing needs server-side sessions; cookie sessions have nothing to list.
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"SESSION_DRIVER": "database"}})
	c := app.Client(t).Inertia("")
	register(t, app, c, "sessions@example.com")
	other := app.Client(t).Inertia("")
	other.PostJSON("/login", map[string]any{"identifier": "sessions@example.com", "password": testPassword})

	page := c.Get("/profile").AssertInertia("Profile/Index").HasProp("sessions", 2)

	var handle string
	for i := range 2 {
		cur, _ := page.Prop("sessions." + string(rune('0'+i)) + ".current")
		if cur == false {
			h, _ := page.Prop("sessions." + string(rune('0'+i)) + ".handle")
			handle, _ = h.(string)
		}
	}

	if handle == "" {
		t.Fatal("no other session listed")
	}

	// Another user cannot end this user's session.
	stranger := app.Client(t).Inertia("")
	register(t, app, stranger, "stranger@example.com")
	stranger.Delete("/profile/sessions/" + handle).AssertStatus(404)
	other.Get("/profile").AssertOK()

	c.Delete("/profile/sessions/" + handle).AssertRedirect("/profile")
	other.Get("/profile").AssertRedirect("/login")
	c.Get("/profile").AssertInertia("Profile/Index").HasProp("sessions", 1)

	third := app.Client(t).Inertia("")
	third.PostJSON("/login", map[string]any{"identifier": "sessions@example.com", "password": testPassword})
	c.PostJSON("/profile/sessions/revoke-all", nil).AssertRedirect("/profile")
	third.Get("/profile").AssertRedirect("/login")
	c.Get("/profile").AssertOK().AssertInertia("Profile/Index")
}

func clearPasswordSet(t *testing.T, app *vukemtest.App, email string) {
	t.Helper()

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := acc.DB.ExecContext(
		t.Context(), "UPDATE users SET password_set_at = NULL WHERE email = "+acc.DB.Dialect().Placeholder(1), email,
	); err != nil {
		t.Fatal(err)
	}
}

// TestSecurity_A07_GoogleOnlyMustSetPasswordBefore2FA: an account whose
// password was never chosen cannot reach two-factor setup.
func TestSecurity_A07_GoogleOnlyMustSetPasswordBefore2FA(t *testing.T) {
	t.Parallel()

	app, c := profileUser(t, "google@example.com")
	clearPasswordSet(t, app, "google@example.com")
	linkGoogle(t, app, "google@example.com")

	c.Get("/profile").AssertInertia("Profile/Index").PropEquals("passwordSet", false).
		PropEquals("twoFactor.pending", false)
	c.PostJSON("/user/two-factor", nil).AssertStatus(403)
	c.PostJSON("/user/two-factor/disable", map[string]any{"password": "x"}).AssertStatus(403)

	c.WithHeader("Referer", "/profile").
		PutJSON("/profile/password", map[string]any{"newPassword": "Another horse 7 staple!"}).
		AssertRedirect("/profile")
	c.PostJSON("/user/two-factor", nil).AssertRedirect("/profile")
	c.Get("/profile").AssertInertia("Profile/Index").PropEquals("twoFactor.pending", true)

	// The explanation exists in every locale.
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	for _, locale := range []string{"en", "pt-BR", "es"} {
		if msg := acc.Messages.T(
			locale,
			"auth.two_factor.password_required",
		); msg == "" ||
			msg == "auth.two_factor.password_required" {
			t.Errorf("%s: missing explanation, got %q", locale, msg)
		}
	}
}
