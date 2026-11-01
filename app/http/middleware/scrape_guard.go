package middleware

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/cache"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/botverify"
)

// ScrapeTrapPath is the honeypot: a hidden link in every page points here and robots.txt disallows it,
// so only a client that ignores both ever asks for it.
const ScrapeTrapPath = "/data/export"

// ScrapeBanTTL is how long a trapped IP is denied pages.
const ScrapeBanTTL = 24 * time.Hour

// scrapeExempt are the path prefixes the guard never sees: the paid and app APIs (own keys and limits),
// assets, health, the API docs a denied client is pointed to, the llms.txt files written for language models, and
// fetches made by the pages themselves.
var scrapeExempt = []string{
	"/api", "/app", "/build", "/covers", "/og", "/up", "/docs", "/.well-known", "/sitemaps",
	"/challenge", "/search/suggest", "/llms.txt", "/llms-full.txt",
}

// toolAgents are User-Agent substrings of HTTP libraries and headless browsers no reader uses.
var toolAgents = []string{
	"python-requests", "python-urllib", "aiohttp", "httpx", "curl/", "wget/", "scrapy", "go-http-client",
	"axios/", "node-fetch", "headlesschrome", "phantomjs", "libwww-perl",
}

// ScrapeGuard protects the public HTML pages and Inertia visits from bulk harvesting (ADR 0022): verified
// search crawlers pass, a crawler UA that fails verification and obvious tools get 403, a client that
// followed the honeypot is denied for ScrapeBanTTL, and everyone else is rate limited per IP (Anon, or
// User when SignedIn says the session holds a signed-in user). It must wrap RenderCache, so cached pages
// are limited too.
type ScrapeGuard struct {
	Bots *botverify.Verifier
	Bans *cache.Cache
	Anon []ratelimit.Limiter
	User []ratelimit.Limiter
	// SignedIn reports whether the request's session belongs to a signed-in user. A bare or forged session
	// cookie must not count, or a scraper keeps one and gets the User budget. Nil treats everyone as Anon.
	SignedIn       func(*http.Request) bool
	LocalePrefixes map[string]string
	// Soft is a per-IP page budget below Anon; past it an anonymous client goes through StepUp first.
	Soft ratelimit.Limiter
	// StepUp asks for a Turnstile solve (challenge.StepUp.Require): true lets the request go on, false means
	// the interstitial was written. Nil, or Turnstile off, never challenges.
	StepUp func(http.ResponseWriter, *http.Request) bool
}

// VerifiedCrawler reports whether the request comes from a verified search crawler (the StepUp skip); those are
// never challenged. botverify caches the verdict.
func (g *ScrapeGuard) VerifiedCrawler(c *web.Ctx) bool {
	crawler, claims := botverify.Claimed(c.Request().UserAgent())

	return claims && g.Bots.Verify(c.Context(), crawler, c.IP()) == botverify.Verified
}

// PageLimitScale parses PAGE_LIMIT_SCALE, the factor the QA stack applies to the page limits so one
// host can run the e2e and Lighthouse suites. Empty keeps 1; anything but a positive integer refuses to boot.
func PageLimitScale(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 1, nil
	}

	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("PAGE_LIMIT_SCALE must be a positive integer, got %q", raw)
	}

	return n, nil
}

// Middleware returns the guard around next.
func (g *ScrapeGuard) Middleware(next http.Handler) http.Handler {
	chain := func(ls []ratelimit.Limiter) http.Handler {
		h := next
		for _, l := range ls {
			h = ratelimit.Middleware(l, ratelimit.ByIP)(h)
		}

		return h
	}
	anon, user := chain(g.Anon), chain(g.User)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := guardedPage(r, g.LocalePrefixes)
		if !ok {
			next.ServeHTTP(w, r)

			return
		}

		ctx := r.Context()
		ua := r.UserAgent()

		verdict := botverify.Unverified
		crawler, claims := botverify.Claimed(ua)

		if claims {
			// workaround(vukem): VF-065 (no raw client IP helper outside a *web.Ctx).
			ip := web.NewCtx(nil, r, web.ServicesFrom(ctx)).IP()
			verdict = g.Bots.Verify(ctx, crawler, ip)
		}

		if verdict == botverify.Verified {
			if p == ScrapeTrapPath {
				http.NotFound(w, r)

				return
			}

			next.ServeHTTP(w, r)

			return
		}

		banKey := "scrape:ban:" + ratelimit.ByIP(r)

		if p == ScrapeTrapPath {
			if err := g.Bans.Put(ctx, banKey, true, ScrapeBanTTL); err != nil {
				web.Fail(w, r, errs.Wrap(http.StatusServiceUnavailable, err))

				return
			}

			web.SecurityEvent(r, "scrape.honeypot", "alert", "scrape", "reason", "honeypot")
			denyScraper(w, r)

			return
		}

		banned, _, err := cache.Get[bool](ctx, g.Bans, banKey)
		if err != nil {
			web.Fail(w, r, errs.Wrap(http.StatusServiceUnavailable, err))

			return
		}

		switch {
		case banned:
			denyScraper(w, r)
		case verdict == botverify.Spoofed:
			web.SecurityEvent(r, "scrape.spoofed_crawler", "alert", "scrape", "reason", "spoofed_crawler",
				"crawler", crawler.Name)
			denyScraper(w, r)
		case toolAgent(ua):
			denyScraper(w, r)
		case g.SignedIn != nil && g.SignedIn(r):
			user.ServeHTTP(w, r)
		default:
			if g.softExceeded(r) && g.StepUp != nil && !g.StepUp(w, r) {
				return
			}

			anon.ServeHTTP(w, r)
		}
	})
}

// guardedPage reports whether r is a page view the guard covers and returns its path without the
// locale prefix: GET or HEAD, outside scrapeExempt, and not a file (an extension means an asset).
func guardedPage(r *http.Request, localePrefixes map[string]string) (string, bool) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return "", false
	}

	p := r.URL.Path
	if first, rest, _ := strings.Cut(strings.TrimPrefix(p, "/"), "/"); localePrefixes[first] != "" {
		p = "/" + rest
	}

	if path.Ext(p) != "" {
		return "", false
	}

	for _, prefix := range scrapeExempt {
		if p == prefix || strings.HasPrefix(p, prefix+"/") {
			return "", false
		}
	}

	return p, true
}

// softExceeded counts the request against Soft. A store error does not challenge: the hard limits that follow
// fail closed on the same store.
func (g *ScrapeGuard) softExceeded(r *http.Request) bool {
	if g.Soft == nil {
		return false
	}

	res, err := g.Soft.Allow(r.Context(), ratelimit.ByIP(r))

	return err == nil && !res.Allowed
}

func toolAgent(ua string) bool {
	ua = strings.ToLower(strings.TrimSpace(ua))
	if ua == "" {
		return true
	}

	for _, t := range toolAgents {
		if strings.Contains(ua, t) {
			return true
		}
	}

	return false
}

// denyScraper answers 403 through the app's error handler (problem+json for JSON clients) and points
// bulk data users to the API.
func denyScraper(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	web.Fail(w, r, errs.New(http.StatusForbidden,
		"Automated access to HexaScore pages is not allowed. For catalogue data, use the HexaScore API: "+
			"https://hexascore.app/docs"))
}
