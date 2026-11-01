package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRenderCache(t *testing.T) {
	prefixes := map[string]string{"pt-br": "pt-BR"}
	cases := []struct {
		name, path, cookie, setCookie, inertia, ctype string
		status                                        int
		want                                          string
	}{
		{name: "anonymous home", path: "/", want: publicPageCache},
		{name: "anonymous listing", path: "/games", want: publicPageCache},
		{name: "item", path: "/movies/heat", want: publicPageCache},
		{name: "prefixed entity", path: "/pt-br/people/ana", want: publicPageCache},
		{name: "session cookie", path: "/games", cookie: "vukem_session", want: privatePageCache},
		{name: "prod session cookie", path: "/games", cookie: "__Host-vukem_session", want: privatePageCache},
		{name: "response sets cookie", path: "/games", setCookie: "a=b", want: privatePageCache},
		{name: "inertia visit untouched", path: "/games", inertia: "true", want: "orig"},
		{name: "other path untouched", path: "/profile", want: "orig"},
		{name: "deep path untouched", path: "/games/x/y/z", want: "orig"},
		{name: "api untouched", path: "/api/v1/games", want: "orig"},
		{name: "404 untouched", path: "/games/x", status: http.StatusNotFound, want: "orig"},
		{name: "non html untouched", path: "/games", ctype: "application/json", want: "orig"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := RenderCache("vukem_session", prefixes)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Cache-Control", "orig")
				ct := tc.ctype
				if ct == "" {
					ct = "text/html; charset=utf-8"
				}
				w.Header().Set("Content-Type", ct)
				if tc.setCookie != "" {
					w.Header().Add("Set-Cookie", tc.setCookie)
				}
				if tc.status != 0 {
					w.WriteHeader(tc.status)
				}
				_, _ = w.Write([]byte("x"))
			}))
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.cookie != "" {
				req.AddCookie(&http.Cookie{Name: tc.cookie, Value: "v"})
			}
			if tc.inertia != "" {
				req.Header.Set("X-Inertia", tc.inertia)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if got := rec.Header().Get("Cache-Control"); got != tc.want {
				t.Fatalf("Cache-Control = %q, want %q", got, tc.want)
			}
			if tc.want == publicPageCache && !strings.Contains(rec.Header().Get("Vary"), "Cookie") {
				t.Fatalf("Vary = %q", rec.Header().Get("Vary"))
			}
		})
	}
}

func TestRenderCachePostUntouched(t *testing.T) {
	h := RenderCache("s", nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("x"))
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/games", nil))
	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("POST got %q", got)
	}
}

func TestRenderCacheReplay(t *testing.T) {
	now := time.Unix(1000, 0)
	var renders atomic.Int32
	h := newRenderCache(
		"vukem_session",
		nil,
		func() time.Time { return now },
	)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			n := renders.Add(1)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Content-Security-Policy", fmt.Sprintf("script-src 'nonce-n%d'", n))
			w.Header().Set("X-Request-Id", fmt.Sprintf("req-%d", n))
			if r.URL.Path == "/games/cookie" {
				http.SetCookie(w, &http.Cookie{Name: "a", Value: "b"})
			}
			_, _ = fmt.Fprintf(w, `body-%d nonce="n%d"`, n, n)
		}),
	)
	get := func(path string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		return rec
	}

	first := get("/games")
	second := get("/games")
	if renders.Load() != 1 || !strings.HasPrefix(second.Body.String(), "body-1") ||
		second.Header().Get("X-Render-Cache") != "HIT" {
		t.Fatalf("no hit: renders=%d body=%q", renders.Load(), second.Body.String())
	}

	// The cache stands in for the handler only; CSP comes from the outer middleware, so here none is set on a hit.
	_ = first

	if second.Header().Get("X-Request-Id") != "" || second.Header().Get("Cache-Control") != publicPageCache {
		t.Fatalf("headers: %v", second.Header())
	}

	get("/games?page=2")
	get("/games", &http.Cookie{Name: "theme", Value: "dark"})
	if renders.Load() != 3 {
		t.Fatalf("query and theme must key separately, renders=%d", renders.Load())
	}

	get("/games", &http.Cookie{Name: "vukem_session", Value: "x"})
	get("/games", &http.Cookie{Name: "vukem_session", Value: "x"})
	if renders.Load() != 5 {
		t.Fatalf("session requests must bypass the cache, renders=%d", renders.Load())
	}

	get("/games/cookie")
	get("/games/cookie")
	if renders.Load() != 7 {
		t.Fatalf("a response that sets a cookie must not be cached, renders=%d", renders.Load())
	}

	now = now.Add(renderCacheTTL)
	if strings.HasPrefix(get("/games").Body.String(), "body-1") {
		t.Fatal("entry outlived its TTL")
	}
}

