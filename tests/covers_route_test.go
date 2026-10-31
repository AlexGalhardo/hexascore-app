package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/app/controllers"
)

func coversServer(t *testing.T) http.Handler {
	t.Helper()

	disk := storage.NewMemory("")
	if err := storage.PutBytes(t.Context(), disk, "covers/ok.jpg", []byte("jpeg-bytes")); err != nil {
		t.Fatal(err)
	}
	// An object outside the covers/ prefix must never be reachable through this route.
	if err := storage.PutBytes(t.Context(), disk, "secret.txt", []byte("secret")); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /covers/{key...}", controllers.CoversFrom(disk))

	return mux
}

func TestCoversRoute_ServesImmutable(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	coversServer(t).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/covers/ok.jpg", nil))

	if rec.Code != http.StatusOK || rec.Body.String() != "jpeg-bytes" {
		t.Fatalf("got %d %q", rec.Code, rec.Body.String())
	}

	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("Cache-Control = %q", got)
	}
}

func TestCoversRoute_MissingIs404(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	coversServer(t).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/covers/nope.jpg", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCoversRoute_NoBucketIs404(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.Handle("GET /covers/{key...}", controllers.CoversFrom(nil))

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/covers/ok.jpg", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestSecurity_A01_CoverKeyTraversal(t *testing.T) {
	t.Parallel()

	srv := coversServer(t)

	for _, target := range []string{
		"/covers/..%2Fsecret.txt",
		"/covers/%2e%2e/secret.txt",
		"/covers/a/../../secret.txt",
		"/covers/..\\secret.txt",
		"/covers/..%5Csecret.txt",
		"/covers/.hidden",
		"/covers//secret.txt",
	} {
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))

		if rec.Body.String() == "secret" || rec.Code == http.StatusOK {
			t.Errorf("%s leaked: %d %q", target, rec.Code, rec.Body.String())
		}
	}
}
