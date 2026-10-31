package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/jpeg"
	"io"
	"log/slog"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/storage"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
)

// TestSecurity_Community_A01_SourceURLSSRF: a cited URL that lands on loopback
// never connects. http, hexascore.app and shorteners are refused before dial.
func TestSecurity_Community_A01_SourceURLSSRF(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })

	var hits atomic.Int32
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			hits.Add(1)
			_ = conn.Close()
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if got := community.SourceStatus(ctx, "https://"+ln.Addr().String()+"/latest/meta-data"); got != "blocked" {
		t.Fatalf("loopback status %q, hits %d", got, hits.Load())
	}
	if hits.Load() != 0 {
		t.Fatalf("loopback accepted %d connections", hits.Load())
	}

	client := community.SourceClient()
	if client.Timeout != 5*time.Second {
		t.Fatalf("source client timeout %s", client.Timeout)
	}

	for _, raw := range []string{
		"http://127.0.0.1/secret",
		"https://hexascore.app/items/1",
		"https://evil.hexascore.app/a",
		"https://bit.ly/abc",
	} {
		if got := community.SourceStatus(ctx, raw); got != "refused" {
			t.Fatalf("%s status %q", raw, got)
		}
	}
}

// TestSecurity_Community_A04_HoneypotLogged: a filled website field is a generic
// 422 and a community.honeypot security event. The field value is not logged.
func TestSecurity_Community_A04_HoneypotLogged(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	req := httptest.NewRequest(http.MethodPost, "/suggestions", http.NoBody)
	req = req.WithContext(web.WithServices(req.Context(), &web.Services{Log: logger}))

	err := community.RejectHoneypot(context.Background(), nil, req, 42, "https://spam.example", time.Now())
	var problem *community.Problem
	if !errors.As(err, &problem) || problem.Status != http.StatusUnprocessableEntity || problem.Code != "validation" {
		t.Fatalf("honeypot err %#v", err)
	}
	if strings.Contains(buf.String(), "spam.example") {
		t.Fatalf("honeypot value was logged: %s", buf.String())
	}
	if !strings.Contains(buf.String(), `"event":"community.honeypot"`) {
		t.Fatalf("log %s", buf.String())
	}

	buf.Reset()
	if err := community.RejectHoneypot(context.Background(), nil, req, 42, "   ", time.Now()); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Fatalf("blank honeypot logged %s", buf.String())
	}
}

func TestCommunityBurst(t *testing.T) {
	t.Parallel()

	f := func(_ string, limit int, window time.Duration) ratelimit.Limiter {
		return ratelimit.NewMemory(limit, window)
	}
	h := community.DraftBurst(f, "web")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := 1; i <= 30; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/suggestions", http.NoBody))
		if rec.Code != http.StatusNoContent {
			t.Fatalf("draft %d status %d", i, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/suggestions", http.NoBody))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("draft 31 status %d", rec.Code)
	}

	if limits.CommunitySubmit.Limit != 10 || limits.CommunitySubmit.Window != time.Minute {
		t.Fatalf("submit burst %+v", limits.CommunitySubmit)
	}
	if limits.CommunityUpload.Limit != 10 || limits.CommunityUpload.Window != time.Hour {
		t.Fatalf("upload burst %+v", limits.CommunityUpload)
	}
}

// TestCoverUpload: AVIF is refused, a file over 5 MiB is 413, and a short
// side under 300 px is 422. A 300 px JPEG is stored under community/covers/.
func TestCoverUpload(t *testing.T) {
	t.Parallel()

	t.Run("avif", func(t *testing.T) {
		t.Parallel()
		disk := newPathDisk()
		rec := postCover(t, disk, "cover.avif", avifHeader)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("avif status %d", rec.Code)
		}
		disk.assertEmpty(t)
	})

	t.Run("too big", func(t *testing.T) {
		t.Parallel()
		disk := newPathDisk()
		body := make([]byte, (5<<20)+1)
		copy(body, []byte{0xff, 0xd8, 0xff})
		rec := postCover(t, disk, "cover.jpg", body)
		if rec.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("oversize status %d", rec.Code)
		}
		disk.assertEmpty(t)
	})

	t.Run("short side", func(t *testing.T) {
		t.Parallel()
		disk := newPathDisk()
		rec := postCover(t, disk, "cover.jpg", jpegOf(t, 299, 400))
		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("short side status %d", rec.Code)
		}
		disk.assertGone(t)
	})

	t.Run("stored", func(t *testing.T) {
		t.Parallel()
		disk := newPathDisk()
		rec := postCover(t, disk, "cover.jpg", jpegOf(t, 300, 320))
		if rec.Code != http.StatusOK {
			t.Fatalf("cover status %d", rec.Code)
		}
		var cover community.Cover
		if err := json.Unmarshal(rec.Body.Bytes(), &cover); err != nil {
			t.Fatal(err)
		}
		if cover.Width != 300 || cover.Height != 320 || !strings.HasPrefix(cover.Key, "community/covers/") {
			t.Fatalf("cover %+v", cover)
		}
	})
}

type pathDisk struct {
	*storage.Memory
	paths []string
}

func newPathDisk() *pathDisk {
	return &pathDisk{Memory: storage.NewMemory("")}
}

func (d *pathDisk) Put(ctx context.Context, p string, r io.Reader) error {
	if err := d.Memory.Put(ctx, p, r); err != nil {
		return err
	}
	d.paths = append(d.paths, p)

	return nil
}

func (d *pathDisk) assertEmpty(t *testing.T) {
	t.Helper()
	if len(d.paths) != 0 {
		t.Fatalf("stored %v", d.paths)
	}
}

func (d *pathDisk) assertGone(t *testing.T) {
	t.Helper()
	if len(d.paths) == 0 {
		t.Fatal("cover was never stored")
	}
	for _, p := range d.paths {
		ok, err := d.Exists(context.Background(), p)
		if err != nil || ok {
			t.Fatalf("cover %s exists=%v err=%v", p, ok, err)
		}
	}
}

func postCover(t *testing.T, disk storage.Disk, name string, body []byte) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("cover", name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/suggestions/uploads/cover", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	community.UploadCover(rec, req, disk)

	return rec
}

func jpegOf(t *testing.T, width, height int) []byte {
	t.Helper()

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, width, height)), nil); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

var avifHeader = []byte{
	0x00, 0x00, 0x00, 0x1c, 'f', 't', 'y', 'p', 'a', 'v', 'i', 'f',
	0x00, 0x00, 0x00, 0x00, 'a', 'v', 'i', 'f', 'm', 'i', 'f', '1', 'm', 'i', 'a', 'f',
}
