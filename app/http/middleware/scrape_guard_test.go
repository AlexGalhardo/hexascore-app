package middleware_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/cache"
	"github.com/AlexGalhardo/vukem/ratelimit"

	"github.com/AlexGalhardo/hexascore-app/app/botverify"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
)

const (
	browserUA   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0 Safari/537.36"
	googlebotUA = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	googleIP    = "66.249.66.1"
	spoofIP     = "203.0.113.9"
)

// dnsFake answers PTR and forward lookups from fixed maps; no live DNS in tests.
type dnsFake struct{}

func (dnsFake) LookupAddr(_ context.Context, addr string) ([]string, error) {
	switch addr {
	case googleIP:
		return []string{"crawl-66-249-66-1.googlebot.com."}, nil
	case spoofIP:
		return []string{"scraper.example.net."}, nil
	}

	return nil, &net.DNSError{Err: "no such host", Name: addr, IsNotFound: true}
}

func (dnsFake) LookupHost(_ context.Context, host string) ([]string, error) {
	if host == "crawl-66-249-66-1.googlebot.com" {
		return []string{googleIP}, nil
	}

	return nil, &net.DNSError{Err: "no such host", Name: host, IsNotFound: true}
}

func memLimiter(_ string, limit int, window time.Duration) ratelimit.Limiter {
	return ratelimit.NewMemory(limit, window)
}

func newScrapeGuard(store cache.Store) http.Handler {
	c := cache.New(store, "")
	g := &middleware.ScrapeGuard{
		Bots: &botverify.Verifier{Resolver: dnsFake{}, Cache: c, Timeout: 50 * time.Millisecond},
		Bans: c,
		Anon: []ratelimit.Limiter{limits.PageViews.From(memLimiter), limits.PageViewsHour.From(memLimiter)},
		User: []ratelimit.Limiter{
			limits.PageViewsUser.From(memLimiter),
			limits.PageViewsUserHour.From(memLimiter),
		},
		SessionCookie:  "vukem_session",
		LocalePrefixes: map[string]string{"pt-br": "pt-BR", "es": "es"},
	}

	return g.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html>page</html>"))
	}))
}

func get(h http.Handler, path, ua, ip string, mod ...func(*http.Request)) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = ip + ":40000"

	if ua != "" {
		req.Header.Set("User-Agent", ua)
	}

	for _, m := range mod {
		m(req)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	return rec
}

func inertia(r *http.Request) { r.Header.Set("X-Inertia", "true") }

// TestSecurity_A06_ScrapeToolUADenied: obvious HTTP tools and headless browsers get 403 on pages and
// Inertia visits.
func TestSecurity_A06_ScrapeToolUADenied(t *testing.T) {
	t.Parallel()

	h := newScrapeGuard(cache.NewMemory(1000))

	for _, ua := range []string{
		"", "python-requests/2.32.3", "curl/8.10.1", "Wget/1.24.5", "Scrapy/2.11.2 (+https://scrapy.org)",
		"Go-http-client/1.1", "axios/1.7.9", "node-fetch/1.0 (+https://github.com/bitinn/node-fetch)",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/142.0 Safari/537.36",
		"Mozilla/5.0 (Unknown; Linux x86_64) AppleWebKit/538.1 (KHTML, like Gecko) PhantomJS/2.1.1 Safari/538.1",
	} {
		for _, path := range []string{"/games/zelda", "/pt-br/games"} {
			if rec := get(h, path, ua, "198.51.100.7"); rec.Code != http.StatusForbidden {
				t.Errorf("UA %q on %s: status %d, want 403", ua, path, rec.Code)
			}
		}

		if rec := get(h, "/games/zelda", ua, "198.51.100.7", inertia); rec.Code != http.StatusForbidden {
			t.Errorf("UA %q Inertia visit: status %d, want 403", ua, rec.Code)
		}
	}

	if rec := get(h, "/games/zelda", browserUA, "198.51.100.7"); rec.Code != http.StatusOK {
		t.Errorf("browser: status %d, want 200", rec.Code)
	}
}

// TestSecurity_A06_ScrapeAPIRoutesUnaffected: the paid API, the mobile app API, assets, health, docs,
// robots and sitemaps never pass through the page guard.
func TestSecurity_A06_ScrapeAPIRoutesUnaffected(t *testing.T) {
	t.Parallel()

	h := newScrapeGuard(cache.NewMemory(1000))

	for _, path := range []string{
		"/api/v1/games", "/api/items/1/ratings", "/app/v1/items/1", "/build/assets/app-1a2b.js", "/up",
		"/covers/abc.webp", "/robots.txt", "/sitemap.xml", "/sitemaps/games-1.xml", "/docs", "/docs/reference",
		"/pt-br/docs", "/favicon.ico", "/.well-known/assetlinks.json", "/search/suggest", "/og/games/zelda",
	} {
		for range 100 {
			if rec := get(h, path, "curl/8.10.1", "198.51.100.8"); rec.Code != http.StatusOK {
				t.Fatalf("%s: status %d, want 200", path, rec.Code)
			}
		}
	}

	post := httptest.NewRequest(http.MethodPost, "/login", nil)
	post.RemoteAddr = "198.51.100.8:1"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, post)

	if rec.Code != http.StatusOK {
		t.Errorf("POST /login: status %d, want 200 (forms keep their own limits)", rec.Code)
	}
}

