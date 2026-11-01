package tests

import (
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestHome(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	// First visit: HTML with the page object; Inertia visits: JSON.
	app.Client(t).Get("/").AssertOK().AssertInertia("Home/Index").HasProp("appName")
	app.Client(t).Inertia("").Get("/").AssertOK().AssertInertia("Home/Index")
	app.Client(t).Get("/up").AssertOK()
}

func TestErrorPage(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	// Full page loads render it too (catch-all route).
	app.Client(t).Inertia("").Get("/no-such-page").AssertStatus(404).
		AssertInertia("Error").
		PropEquals("status", 404).
		PropEquals("message", "Not Found")
}

func TestShell_SEO(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"APP_URL": "https://example.test/"}},
	)

	app.Client(t).Get("/pt-br/").AssertOK().AssertInertia("Home/Index").
		PropEquals("seo.canonical", "https://example.test/pt-br/").
		PropEquals("seo.alternates.en", "https://example.test/").
		PropEquals("seo.alternates.es", "https://example.test/es/").
		PropEquals("seo.alternates.x-default", "https://example.test/").
		PropEquals("seo.robots", "index,follow")
	app.Client(t).Get("/login").AssertOK().AssertInertia("Auth/Login").PropEquals("seo.robots", "noindex,nofollow")
}

func TestErrorPage_FullLoad(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	app.Client(t).Get("/no-such-page").AssertStatus(404).AssertInertia("Error").
		PropEquals("status", 404).
		HasProp("seo")
	app.Client(t).Get("/api/no-such").AssertStatus(404)
	app.Client(t).Get("/pt-br/no-such-page").AssertStatus(404).AssertInertia("Error")
}

func TestHelpScorePage(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	app.Client(t).Get("/help/score").AssertOK().AssertInertia("Help/Score").HasProp("categoryKeys")
}
