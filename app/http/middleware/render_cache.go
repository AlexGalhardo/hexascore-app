package middleware

import (
	"bytes"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	publicPageCache  = "public, max-age=60, stale-while-revalidate=300"
	privatePageCache = "private, no-store"

	renderCacheTTL     = 60 * time.Second
	renderCacheEntries = 1000
	renderCacheMaxBody = 512 << 10
	// Cookie values that vary the page are short enums; a longer one is not ours and skips the cache.
	renderCacheMaxCookie = 16
	// Browsers send a short Accept-Language; a longer one skips the cache instead of filling it.
	renderCacheMaxAcceptLanguage = 64
)

var nonceRe = regexp.MustCompile(`'nonce-([^']+)'`)

func cspNonce(h http.Header) []byte {
	if m := nonceRe.FindStringSubmatch(h.Get("Content-Security-Policy")); m != nil {
		return []byte(m[1])
	}

	return nil
}

// volatileHeaders are per-request and never replayed from a cached response.
var volatileHeaders = []string{"X-Request-Id", "Date", "Set-Cookie", "Content-Security-Policy"}

var cacheableSections = map[string]bool{
	"games": true, "movies": true, "series": true, "books": true,
	"people": true, "companies": true, "franchises": true, "platforms": true,
	"genres": true, "game-modes": true, "countries": true, "languages": true,
}

// RenderCache sets Cache-Control on the public catalogue HTML pages (SDD 09
// §7.2): home, listings, item and entity pages. Anonymous 200 responses that
// set no cookie are public for 60 s with a 300 s stale window, so a CDN or
// browser serves the page together with the CSP header whose nonce it embeds.
// A request carrying the session cookie (sessionCookie, with or without the
// "__Host-" prefix), or any response that sets a cookie,
// is private and uncacheable. Inertia JSON visits and every other path are
// left alone. localePrefixes are the first path segments (no slashes) the app
// serves a locale under, since the prefix may still be on the URL here.
//
// The same anonymous 200 pages are also kept in process for 60 s (SDD 09 §7.2;
// vukem has no SSR output cache hook, VF-025) and replayed with their stored
// headers, except the Content-Security-Policy header: that stays the one
// SecureHeaders built for this request, and the nonce the page was rendered
// with is swapped for it, so the nonce is still per request (A05).
func RenderCache(sessionCookie string, localePrefixes map[string]string) func(http.Handler) http.Handler {
	return newRenderCache(sessionCookie, localePrefixes, time.Now)
}

func newRenderCache(
	sessionCookie string,
	localePrefixes map[string]string,
	now func() time.Time,
) func(http.Handler) http.Handler {
	store := &pageStore{now: now, entries: map[string]*pageEntry{}, flights: map[string]chan struct{}{}}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cacheablePage(r, localePrefixes) {
				next.ServeHTTP(w, r)

				return
			}

			anonymous := !hasSession(r, sessionCookie)
			key, ok := pageKey(r, localePrefixes)
			if !anonymous || !ok || r.Method != http.MethodGet {
				next.ServeHTTP(&cacheWriter{ResponseWriter: w, anonymous: anonymous}, r)

				return
			}

			for {
				if e := store.get(key); e != nil {
					e.replay(w)

					return
				}

				wait, leader := store.join(key)
				if leader {
					break
				}

				<-wait
			}

			defer store.leave(key)

			cw := &cacheWriter{ResponseWriter: w, anonymous: true}
			rec := &recorder{cacheWriter: cw}
			next.ServeHTTP(rec, r)
			store.put(key, rec)
		})
	}
}

// pageKey is the path, the raw query and the cookies that change the markup
// (theme, consent) or the redirect decision (locale). On an unprefixed path
// Accept-Language also picks the redirect (SDD 09 FR-09-3), so it is part of
// the key there; this middleware runs before that redirect.
func pageKey(r *http.Request, localePrefixes map[string]string) (string, bool) {
	var b strings.Builder
	b.WriteString(r.URL.Path)
	b.WriteByte('?')
	b.WriteString(r.URL.RawQuery)

	first, _, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if _, prefixed := localePrefixes[first]; !prefixed {
		al := r.Header.Get("Accept-Language")
		if len(al) > renderCacheMaxAcceptLanguage {
			return "", false
		}

		b.WriteString("|al=")
		b.WriteString(al)
	}

	for _, name := range []string{"theme", "consent", "locale"} {
		b.WriteByte('|')

		if c, err := r.Cookie(name); err == nil {
			if len(c.Value) > renderCacheMaxCookie {
				return "", false
			}

			b.WriteString(c.Value)
		}
	}

	return b.String(), true
}

type pageEntry struct {
	status  int
	header  http.Header
	body    []byte
	nonce   []byte
	expires time.Time
}

