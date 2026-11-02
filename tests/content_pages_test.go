package tests

import (
	"net/http"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestDocsPages(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	app.Client(t).Get("/docs").AssertOK().AssertInertia("Docs/Page").PropEquals("slug", "index")

	for _, prefix := range []string{"", "/pt-br", "/es"} {
		app.Client(t).Get(prefix+"/docs/errors").AssertOK().AssertInertia("Docs/Page").
			PropEquals("slug", "errors").
			HasProp("body").
			HasProp("toc")
	}

	app.Client(t).Get("/docs/no-such-doc").AssertStatus(301).AssertHeader("Location", "/")
	app.Client(t).Get("/docs/index").AssertStatus(301)
}

// TestAboutPage: owner round 2026-11-01 item 25, served in every locale with only its own namespace.
func TestAboutPage(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	for _, prefix := range []string{"", "/pt-br", "/es"} {
		res := app.Client(t).Get(prefix + "/about").AssertOK()
		res.AssertInertia("About/Index")

		body := res.Body()
		for _, key := range []string{`"about.title"`, `"about.openSourceLink"`, `"footer.about"`} {
			if !strings.Contains(body, key) {
				t.Errorf("%s/about: messages lack %s", prefix, key)
			}
		}
	}

	if strings.Contains(app.Client(t).Get("/").AssertOK().Body(), `"about.title"`) {
		t.Error("the about namespace ships on the home page")
	}
}

func TestLegalPages(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	for path, slug := range map[string]string{"/terms": "terms", "/privacy": "policy"} {
		for _, prefix := range []string{"", "/pt-br", "/es"} {
			app.Client(t).Get(prefix+path).AssertOK().AssertInertia("Legal/Page").
				PropEquals("slug", slug).
				HasProp("updatedAt").
				HasProp("messages.updatedLabel")
		}
	}
}

// TestContentPageRedirects: the old content paths answer one 301 straight to the new page in every locale
// (owner round 2026-11-01 item 10); /legal/privacy is the URL SDD 10 §7.8 and the Play listing used.
func TestContentPageRedirects(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)

	for old, target := range map[string]string{
		"/policy":        "/privacy",
		"/legal/privacy": "/privacy",
		"/help/score":    "/how-score-works",
	} {
		for _, prefix := range []string{"", "/pt-br", "/es"} {
			app.Client(t).Get(prefix+old).
				AssertStatus(http.StatusMovedPermanently).
				AssertHeader("Location", prefix+target)
		}
	}
}

func TestRobots(t *testing.T) {
	t.Parallel()

	// Non-production environments must not be indexable.
	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"APP_URL": "https://example.test/"}},
	)
	app.Client(t).Get("/robots.txt").AssertOK().
		AssertSee("Disallow: /\n").
		AssertSee("Sitemap: https://example.test/sitemap.xml")
}

func TestRobots_Production(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"APP_ENV":               "production",
		"APP_URL":               "https://example.test/",
		"RATE_LIMIT_STORE":      "cache",
		"CACHE_STORE":           "database",
		"STRIPE_KEY":            stripeKey,
		"STRIPE_WEBHOOK_SECRET": webhookSecret,
		"STRIPE_PRICE_PRO":      "price_pro",
		"STRIPE_PRICE_BUSINESS": "price_business",
	}})
	// Production redirects any host but APP_URL's, and vukemtest requests default to example.com.
	res := app.Client(t).Get("https://example.test/robots.txt").AssertOK().
		AssertSee("Allow: /" + nl).
		AssertSee("Disallow: /admin" + nl).
		AssertSee("Disallow: /profile" + nl).
		AssertSee("Disallow: /me" + nl).
		AssertSee("Disallow: /*?cursor=" + nl).
		// The honeypot is off limits for every crawler that reads robots.txt (Q-A18).
		AssertSee("Disallow: /data/export" + nl).
		AssertSee("Sitemap: https://example.test/sitemap.xml")

	// Search engines and AI crawlers are named in the allowed group (Q-A18 a).
	for _, bot := range []string{
		"Googlebot", "Bingbot", "DuckDuckBot", "Applebot", "YandexBot",
		"GPTBot", "ClaudeBot", "CCBot", "PerplexityBot", "Google-Extended", "Applebot-Extended",
	} {
		res.AssertSee("User-agent: " + bot + nl)
	}

	res.AssertSee("User-agent: *" + nl)
}

const nl = "\n"

