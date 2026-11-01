package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const gzipPageBody = "<!doctype html><html><body>" + "catalogue page text " + "</body></html>"

func gzipPageChain(setCookie, ctype string, status int) http.Handler {
	prefixes := map[string]string{"pt-br": "pt-BR"}
	page := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if ctype == "" {
			ctype = "text/html; charset=utf-8"
		}

		w.Header().Set("Content-Type", ctype)
		if setCookie != "" {
			w.Header().Add("Set-Cookie", setCookie)
		}

		if status != 0 {
			w.WriteHeader(status)
		}

		_, _ = fmt.Fprint(w, strings.Repeat(gzipPageBody, 100))
	})

	return GzipPublicPages("vukem_session", prefixes)(RenderCache("vukem_session", prefixes)(page))
}

func gunzipBody(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	zr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}

	b, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

func TestGzipPublicPages_AnonymousCataloguePageIsCompressed(t *testing.T) {
	h := gzipPageChain("", "", 0)
	for _, call := range []string{"miss", "hit"} {
		req := httptest.NewRequest(http.MethodGet, "/movies/heat", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Header().Get("Content-Encoding") != "gzip" {
			t.Fatalf("%s: Content-Encoding = %q, want gzip", call, rec.Header().Get("Content-Encoding"))
		}

		if !strings.Contains(strings.Join(rec.Header().Values("Vary"), ","), "Accept-Encoding") {
			t.Fatalf("%s: Vary = %q lacks Accept-Encoding", call, rec.Header().Get("Vary"))
		}

		if got := gunzipBody(t, rec); got != strings.Repeat(gzipPageBody, 100) {
			t.Fatalf("%s: gunzipped body changed (%d bytes)", call, len(got))
		}
	}
}

// BREACH (CWE-200): a compressed body that carries a secret next to attacker-chosen input leaks the secret.
// Only the anonymous, cookie-less, public catalogue HTML may be compressed; everything session-bound or
// CSRF-bearing stays plain.
func TestSecurity_A02_GzipNeverCompressesSessionOrCSRFResponses(t *testing.T) {
	cases := []struct {
		name, path, cookie, setCookie, ctype string
		status                               int
	}{
		{name: "session cookie", path: "/games", cookie: "vukem_session"},
		{name: "prod session cookie", path: "/games", cookie: "__Host-vukem_session"},
		{name: "response sets cookie (CSRF token)", path: "/games", setCookie: "csrf=secret; Path=/"},
		{name: "admin console", path: "/admin/moderation"},
		{name: "login form", path: "/login"},
		{name: "profile", path: "/profile"},
		{name: "404", path: "/games/missing", status: http.StatusNotFound},
		{name: "json", path: "/games", ctype: "application/json"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Header.Set("Accept-Encoding", "gzip")
			if tc.cookie != "" {
				req.AddCookie(&http.Cookie{Name: tc.cookie, Value: "v"})
			}

			rec := httptest.NewRecorder()
			gzipPageChain(tc.setCookie, tc.ctype, tc.status).ServeHTTP(rec, req)

			if enc := rec.Header().Get("Content-Encoding"); enc != "" {
				t.Fatalf("Content-Encoding = %q, want none", enc)
			}

			if rec.Body.String() != strings.Repeat(gzipPageBody, 100) {
				t.Fatalf("body altered (%d bytes)", rec.Body.Len())
			}
		})
	}
}

func TestGzipPublicPages_NoncePerRequestSurvivesCompression(t *testing.T) {
	var seq int
	prefixes := map[string]string{}
	page := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, `<script nonce="%s"></script>%s`, cspNonce(w.Header()), strings.Repeat("x", 2000))
	})
	h := GzipPublicPages("vukem_session", prefixes)(RenderCache("vukem_session", prefixes)(page))

	var bodies, nonces []string
	for range 2 {
		seq++
		nonce := fmt.Sprintf("N%021d", seq)
		req := httptest.NewRequest(http.MethodGet, "/games", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()
		rec.Header().Set("Content-Security-Policy", "script-src 'nonce-"+nonce+"'")
		h.ServeHTTP(rec, req)
		nonces = append(nonces, nonce)
		bodies = append(bodies, gunzipBody(t, rec))
	}

	if !strings.Contains(bodies[1], nonces[1]) || strings.Contains(bodies[1], nonces[0]) {
		t.Fatalf("second (cached) response must carry its own nonce: %.60s", bodies[1])
	}
}

func TestGzipPublicPages_ClientWithoutGzipGetsPlain(t *testing.T) {
	rec := httptest.NewRecorder()
	gzipPageChain("", "", 0).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/games", nil))

	if rec.Header().Get("Content-Encoding") != "" {
		t.Fatalf("Content-Encoding = %q without Accept-Encoding", rec.Header().Get("Content-Encoding"))
	}
}
