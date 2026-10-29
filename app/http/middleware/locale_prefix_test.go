package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
)

func TestLocalePrefix(t *testing.T) {
	t.Parallel()

	cfg := middleware.LocalePrefixConfig{
		Prefixes: map[string]string{"pt-br": "pt-BR"},
		Query:    "_locale",
		Cookie:   "locale",
	}

	for _, tc := range []struct {
		name, method, target, cookie, inertia string
		wantStatus                            int
		wantLoc, wantPath, wantQuery          string
	}{
		{"prefix is stripped", "GET", "/pt-br/items?a=1", "", "", 200, "", "/items", "_locale=pt-BR&a=1"},
		{"prefix is case-insensitive", "GET", "/PT-BR/", "", "", 200, "", "/", "_locale=pt-BR"},
		{"bare prefix is root", "GET", "/pt-br", "", "", 200, "", "/", "_locale=pt-BR"},
		{"lookalike is not a prefix", "GET", "/pt-brx/items", "", "", 200, "", "/pt-brx/items", ""},
		{"client locale param is dropped", "GET", "/items?_locale=pt-BR", "", "", 200, "", "/items", ""},
		{"client locale param cannot beat prefix", "GET", "/pt-br/?_locale=en", "", "", 200, "", "/", "_locale=pt-BR"},
		{"cookie redirects root", "GET", "/", "pt-BR", "", 307, "/pt-br/", "", ""},
		{"cookie redirect keeps query", "GET", "/items?q=a", "pt-BR", "", 307, "/pt-br/items?q=a", "", ""},
		{"cookie redirect cannot leave the host", "GET", "http://localhost//evil", "pt-BR", "", 307, "/pt-br/evil", "", ""},
		{"default locale cookie does not redirect", "GET", "/", "en", "", 200, "", "/", ""},
		{"unknown cookie does not redirect", "GET", "/", "xx", "", 200, "", "/", ""},
		{"post is not redirected", "POST", "/login", "pt-BR", "", 200, "", "/login", ""},
		{"inertia visit is not redirected", "GET", "/items", "pt-BR", "true", 200, "", "/items", ""},
		{"asset is not redirected", "GET", "/build/app.js", "pt-BR", "", 200, "", "/build/app.js", ""},
		{"file is not redirected", "GET", "/robots.txt", "pt-BR", "", 200, "", "/robots.txt", ""},
		{"health is not redirected", "GET", "/up", "pt-BR", "", 200, "", "/up", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotPath, gotQuery, gotOriginal string

			// RestoreURL is route-level: observe the stripped URL first.
			h := middleware.LocalePrefix(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
				middleware.RestoreURL(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
					gotOriginal = r.URL.Path
				})).ServeHTTP(w, r)
			}))

			r := httptest.NewRequestWithContext(t.Context(), tc.method, tc.target, http.NoBody)
			if tc.cookie != "" {
				r.AddCookie(&http.Cookie{Name: "locale", Value: tc.cookie})
			}

			if tc.inertia != "" {
				r.Header.Set("X-Inertia", tc.inertia)
			}

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, r)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status %d, want %d", rec.Code, tc.wantStatus)
			}

			if tc.wantStatus == 307 {
				if got := rec.Header().Get("Location"); got != tc.wantLoc {
					t.Fatalf("Location %q, want %q", got, tc.wantLoc)
				}

				return
			}

			if gotPath != tc.wantPath || gotQuery != tc.wantQuery {
				t.Fatalf("handler saw %q?%q, want %q?%q", gotPath, gotQuery, tc.wantPath, tc.wantQuery)
			}

			if want := "/pt-br/items"; tc.target == "/pt-br/items?a=1" && gotOriginal != want {
				t.Fatalf("restored path %q, want %q", gotOriginal, want)
			}
		})
	}
}
