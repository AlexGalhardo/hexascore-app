package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// minGzipSize is where a compressed body beats the encoding header overhead.
const minGzipSize = 1024

// GzipStatic compresses immutable /build assets (JS, CSS, SVG, fonts, source
// maps) when the client accepts gzip. HTML is never compressed, and a
// response that sets a cookie is passed through, so a secret in a page cannot
// leak through a compression oracle (CWE-200, BREACH).
func GzipStatic() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !gzipAsset(r) {
				next.ServeHTTP(w, r)

				return
			}

			w.Header().Add("Vary", "Accept-Encoding")
			gw := &gzipStaticWriter{ResponseWriter: w}
			defer gw.finish()
			next.ServeHTTP(gw, r)
		})
	}
}

func gzipAsset(r *http.Request) bool {
	if r.Method == http.MethodHead || !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		return false
	}

	p := r.URL.Path
	if !strings.HasPrefix(p, "/build/") {
		return false
	}

	dot := strings.LastIndex(p, ".")
	if dot < 0 {
		return false
	}

	switch strings.ToLower(p[dot:]) {
	case ".js", ".css", ".svg", ".map", ".woff", ".woff2":
		return true
	default:
		return false
	}
}

// gzipStaticWriter buffers one asset, then either gzips it or copies it unchanged.
type gzipStaticWriter struct {
	http.ResponseWriter

	status int
	buf    []byte
	wrote  bool
}

func (g *gzipStaticWriter) WriteHeader(code int) {
	if g.status == 0 {
		g.status = code
	}
}

func (g *gzipStaticWriter) Write(b []byte) (int, error) {
	if g.wrote {
		return g.ResponseWriter.Write(b)
	}

	if g.status == 0 {
		g.status = http.StatusOK
	}

	g.buf = append(g.buf, b...)

	return len(b), nil
}

func (g *gzipStaticWriter) Flush() {
	g.finish()

	_ = http.NewResponseController(g.ResponseWriter).Flush()
}

func (g *gzipStaticWriter) finish() {
	if g.wrote {
		return
	}

	g.wrote = true

	if g.status == 0 {
		return
	}

	if g.compressible() {
		g.Header().Set("Content-Encoding", "gzip")
		g.Header().Del("Content-Length")
		g.ResponseWriter.WriteHeader(g.status)
		zw := gzip.NewWriter(g.ResponseWriter)
		_, _ = zw.Write(g.buf)
		_ = zw.Close()

		return
	}

	g.ResponseWriter.WriteHeader(g.status)
	_, _ = g.ResponseWriter.Write(g.buf)
}

func (g *gzipStaticWriter) compressible() bool {
	if len(g.buf) < minGzipSize || g.status < 200 || g.status >= 300 {
		return false
	}

	h := g.Header()
	if h.Get("Set-Cookie") != "" || h.Get("Content-Encoding") != "" {
		return false
	}

	ct := strings.ToLower(h.Get("Content-Type"))

	return !strings.Contains(ct, "text/html")
}