func TestRenderCacheSingleFlight(t *testing.T) {
	var renders atomic.Int32
	release := make(chan struct{})
	h := RenderCache("vukem_session", nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		renders.Add(1)
		<-release
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("x"))
	}))

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/games", nil))
		})
	}

	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()

	if renders.Load() != 1 {
		t.Fatalf("renders = %d, want 1", renders.Load())
	}
}

func TestSecurity_A05_RenderCacheKeepsNoncePerRequest(t *testing.T) {
	var renders, seq atomic.Int32
	cached := RenderCache("vukem_session", nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		renders.Add(1)
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, `<script nonce="%s"></script>`, cspNonce(w.Header()))
	}))
	// Stand-in for SecureHeaders: a fresh nonce in the header before the cache runs.
	serve := func() (nonce, body string) {
		nonce = fmt.Sprintf("N%021d", seq.Add(1))
		rec := httptest.NewRecorder()
		rec.Header().Set("Content-Security-Policy", "script-src 'nonce-"+nonce+"'")
		cached.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/games", nil))

		return nonce, rec.Body.String()
	}

	n1, b1 := serve()
	n2, b2 := serve()
	if renders.Load() != 1 {
		t.Fatalf("second request was not a hit, renders=%d", renders.Load())
	}

	if !strings.Contains(b1, n1) || !strings.Contains(b2, n2) || strings.Contains(b2, n1) {
		t.Fatalf("nonce must follow the request: %q %q", b1, b2)
	}
}

// An unprefixed page redirects by Accept-Language downstream (SDD 09 FR-09-3); a cached en render must not
// answer a pt-BR visitor, while prefixed pages share one entry whatever the header.
func TestRenderCacheKeysUnprefixedByAcceptLanguage(t *testing.T) {
	var renders atomic.Int32
	h := RenderCache("vukem_session", map[string]string{"pt-br": "pt-BR"})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			renders.Add(1)
			if r.URL.Path == "/games" && strings.HasPrefix(r.Header.Get("Accept-Language"), "pt") {
				http.Redirect(w, r, "/pt-br/games", http.StatusTemporaryRedirect)

				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte("page"))
		}),
	)
	get := func(path, al string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Accept-Language", al)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		return rec
	}

	get("/games", "en-US,en;q=0.9")
	if rec := get("/games", "pt-BR,pt;q=0.9"); rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("pt-BR visitor on /games = %d, want the 307 from the locale redirect", rec.Code)
	}

	get("/pt-br/games", "en-US")
	before := renders.Load()
	if rec := get("/pt-br/games", "es"); rec.Header().Get("X-Render-Cache") != "HIT" || renders.Load() != before {
		t.Fatalf("prefixed page not shared across Accept-Language: renders %d -> %d", before, renders.Load())
	}

	if rec := get(
		"/games",
		strings.Repeat("a", renderCacheMaxAcceptLanguage+1),
	); rec.Header().
		Get("X-Render-Cache") ==
		"HIT" {
		t.Fatal("oversized Accept-Language served from the cache")
	}
}
