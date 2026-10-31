package appapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestKnownType_UnknownSegmentIsProblemJSON(t *testing.T) {
	t.Parallel()

	h := KnownType(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("unknown segment reached the catalogue")
	}))

	for _, path := range []string{"/app/v1/0", "/app/v1/0/0"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s status = %d", path, rec.Code)
		}

		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
			t.Fatalf("%s content-type = %q body %s", path, ct, rec.Body.String())
		}
	}
}

func TestKnownType_AllowsCatalogueSections(t *testing.T) {
	t.Parallel()

	var path string
	h := KnownType(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/app/v1/games", nil))

	if path != "/app/v1/games" || rec.Code != http.StatusNoContent {
		t.Fatalf("path %q status %d", path, rec.Code)
	}
}
