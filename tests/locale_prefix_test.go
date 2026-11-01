package tests

import (
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestLocalePrefix_Routing(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	app.Client(t).Get("/pt-br/").AssertOK().AssertInertia("Home/Index")
	app.Client(t).Get("/es").AssertOK().AssertInertia("Home/Index")
	app.Client(t).Get("/").AssertOK().AssertInertia("Home/Index")
	app.Client(t).Get("/xx/").AssertStatus(301).AssertHeader("Location", "/")

	// No cookie: never redirected. Cookie naming a prefixed locale: 307.
	app.Client(t).WithCookie("locale", "pt-BR").Get("/").AssertStatus(307).AssertRedirect("/pt-br/")
	app.Client(t).WithCookie("locale", "en").Get("/").AssertOK()
}

func TestLocalePrefix_Theme(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	app.Client(t).WithCookie("theme", "dark").Get("/").AssertOK().AssertSee(`data-theme="dark"`)
	app.Client(t).WithCookie("theme", "evil").Get("/").AssertOK().AssertDontSee("data-theme")
}
