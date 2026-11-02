//go:build integration

package tests

import (
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestAPIReference_PublicInProduction(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL": pgtest.Database(t, ""), "APP_ENV": "production", "APP_URL": "https://example.test/",
		"RATE_LIMIT_STORE": "cache", "CACHE_STORE": "database",
		"STRIPE_MODE":        "sandbox",
		"STRIPE_KEY_SANDBOX": stripeKey, "STRIPE_WEBHOOK_SECRET_SANDBOX": webhookSecret,
		"STRIPE_PRICE_HOBBY_SANDBOX": "price_hobby", "STRIPE_PRICE_PRO_SANDBOX": "price_pro",
	}})

	app.Client(t).Get("https://example.test/docs/reference").AssertOK()
	app.Client(t).Get("https://example.test/docs/reference/scalar.js").AssertOK()

	var doc struct {
		OpenAPI string                               `json:"openapi"`
		Paths   map[string]map[string]map[string]any `json:"paths"`
		Info    map[string]string                    `json:"info"`
	}
	app.Client(t).GetJSON("https://example.test/api/v1/openapi.json").AssertOK().JSON(&doc)

	if len(doc.OpenAPI) < 3 || doc.OpenAPI[:3] != "3.1" {
		t.Errorf("openapi = %q, want 3.1.x", doc.OpenAPI)
	}

	// The server URL already ends in /v1 (api host form), so no path repeats /api/v1.
	for path := range doc.Paths {
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/v1/") || !strings.HasPrefix(path, "/") {
			t.Errorf("path %q is not relative to the /v1 server", path)
		}
	}

	// The reference sidebar shows readable titles ("List items"), never the operationId (owner 2026-11-02).
	for path, ops := range doc.Paths {
		for method, op := range ops {
			sum, _ := op["summary"].(string)
			if id, _ := op["operationId"].(string); sum == "" || sum == id {
				t.Errorf("%s %s: summary %q, want a readable title, not the operationId", method, path, sum)
			}
		}
	}

	// The app document stays private: it maps the token surface.
	app.Client(t).GetJSON("https://example.test/app/v1/openapi.json").AssertStatus(404)
}
