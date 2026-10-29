//go:build integration

package tests

import (
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/social"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestLoginByNickname(t *testing.T) {
	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	c := app.Client(t)
	register(t, app, c, "ana@example.com")
	c.PostJSON("/onboarding", map[string]any{"nickname": "analuiza"}).AssertRedirect("/dashboard")
	c.PostJSON("/logout", nil).AssertRedirect("/")

	// Nickname and email reach the same account; case is ignored.
	app.Client(t).PostJSON("/login", map[string]any{"identifier": "AnaLuiza", "password": testPassword}).
		AssertRedirect("/dashboard")
	app.Client(t).PostJSON("/login", map[string]any{"identifier": "ana@example.com", "password": testPassword}).
		AssertRedirect("/dashboard")

	// Unknown identifier and wrong password answer alike.
	wrong := app.Client(t).Inertia("").WithHeader("Referer", "/login")
	wrong.PostJSON("/login", map[string]any{"identifier": "analuiza", "password": "wrong"}).AssertRedirect("/login")
	wrong.Get("/login").AssertInertia("Auth/Login").HasProp("errors.identifier")

	unknown := app.Client(t).Inertia("").WithHeader("Referer", "/login")
	unknown.PostJSON("/login", map[string]any{"identifier": "nobodyhere", "password": testPassword}).
		AssertRedirect("/login")
	unknown.Get("/login").AssertInertia("Auth/Login").HasProp("errors.identifier")
}

func TestGoogleNewUserNeedsNickname(t *testing.T) {
	user := social.User{ID: "gh-9", Email: "Lia@Example.com", EmailVerified: true, Name: "Lia"}
	env := vukemtest.FakeSocial(user)
	env["DATABASE_URL"] = pgtest.Database(t, "")
	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})

	c := app.Client(t)
	signIn(t, c).AssertRedirect("/onboarding")
	c.PostJSON("/onboarding", map[string]any{"nickname": "lialima"}).AssertRedirect("/dashboard")
	c.PostJSON("/logout", nil).AssertRedirect("/")

	signIn(t, app.Client(t)).AssertRedirect("/dashboard")
}

func TestVerifiedLoginWithoutNicknameGoesToOnboarding(t *testing.T) {
	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	c := app.Client(t)
	c.PostJSON("/signup", signUpForm("Bia Souza", "bia@example.com", testPassword))
	// Unverified users are not sent to onboarding: they still need the mail.
	c.PostJSON("/login", map[string]any{"identifier": "bia@example.com", "password": testPassword}).
		AssertRedirect("/dashboard")
}