// TestSecurity_A06_ScrapeSpoofedGooglebot: a Googlebot UA whose IP fails reverse plus forward DNS is denied.
func TestSecurity_A06_ScrapeSpoofedGooglebot(t *testing.T) {
	t.Parallel()

	h := newScrapeGuard(cache.NewMemory(1000))

	for _, ip := range []string{spoofIP, "198.51.100.9"} {
		if rec := get(h, "/games/zelda", googlebotUA, ip); rec.Code != http.StatusForbidden {
			t.Errorf("spoofed Googlebot from %s: status %d, want 403", ip, rec.Code)
		}
	}
}

// TestSecurity_A06_ScrapeVerifiedGooglebot: a verified Googlebot is exempt from the page limits, gets the
// same 200 as a browser and is never banned by the trap.
func TestSecurity_A06_ScrapeVerifiedGooglebot(t *testing.T) {
	t.Parallel()

	h := newScrapeGuard(cache.NewMemory(1000))

	for i := range limits.PageViewsHour.Limit + 10 {
		rec := get(h, "/games/zelda", googlebotUA, googleIP)
		if rec.Code != http.StatusOK || rec.Body.String() != "<html>page</html>" {
			t.Fatalf("request %d: status %d body %q", i+1, rec.Code, rec.Body.String())
		}
	}

	if rec := get(h, middleware.ScrapeTrapPath, googlebotUA, googleIP); rec.Code != http.StatusNotFound {
		t.Errorf("verified bot on the trap: status %d, want 404", rec.Code)
	}

	if rec := get(h, "/games/zelda", googlebotUA, googleIP); rec.Code != http.StatusOK {
		t.Errorf("verified bot after the trap: status %d, want 200 (no ban)", rec.Code)
	}
}

// TestSecurity_A06_ScrapePageRateLimit: anonymous page views per IP get 429 with Retry-After past the
// budget; a signed-in browser has the larger budget; another IP is unaffected.
func TestSecurity_A06_ScrapePageRateLimit(t *testing.T) {
	t.Parallel()

	h := newScrapeGuard(cache.NewMemory(1000))
	ip := "198.51.100.10"

	for i := range limits.PageViews.Limit {
		if rec := get(h, "/games/zelda", browserUA, ip, inertia); rec.Code != http.StatusOK {
			t.Fatalf("request %d: status %d", i+1, rec.Code)
		}
	}

	rec := get(h, "/games/zelda", browserUA, ip)
	if rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") == "" {
		t.Fatalf("over budget: status %d, Retry-After %q", rec.Code, rec.Header().Get("Retry-After"))
	}

	if rec := get(h, "/games/zelda", browserUA, "198.51.100.11"); rec.Code != http.StatusOK {
		t.Errorf("other IP: status %d, want 200", rec.Code)
	}

	session := func(r *http.Request) { r.AddCookie(&http.Cookie{Name: "__Host-vukem_session", Value: "x"}) }
	for i := range limits.PageViews.Limit + 1 {
		if rec := get(h, "/games/zelda", browserUA, "198.51.100.12", session); rec.Code != http.StatusOK {
			t.Fatalf("signed-in request %d: status %d", i+1, rec.Code)
		}
	}
}

// TestSecurity_A06_ScrapeHoneypotBan: an unverified client that follows the hidden trap link is denied
// every page from that IP; other IPs and the API are unaffected.
func TestSecurity_A06_ScrapeHoneypotBan(t *testing.T) {
	t.Parallel()

	h := newScrapeGuard(cache.NewMemory(1000))
	ip := "198.51.100.13"

	if rec := get(h, middleware.ScrapeTrapPath, browserUA, ip); rec.Code != http.StatusForbidden {
		t.Fatalf("trap: status %d, want 403", rec.Code)
	}

	if rec := get(h, "/games/zelda", browserUA, ip); rec.Code != http.StatusForbidden {
		t.Errorf("banned IP: status %d, want 403", rec.Code)
	}

	if rec := get(h, "/api/v1/games", browserUA, ip); rec.Code != http.StatusOK {
		t.Errorf("banned IP on the API: status %d, want 200", rec.Code)
	}

	if rec := get(h, "/games/zelda", browserUA, "198.51.100.14"); rec.Code != http.StatusOK {
		t.Errorf("other IP: status %d, want 200", rec.Code)
	}
}

type downStore struct{ cache.Store }

func (downStore) Get(context.Context, string) ([]byte, bool, error) {
	return nil, false, errors.New("store down")
}

// TestSecurity_A10_ScrapeBanStoreDown503: when the deny list cannot be read, pages fail closed.
func TestSecurity_A10_ScrapeBanStoreDown503(t *testing.T) {
	t.Parallel()

	h := newScrapeGuard(downStore{})

	if rec := get(h, "/games/zelda", browserUA, "198.51.100.15"); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("store down: status %d, want 503", rec.Code)
	}
}

func TestPageLimitScale(t *testing.T) {
	t.Parallel()

	for raw, want := range map[string]int{"": 1, "1000": 1000} {
		if n, err := middleware.PageLimitScale(raw); err != nil || n != want {
			t.Errorf("%q = %d, %v; want %d", raw, n, err, want)
		}
	}

	for _, raw := range []string{"0", "-1", "x", "1.5"} {
		if _, err := middleware.PageLimitScale(raw); err == nil {
			t.Errorf("%q: want a refusal", raw)
		}
	}
}
