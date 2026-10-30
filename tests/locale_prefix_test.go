package tests

import (
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestLocalePrefix_Routing(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	app.Client(t).Get("/pt-br/").AssertOK().AssertInertia("Home/Index")
	app.Client(t).Get("/es").AssertOK().AssertInertia("Home/Index")
	app.Client(t).Get("/").AssertOK().AssertInertia("Home/Index")
	app.Client(t).Get("/xx/").AssertStatus(404)

	// No cookie: never redirected. Cookie naming a prefixed locale: 307.
	app.Client(t).WithCookie("locale", "pt-BR").Get("/").AssertStatus(307).AssertRedirect("/pt-br")
	app.Client(t).WithCookie("locale", "en").Get("/").AssertOK()
}

func TestLocalePrefix_Theme(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	app.Client(t).WithCookie("theme", "dark").Get("/").AssertOK().AssertSee(`data-theme="dark"`)
	app.Client(t).WithCookie("theme", "evil").Get("/").AssertOK().AssertDontSee("data-theme")
}

// Redirects from account flows keep the locale prefix of the request (SDD 09 §7.1).
func TestLocalePrefix_AccountRedirects(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL": pgtest.Database(t, ""), "OPENAPI_PREFIX": "/api/v1",
	}})

	owner := app.Client(t).Inertia("")
	enableTwoFactor(t, app, owner, "pt2fa@example.com")
	owner.PostJSON("/pt-br/logout", nil).AssertRedirect("/pt-br")

	app.Client(t).Inertia("").
		PostJSON("/pt-br/login", map[string]any{"identifier": "pt2fa@example.com", "password": testPassword}).
		AssertRedirect("/pt-br/two-factor-challenge")

	c := app.Client(t).Inertia("")
	register(t, app, c, "ptflows@example.com")

	c.PutJSON("/es/profile/password", map[string]any{
		"currentPassword": testPassword, "newPassword": "Another horse 7 staple!",
	}).AssertRedirect("/es/profile")
	c.PostJSON("/pt-br/profile/api/keys", map[string]any{"name": "k"}).AssertRedirect("/pt-br/profile/api")
	c.PostJSON("/pt-br/profile/deletion", map[string]any{"confirm": true}).AssertRedirect("/pt-br/login")
}