func (e *pageEntry) replay(w http.ResponseWriter) {
	h := w.Header()
	body := e.body
	if now := cspNonce(h); len(e.nonce) > 0 && len(now) > 0 {
		body = bytes.ReplaceAll(body, e.nonce, now)
	}

	for k, v := range e.header {
		h[k] = append([]string(nil), v...)
	}

	h.Set("X-Render-Cache", "HIT")
	w.WriteHeader(e.status)
	_, _ = w.Write(body)
}

type pageStore struct {
	now     func() time.Time
	mu      sync.Mutex
	entries map[string]*pageEntry
	flights map[string]chan struct{}
}

func (s *pageStore) get(key string) *pageEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	e := s.entries[key]
	if e != nil && !s.now().Before(e.expires) {
		delete(s.entries, key)

		return nil
	}

	return e
}

// join returns leader=true for the first caller of a key; the others get a
// channel that closes when the leader is done (one render per key and expiry).
func (s *pageStore) join(key string) (<-chan struct{}, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.flights[key]; ok {
		return ch, false
	}

	s.flights[key] = make(chan struct{})

	return nil, true
}

func (s *pageStore) leave(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	close(s.flights[key])
	delete(s.flights, key)
}

func (s *pageStore) put(key string, rec *recorder) {
	if !rec.cacheable() {
		return
	}

	h := rec.Header().Clone()
	nonce := cspNonce(h)
	for _, k := range volatileHeaders {
		h.Del(k)
	}

	h.Del("Content-Length")

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.entries) >= renderCacheEntries {
		now := s.now()
		for k, e := range s.entries {
			if !now.Before(e.expires) {
				delete(s.entries, k)
			}
		}

		// Full of live entries (a flood of distinct query strings): do not grow.
		if len(s.entries) >= renderCacheEntries {
			return
		}
	}

	s.entries[key] = &pageEntry{
		status:  rec.status,
		header:  h,
		body:    rec.body.Bytes(),
		nonce:   nonce,
		expires: s.now().Add(renderCacheTTL),
	}
}

// recorder passes the response through and keeps a copy of a public 200.
type recorder struct {
	*cacheWriter

	status int
	body   bytes.Buffer
	over   bool
}

func (r *recorder) WriteHeader(code int) {
	r.status = code
	r.cacheWriter.WriteHeader(code)
}

func (r *recorder) Write(p []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}

	if r.body.Len()+len(p) > renderCacheMaxBody {
		r.over = true
	} else {
		r.body.Write(p)
	}

	return r.cacheWriter.Write(p)
}

func (r *recorder) cacheable() bool {
	return r.status == http.StatusOK && !r.over &&
		r.Header().Get("Cache-Control") == publicPageCache && len(r.Header().Values("Set-Cookie")) == 0
}

// hasSession matches by suffix: vukem prefixes the cookie with "__Host-" when
// it is Secure on path "/" (production).
func hasSession(r *http.Request, name string) bool {
	for _, c := range r.Cookies() {
		if strings.HasSuffix(c.Name, name) {
			return true
		}
	}

	return false
}

func cacheablePage(r *http.Request, localePrefixes map[string]string) bool {
	if (r.Method != http.MethodGet && r.Method != http.MethodHead) || r.Header.Get("X-Inertia") != "" {
		return false
	}

	segs := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if _, ok := localePrefixes[segs[0]]; ok {
		segs = segs[1:]
	}

	switch len(segs) {
	case 0:
		return true
	case 1:
		return segs[0] == "" || cacheableSections[segs[0]]
	case 2:
		return cacheableSections[segs[0]] && segs[1] != ""
	default:
		return false
	}
}

type cacheWriter struct {
	http.ResponseWriter

	anonymous bool
	wrote     bool
}

func (c *cacheWriter) WriteHeader(code int) {
	c.decide(code)
	c.ResponseWriter.WriteHeader(code)
}

func (c *cacheWriter) Write(p []byte) (int, error) {
	c.decide(http.StatusOK)

	return c.ResponseWriter.Write(p)
}

func (c *cacheWriter) Unwrap() http.ResponseWriter { return c.ResponseWriter }

func (c *cacheWriter) decide(code int) {
	if c.wrote {
		return
	}

	c.wrote = true
	h := c.Header()
	if code == http.StatusOK && !wantsHTML(h.Get("Content-Type")) {
		return
	}

	switch {
	case code != http.StatusOK:
		// Redirects, 404 and errors keep the headers their handler chose.
	case c.anonymous && len(h.Values("Set-Cookie")) == 0:
		h.Set("Cache-Control", publicPageCache)
		h.Add("Vary", "Cookie")
		h.Add("Vary", "Accept-Language")
	default:
		h.Set("Cache-Control", privatePageCache)
	}
}
