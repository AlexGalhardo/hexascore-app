//go:build integration

package tests

import (
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
			{"/api/v1/games/" + uuid.NewV7().String(), "not-found", 404},
			{"/api/v1/nope/a/b/c", "not-found", 404},
		} {
			r.client.GetJSON(tc.path).AssertStatus(tc.status).
				AssertJSONPath("code", tc.code).AssertJSONPath("type", base+tc.code)
		}

		r.app.Client(t).GetJSON("/api/v1/games").AssertStatus(401).AssertJSONPath("code", "unauthorized")
	})

	t.Run("openapi_servers", func(t *testing.T) {
		doc, err := r.app.OpenAPI()
		if err != nil {
			t.Fatal(err)
		}

		if len(doc.Servers) != 1 || doc.Servers[0].URL != "https://api.hexascore.app" {
			t.Fatalf("servers = %+v", doc.Servers)
		}
	})
}
