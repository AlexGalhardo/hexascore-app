package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	vmw "github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/http/problems"
)

const (
	apiPrefix    = "/api/v1"
	appPrefix    = "/app/v1"
	hostV1Prefix = "/v1"
)

// APIEdge is the public API edge. Install it with Router.UseServer before every other server middleware. It (1)
// rewrites /v1/... to /api/v1/... when the request Host is apiHost (port ignored) and (2) answers CORS for /api/v1
// with origin "*" and no credentials. Problem codes come from ProblemDecorator.
// workaround(vukem): VF-NEW-ADOPT-vukem-0.36.0 (router.Group.Only cannot keep /up on the API host).
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
	site := strings.TrimSuffix(appURL, "/")
	index := mustJSON(map[string]any{
		"name":    "HexaScore API",
		"version": "v1",
		"docs":    site + "/docs/reference",
		"openapi": "https://" + apiHost + "/v1/openapi.json",
	}, nil)

	return func(next http.Handler) http.Handler {
		api := cors(next)

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

			// /app/v1 never gets CORS headers (SDD 13: first-party clients only).
			if underPrefix(r.URL.Path, apiPrefix) {
				api.ServeHTTP(w, r)
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

// ProblemDecorator gives every problem+json under /api/v1 and /app/v1, and every one on apiHost, a catalogue
// `code` and a `type` ending in it (SDD 05 section 6.1), including the 404, 405 and typed-route 400s vukem raises.
// A handler picks a code other than the status default with errs.HTTPError.WithExtension("code", ...).
func ProblemDecorator(apiHost, problemBase string) func(*web.Ctx, *web.Problem) {
	base := strings.TrimSuffix(problemBase, "/") + "/"

	return func(c *web.Ctx, p *web.Problem) {
		r := c.Request()
		if hostOnly(r.Host) != apiHost && !underPrefix(r.URL.Path, apiPrefix) && !underPrefix(r.URL.Path, appPrefix) {
			return
		}

		code, has := p.Extensions["code"].(string)
		if !has {
			code = problems.ForStatus(p.Status)
		}

		if code == "" {
			return
		}

		if p.Extensions == nil {
			p.Extensions = map[string]any{}
		}

		p.Extensions["code"] = code
		p.Type = base + code
	}
}

func mustJSON(doc map[string]any, fallback []byte) []byte {
	if out, err := json.Marshal(doc); err == nil {
		return out
	}

	return fallback
}
