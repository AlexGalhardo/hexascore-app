package tests

import (
	"net/http"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestHome(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	// First visit: HTML with the page object; Inertia visits: JSON.
	app.Client(t).Get("/").AssertOK().AssertInertia("Home/Index").HasProp("appName").HasProp("coversBase")
	app.Client(t).Inertia("").Get("/").AssertOK().AssertInertia("Home/Index")
	app.Client(t).Get("/up").AssertOK()
}

func TestErrorPage(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	// A miss that keeps its 404 (outside the HTML pages, item 14) still renders the Error component.
	app.Client(t).Inertia("").Get("/build/no-such-asset").AssertStatus(404).
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

// TestMissingPage_RedirectsHome: an unknown HTML page answers 301 to the home page of its locale (owner decision
// 2026-11-01, item 14); the APIs keep problem+json and asset misses keep a real 404.
func TestMissingPage_RedirectsHome(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	c := app.Client(t)

	for _, prefix := range []string{"", "/pt-br", "/es"} {
		c.Get(prefix+"/no-such-page").AssertStatus(301).AssertHeader("Location", prefix+"/")
		c.Get(prefix+"/docs/no-such-doc").AssertStatus(301).AssertHeader("Location", prefix+"/")
		app.Client(t).Inertia("").Get(prefix+"/no/such/page").AssertStatus(301).AssertHeader("Location", prefix+"/")
	}

	c.Do(http.MethodHead, "/no-such-page", nil, "").AssertStatus(301)

	for _, p := range []string{
		"/api/no-such", "/app/v1/no-such", "/build/assets/missing.js", "/covers/missing.webp",
		"/og/item/missing.png", "/sitemaps/missing.xml", "/missing.png", "/pt-br/missing.css",
	} {
		c.Get(p).AssertStatus(404)
	}

	app.Client(t).AcceptJSON().Get("/api/no-such").AssertStatus(404).
		AssertHeader("Content-Type", "application/problem+json")
	c.PostForm("/no-such-page", nil).AssertStatus(405)
}

func TestHelpScorePage(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	for _, prefix := range []string{"", "/pt-br", "/es"} {
		app.Client(t).Get(prefix + "/how-score-works").AssertOK().AssertInertia("Help/Score").
			HasProp("categoryKeys")
	}
}
