package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
)

func TestCanonicalHost(t *testing.T) {
	h, err := middleware.CanonicalHost("https://hexascore.app", "api.hexascore.app")
	if err != nil {
		t.Fatal(err)
	}
	next := h(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) }))

	cases := []struct {
		name, host, uri string
		wantStatus      int
		wantLoc         string
	}{
		{"www to apex", "www.hexascore.app", "/pt-br/items?q=a&page=2", 301, "https://hexascore.app/pt-br/items?q=a&page=2"},
		{"railway to apex", "app-production-785f.up.railway.app", "/x", 301, "https://hexascore.app/x"},
		{"root keeps slash", "www.hexascore.app", "/", 301, "https://hexascore.app/"},
		{"apex untouched", "hexascore.app", "/x", http.StatusTeapot, ""},
		{"apex case and port", "HexaScore.app:443", "/x", 301, "https://hexascore.app/x"},
		{"api untouched", "api.hexascore.app", "/v1/items", http.StatusTeapot, ""},
		{"api with port untouched", "API.hexascore.app:443", "/v1/items", http.StatusTeapot, ""},
		{"up exempt on www", "www.hexascore.app", "/up", http.StatusTeapot, ""},
		{"up exempt on railway healthcheck", "healthcheck.railway.app", "/up", http.StatusTeapot, ""},
		{"up prefix is not exempt", "www.hexascore.app", "/upload", 301, "https://hexascore.app/upload"},
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, c.uri, nil)
		req.Host = c.host
		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, req)
		if rec.Code != c.wantStatus || rec.Header().Get("Location") != c.wantLoc {
			t.Errorf("%s: got %d %q, want %d %q", c.name, rec.Code, rec.Header().Get("Location"), c.wantStatus, c.wantLoc)
		}
	}
}

func TestCanonicalHostLocalPort(t *testing.T) {
	h, _ := middleware.CanonicalHost("http://localhost:8080", "api.hexascore.app")
	next := h(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "localhost:8080"
	rec := httptest.NewRecorder()
	next.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("canonical host with port redirected: %d", rec.Code)
	}
}

func TestCanonicalHostRejectsBadURL(t *testing.T) {
	if _, err := middleware.CanonicalHost("not a url", "api.hexascore.app"); err == nil {
		t.Error("want error for APP_URL without a host")
	}
}
