package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexGalhardo/vukem/i18n"

	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
)

func TestLocaleCookieRedirect(t *testing.T) {
	t.Parallel()

	prefixes := map[string]string{"pt-br": "pt-BR"}
	cfg := middleware.LocalePrefixConfig{Prefixes: prefixes, Cookie: "locale"}
	h := i18n.StripPathPrefix(prefixes)(middleware.LocaleCookieRedirect(cfg)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/login-redirect" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)

				return
			}

			w.WriteHeader(http.StatusOK)
		})))

	cases := []struct {
		name, method, path, cookie, inertia string
		wantStatus                          int
		wantLocation                        string
	}{
		{"cookie sends a page navigation to the prefix", "GET", "/games?x=1", "pt-BR", "", 307, "/pt-br/games?x=1"},
		{"default locale cookie stays", "GET", "/games", "en", "", 200, ""},
		{"no cookie stays", "GET", "/games", "", "", 200, ""},
		{"inertia visit is not redirected", "GET", "/games", "pt-BR", "true", 200, ""},
		{"post is not redirected", "POST", "/games", "pt-BR", "", 200, ""},
		{"health check is not redirected", "GET", "/up", "pt-BR", "", 200, ""},
		{"asset is not redirected", "GET", "/logo.svg", "pt-BR", "", 200, ""},
		{"prefixed request redirects keep the prefix", "POST", "/pt-br/login-redirect", "", "", 303, "/pt-br/login"},
		{"unprefixed request redirects are untouched", "POST", "/login-redirect", "", "", 303, "/login"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.cookie != "" {
				req.AddCookie(&http.Cookie{Name: "locale", Value: tc.cookie})
			}

			if tc.inertia != "" {
				req.Header.Set("X-Inertia", tc.inertia)
			}

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus || rec.Header().Get("Location") != tc.wantLocation {
				t.Fatalf(
					"got %d %q, want %d %q",
					rec.Code,
					rec.Header().Get("Location"),
					tc.wantStatus,
					tc.wantLocation,
				)
			}
		})
	}
}
