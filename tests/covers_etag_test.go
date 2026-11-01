//go:build integration

package tests

import (
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/config"
	"github.com/AlexGalhardo/vukem/drivers/s3"
	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/app/controllers"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

// The in-memory disk has no ETag; drivers/s3 v0.4.0 does (Q-A14), so this runs against the compose RustFS.
func TestCoversRoute_S3ConditionalGetIs304(t *testing.T) {
	disk, err := s3.New(s3.Config{
		Bucket:          "covers-" + strings.ToLower(rand.Text()[:12]),
		Region:          "us-east-1",
		Endpoint:        envOr("S3_ENDPOINT", "http://127.0.0.1:9000"),
		AccessKeyID:     envOr("S3_ACCESS_KEY_ID", "hexascore"),
		SecretAccessKey: config.NewSecret(envOr("S3_SECRET_ACCESS_KEY", "hexascore-local-secret")),
		UsePathStyle:    true,
		InsecureHTTP:    true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := disk.CreateBucket(t.Context()); err != nil {
		t.Skipf("RustFS not reachable (docker compose up -d --wait): %v", err)
	}

	if err := storage.PutBytes(t.Context(), disk, "covers/ok.jpg", []byte("jpeg-bytes")); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = disk.Delete(t.Context(), "covers/ok.jpg")
	})

	mux := http.NewServeMux()
	mux.Handle("GET /covers/{key...}", controllers.CoversFrom(disk))

	first := httptest.NewRecorder()
	mux.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/covers/ok.jpg", nil))

	etag := first.Header().Get("ETag")
	if first.Code != http.StatusOK || etag == "" {
		t.Fatalf("first GET = %d, ETag %q", first.Code, etag)
	}

	req := httptest.NewRequest(http.MethodGet, "/covers/ok.jpg", nil)
	req.Header.Set("If-None-Match", etag)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotModified || rec.Body.Len() != 0 {
		t.Fatalf("conditional GET = %d with %d body bytes", rec.Code, rec.Body.Len())
	}
}
