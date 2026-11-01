package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	vmw "github.com/AlexGalhardo/vukem/middleware"

	"github.com/AlexGalhardo/hexascore-app/app/http/problems"
)

// ProblemCodeHeader lets a handler pick a catalogue code other than the status default (for example
// invalid-cursor on a 400). APIEdge consumes it and never forwards it.
const ProblemCodeHeader = "X-Problem-Code"

const (
	apiPrefix    = "/api/v1"
	appPrefix    = "/app/v1"
	hostV1Prefix = "/v1"
)

// APIEdge is the public API edge, a workaround(vukem) for VF-NEW-HOST-ONLY and VF-NEW-PROBLEM-TYPE. Install it with
// Router.UseServer before every other server middleware. It (1) rewrites /v1/... to /api/v1/... when the request Host
// is apiHost (port ignored), (2) answers CORS for /api/v1 with origin "*" and no credentials, and (3) makes every
// application/problem+json under /api/v1 and /app/v1 carry a catalogue `code` and a `type` ending in it, including
// problems vukem raises itself (404, 405, typed-route 400s), which cannot carry a code or type (VF-NEW-PROBLEM-TYPE).
// On apiHost only the API is served (SDD 05 section 6.1): /v1 and /api/v1, /up, and a JSON index at "/"; every other
// path is a 404 problem+json that never reaches the web stack (no HTML, no session cookie, no Inertia). appURL is the
// web origin the index links to.
func APIEdge(apiHost, appURL, problemBase string) (func(http.Handler) http.Handler, error) {
	cors, err := vmw.CORS(vmw.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodGet, http.MethodHead, http.MethodPost, http.MethodDelete},
		AllowedHeaders: []string{"Authorization", "Content-Type", "If-None-Match", "Accept"},
		ExposedHeaders: []string{
			"X-Request-Id", "RateLimit-Limit", "RateLimit-Remaining", "RateLimit-Reset", "X-Quota-Limit",
			"X-Quota-Remaining", "X-Quota-Reset", "ETag", "Retry-After", "Deprecation", "Sunset", "Link",
		},
		MaxAge: 24 * time.Hour,
	})
	if err != nil {
		return nil, err
	}

	base := strings.TrimSuffix(problemBase, "/") + "/"
	web := strings.TrimSuffix(appURL, "/")
	index := mustJSON(map[string]any{
		"name":    "HexaScore API",
		"version": "v1",
		"docs":    web + "/docs/reference",
		"openapi": "https://" + apiHost + "/v1/openapi.json",
	}, nil)

	return func(next http.Handler) http.Handler {
		coded := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pw := &problemWriter{ResponseWriter: w, base: base}
			next.ServeHTTP(pw, r)
			pw.finish()
		})
		api := cors(coded)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if hostOnly(r.Host) == apiHost {
				switch {
				case underPrefix(r.URL.Path, hostV1Prefix):
					r = r.Clone(r.Context())
					r.URL.Path = apiPrefix + strings.TrimPrefix(r.URL.Path, hostV1Prefix)
					r.URL.RawPath = ""
				case r.URL.Path == "/" && (r.Method == http.MethodGet || r.Method == http.MethodHead):
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Cache-Control", "public, max-age=300")
					_, _ = w.Write(index)

					return
				case r.URL.Path == "/up" || underPrefix(r.URL.Path, apiPrefix):
				default:
					writeNotFound(w, base, r.URL.Path)

					return
				}
			}

			if underPrefix(r.URL.Path, apiPrefix) {
				api.ServeHTTP(w, r)
				return
			}

			// /app/v1 gets coded problems but never CORS headers (SDD 13: first-party clients only).
			if underPrefix(r.URL.Path, appPrefix) {
				coded.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}, nil
}

func writeNotFound(w http.ResponseWriter, base, path string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write(mustJSON(map[string]any{
		"type": base + "not-found", "title": "Not Found", "status": http.StatusNotFound,
		"code": "not-found", "instance": path,
	}, nil))
}

func underPrefix(path, prefix string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func hostOnly(h string) string {
	if i := strings.LastIndexByte(h, ':'); i >= 0 && !strings.Contains(h[i:], "]") {
		h = h[:i]
	}

	return strings.ToLower(h)
}

// problemWriter holds application/problem+json bodies to add the code; every other response streams through.
type problemWriter struct {
	http.ResponseWriter

	base    string
	status  int
	buf     bytes.Buffer
	decided bool
	held    bool
}

func (p *problemWriter) Unwrap() http.ResponseWriter { return p.ResponseWriter }

func (p *problemWriter) WriteHeader(status int) {
	if p.decided {
		return
	}

	p.decided = true
	p.status = status

	if strings.HasPrefix(p.Header().Get("Content-Type"), "application/problem+json") {
		p.held = true
		return
	}

	p.Header().Del(ProblemCodeHeader)
	p.ResponseWriter.WriteHeader(status)
}

func (p *problemWriter) Write(b []byte) (int, error) {
	if !p.decided {
		p.WriteHeader(http.StatusOK)
	}

	if p.held {
		return p.buf.Write(b)
	}

	return p.ResponseWriter.Write(b)
}

func (p *problemWriter) finish() {
	if !p.held {
		return
	}

	h := p.Header()
	forced := h.Get(ProblemCodeHeader)
	h.Del(ProblemCodeHeader)

	body := p.buf.Bytes()

	var doc map[string]any
	if json.Unmarshal(body, &doc) == nil {
		if c, has := doc["code"].(string); has {
			// Handlers raise errs.HTTPError.WithExtension("code", ...); vukem keeps type about:blank.
			if c != "" {
				doc["type"] = p.base + c
				body = mustJSON(doc, body)
			}
		} else {
			code := forced
			if !problems.Known(code) {
				code = problems.ForStatus(p.status)
			}

			if code != "" {
				doc["code"] = code
				doc["type"] = p.base + code

				body = mustJSON(doc, body)
			}
		}
	}

	h.Del("Content-Length")
	p.ResponseWriter.WriteHeader(p.status)
	_, _ = p.ResponseWriter.Write(body)
}

func mustJSON(doc map[string]any, fallback []byte) []byte {
	if out, err := json.Marshal(doc); err == nil {
		return out
	}

	return fallback
}