// Staging runs APP_ENV=production (SDD 11 §6.4), so SEO_NOINDEX is what keeps it out of search results.
func TestRobots_ProductionNoindex(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"APP_ENV":               "production",
		"APP_URL":               "https://example.test/",
		"RATE_LIMIT_STORE":      "cache",
		"CACHE_STORE":           "database",
		"STRIPE_KEY":            stripeKey,
		"STRIPE_WEBHOOK_SECRET": webhookSecret,
		"STRIPE_PRICE_PRO":      "price_pro",
		"STRIPE_PRICE_BUSINESS": "price_business",
		"SEO_NOINDEX":           "true",
	}
	c := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env}).Client(t)

	res := c.Get("https://example.test/robots.txt").AssertOK().AssertSee("Disallow: /" + nl)
	res.AssertDontSee("Allow: /" + nl).AssertDontSee("Googlebot")

	for _, path := range []string{"/", "/robots.txt", "/up"} {
		c.Get("https://example.test"+path).AssertHeader("X-Robots-Tag", "noindex, nofollow")
	}

	delete(env, "SEO_NOINDEX")
	// Production leaves SEO_NOINDEX unset: the normal robots.txt with the ADR 0022 honeypot, and no X-Robots-Tag.
	prod := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env}).Client(t)
	prod.Get("https://example.test/robots.txt").AssertOK().
		AssertSee("Allow: /" + nl).
		AssertSee("Disallow: /*?cursor=" + nl).
		AssertSee("Disallow: /data/export" + nl).
		AssertSee("Googlebot").
		AssertDontSee("Disallow: /" + nl)
	for _, path := range []string{"/", "/robots.txt", "/up"} {
		if got := prod.Get("https://example.test" + path).Header("X-Robots-Tag"); got != "" {
			t.Errorf("production %s X-Robots-Tag = %q, want none", path, got)
		}
	}
}

// Outside production robots.txt disallows everything, but only SEO_NOINDEX=true adds the header to every response.
func TestRobots_NonProductionHeader(t *testing.T) {
	t.Parallel()

	c := vukemtest.New(t, bootstrap.New).Client(t)
	c.Get("/robots.txt").AssertSee("Disallow: /" + nl)
	if got := c.Get("/robots.txt").Header("X-Robots-Tag"); got != "" {
		t.Errorf("non-production X-Robots-Tag = %q, want none", got)
	}

	forced := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{"SEO_NOINDEX": "true"}}).Client(t)
	forced.Get("/robots.txt").AssertHeader("X-Robots-Tag", "noindex, nofollow")
}

func TestCanonicalHost_Production(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"APP_ENV":               "production",
		"APP_URL":               "https://example.test/",
		"RATE_LIMIT_STORE":      "cache",
		"CACHE_STORE":           "database",
		"STRIPE_KEY":            stripeKey,
		"STRIPE_WEBHOOK_SECRET": webhookSecret,
		"STRIPE_PRICE_PRO":      "price_pro",
		"STRIPE_PRICE_BUSINESS": "price_business",
	}})
	c := app.Client(t)

	c.Get("http://www.example.test/robots.txt?a=1").AssertStatus(301).
		AssertHeader("Location", "https://example.test/robots.txt?a=1")
	c.Get("http://app-production-785f.up.railway.app/robots.txt").AssertStatus(301)
	c.Get("http://www.example.test/up").AssertOK()
	// SQLite has no catalogue routes: 404 proves the API host was not redirected.
	c.Get("http://api.hexascore.app/v1/games").AssertStatus(404)
}

// /llms.txt lists the catalogue, the API docs and the legal pages; /llms-full.txt carries the docs Markdown and no
// item page.
func TestLLMsTxt(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"APP_URL": "https://example.test/"}},
	)
	app.Client(t).Get("/llms.txt").AssertOK().
		AssertHeader("Content-Type", "text/plain; charset=utf-8").
		AssertSee("# HexaScore\n").
		AssertSee("[Games](https://example.test/games)").
		AssertSee("[Books](https://example.test/books)").
		AssertSee("[Authentication](https://example.test/docs/authentication)").
		AssertSee("[Pricing](https://example.test/pricing)").
		AssertSee("[Terms of use](https://example.test/terms)").
		AssertSee("[Privacy policy](https://example.test/privacy)").
		AssertSee("/games/platforms/playstation-5")

	full := app.Client(t).Get("/llms-full.txt").AssertOK().
		AssertSee("Source: https://example.test/docs/authentication").
		AssertSee("Every request is authenticated with a secret API key")
	if strings.Contains(full.Body(), "Source: https://example.test/games") {
		t.Error("llms-full.txt streams an item listing")
	}
}
