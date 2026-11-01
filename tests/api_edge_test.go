//go:build integration

package tests

import (
	"net/http"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/uuid"
)

func TestAPIEdge(t *testing.T) {
	r := newAPICatalog(t)

	t.Run("host_routing", func(*testing.T) {
		r.client.GetJSON("http://api.hexascore.app/v1/games").AssertOK().AssertJSONCount("data", 1)
		r.client.GetJSON("http://api.hexascore.app:8080/v1/games").AssertOK()
		r.client.GetJSON("http://api.hexascore.app/api/v1/games").AssertOK()
		// The alias exists only on the API host.
		r.client.GetJSON("http://hexascore.app/v1/games").AssertNotFound()
	})

	t.Run("cors", func(t *testing.T) {
		res := r.client.WithHeader("Origin", "https://example.org").GetJSON("/api/v1/games").AssertOK()
		h := res.Recorder().Header()

		if h.Get("Access-Control-Allow-Origin") != "*" || h.Get("Access-Control-Allow-Credentials") != "" {
			t.Fatalf("CORS headers = %v, want origin * and no credentials", h)
		}

		if !strings.Contains(h.Get("Access-Control-Expose-Headers"), "X-Quota-Remaining") {
			t.Fatalf("quota headers not exposed: %v", h)
		}

		pre := r.app.Client(t).WithHeader("Origin", "https://example.org").
			WithHeader("Access-Control-Request-Method", "GET").
			Do("OPTIONS", "/api/v1/games", nil, "")
		if s := pre.Recorder().Code; s != 204 {
			t.Fatalf("preflight status = %d, want 204", s)
		}

		// SDD 13: /app/v1 is first-party only, so no CORS, but its problems are still coded.
		app := r.app.Client(t).WithHeader("Origin", "https://example.org").GetJSON("/app/v1/nope").
			AssertStatus(404).AssertJSONPath("type", "https://hexascore.app/problems/not-found")
		if o := app.Recorder().Header().Get("Access-Control-Allow-Origin"); o != "" {
			t.Fatalf("/app/v1 Access-Control-Allow-Origin = %q, want none", o)
		}
	})

	t.Run("problem_codes", func(t *testing.T) {
		const base = "https://hexascore.app/problems/"

		for _, tc := range []struct {
			path, code string
			status     int
		}{
			{"/api/v1/games?bogus=1", "invalid-query", 400},
			{"/api/v1/games?cursor=garbage", "invalid-cursor", 400},
			{"/api/v1/films", "invalid-query", 400},
			{"/api/v1/films/" + uuid.NewV7().String(), "invalid-query", 400},
			{"/api/v1/films/" + uuid.NewV7().String() + "/activity", "invalid-query", 400},
			{"/api/v1/games/" + uuid.NewV7().String(), "not-found", 404},
			{"/api/v1/nope/a/b/c", "not-found", 404},
		} {
			r.client.GetJSON(tc.path).AssertStatus(tc.status).
				AssertJSONPath("code", tc.code).AssertJSONPath("type", base+tc.code)
		}

		r.app.Client(t).GetJSON("/api/v1/games").AssertStatus(401).AssertJSONPath("code", "unauthorized")
	})

	// server URL + documented path must reach the same handler as the /api/v1 route (SDD 05 section 6.1).
	t.Run("openapi_urls_resolve", func(t *testing.T) {
		doc, err := r.app.OpenAPI()
		if err != nil {
			t.Fatal(err)
		}

		server := strings.Replace(doc.Servers[0].URL, "https://", "http://", 1)
		fill := strings.NewReplacer("{type}", "games", "{id}", r.game)

		for path, ops := range doc.Paths {
			for method := range ops {
				p := fill.Replace(path)
				documented := r.client.Do(strings.ToUpper(method), server+p, nil, "").Status()
				route := r.client.Do(strings.ToUpper(method), "/api/v1"+p, nil, "").Status()

				if documented != route || documented == http.StatusMethodNotAllowed {
					t.Errorf("%s %s%s = %d, /api/v1%s = %d", method, server, path, documented, p, route)
				}
			}
		}

		r.client.GetJSON(server + "/games").AssertOK()
		r.client.GetJSON(server + "/me/usage").AssertOK()
	})

	t.Run("openapi_servers", func(t *testing.T) {
		doc, err := r.app.OpenAPI()
		if err != nil {
			t.Fatal(err)
		}

		if len(doc.Servers) != 1 || doc.Servers[0].URL != "https://api.hexascore.app/v1" {
			t.Fatalf("servers = %+v", doc.Servers)
		}
	})
}

func TestAPIEdgeHostOnlyAPI(t *testing.T) {
	r := newAPICatalog(t)
	const api = "http://api.hexascore.app"

	t.Run("index", func(t *testing.T) {
		res := r.app.Client(t).Get(api + "/").AssertOK()
		if ct := res.Recorder().Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Fatalf("Content-Type = %q", ct)
		}

		res.AssertJSONPath("name", "HexaScore API").AssertJSONPath("version", "v1").
			AssertJSONPath("openapi", "https://api.hexascore.app/v1/openapi.json")
	})

	t.Run("up_stays", func(t *testing.T) {
		r.app.Client(t).Get(api + "/up").AssertOK()
	})

	t.Run("openapi_stays", func(t *testing.T) {
		r.app.Client(t).Get(api + "/v1/openapi.json").AssertOK()
		r.app.Client(t).Get(api + "/api/v1/openapi.json").AssertOK()
	})

	t.Run("unknown_path_is_problem", func(t *testing.T) {
		res := r.app.Client(t).Get(api+"/nope").AssertStatus(404).
			AssertJSONPath("code", "not-found").AssertJSONPath("type", "https://hexascore.app/problems/not-found")
		if ct := res.Recorder().Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
			t.Fatalf("Content-Type = %q", ct)
		}
	})

	t.Run("web_host_does_not_serve_v1", func(*testing.T) {
		r.app.Client(t).Get("http://hexascore.app/v1/openapi.json").AssertNotFound()
		r.app.Client(t).Get("http://hexascore.app/v1/games").AssertNotFound()
	})
}

// A05: the API host must never answer with the web stack.
func TestSecurity_A05_ApiHostServesNoWebRoutes(t *testing.T) {
	r := newAPICatalog(t)

	for _, tc := range []struct{ name, path, accept string }{
		{"login", "/login", "text/html"},
		{"admin", "/admin", "text/html"},
		{"home_html", "/", "text/html"},
		{"docs", "/docs/reference", "text/html"},
		{"inertia", "/games", "text/html"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := r.app.Client(t).WithHeader("Accept", tc.accept)
			if tc.name == "inertia" {
				c = c.WithHeader("X-Inertia", "true").WithHeader("X-Inertia-Version", "x")
			}

			rec := c.Get("http://api.hexascore.app" + tc.path).Recorder()
			h := rec.Header()

			if sc := h.Values("Set-Cookie"); len(sc) != 0 {
				t.Fatalf("Set-Cookie = %v", sc)
			}

			if ct := h.Get("Content-Type"); strings.Contains(ct, "html") || h.Get("X-Inertia") != "" {
				t.Fatalf("%s answered %q (X-Inertia %q)", tc.path, ct, h.Get("X-Inertia"))
			}

			if strings.Contains(strings.ToLower(rec.Body.String()), "<html") ||
				strings.Contains(rec.Body.String(), "data-page") {
				t.Fatalf("%s body looks like a page", tc.path)
			}
		})
	}
}
