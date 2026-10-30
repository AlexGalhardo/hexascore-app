package middleware_test

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
)

func TestGzipStatic_HTMLUncompressedAndBuildJSCompressed(t *testing.T) {
	js := "console.log(" + strings.Repeat("'asset';", 200) + ");"
	html := "<!doctype html><html><body>" + strings.Repeat("page ", 400) + "</body></html>"

	h := middleware.GzipStatic()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = io.WriteString(w, html)
		case "/build/app.js":
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			_, _ = io.WriteString(w, js)
		case "/build/cookie.js":
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "secret"})
			_, _ = io.WriteString(w, js)
		case "/build/page.html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = io.WriteString(w, html)
		case "/build/trap.js":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = io.WriteString(w, html)
		default:
			http.NotFound(w, r)
		}
	}))

	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)

	pageEnc, pageBody := getGzip(t, srv.URL+"/")
	if pageEnc != "" {
		t.Fatalf("HTML Content-Encoding = %q, want empty", pageEnc)
	}

	if pageBody != html {
		t.Fatalf("HTML body changed (%d bytes)", len(pageBody))
	}

	assetEnc, assetBody := getGzip(t, srv.URL+"/build/app.js")
	if assetEnc != "gzip" {
		t.Fatalf("/build/app.js Content-Encoding = %q, want gzip", assetEnc)
	}

	zr, err := gzip.NewReader(strings.NewReader(assetBody))
	if err != nil {
		t.Fatal(err)
	}

	got, err := io.ReadAll(zr)
	_ = zr.Close()
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != js {
		t.Fatalf("gunzipped JS = %q", string(got[:40]))
	}

	cookieEnc, _ := getGzip(t, srv.URL+"/build/cookie.js")
	if cookieEnc != "" {
		t.Fatalf("Set-Cookie asset Content-Encoding = %q, want empty", cookieEnc)
	}

	htmlEnc, _ := getGzip(t, srv.URL+"/build/page.html")
	if htmlEnc != "" {
		t.Fatalf("HTML under /build Content-Encoding = %q, want empty", htmlEnc)
	}

	trapEnc, _ := getGzip(t, srv.URL+"/build/trap.js")
	if trapEnc != "" {
		t.Fatalf("text/html disguised as JS Content-Encoding = %q, want empty", trapEnc)
	}
}

func getGzip(t *testing.T, url string) (encoding, body string) {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Accept-Encoding", "gzip")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = res.Body.Close() }()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	return res.Header.Get("Content-Encoding"), string(b)
}
