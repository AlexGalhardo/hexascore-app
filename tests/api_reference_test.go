//go:build integration

package tests

import (
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestAPIReference_PublicInProduction(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"DATABASE_URL": pgtest.Database(t, ""), "APP_ENV": "production", "APP_URL": "https://example.test/",
		"STRIPE_KEY": stripeKey, "STRIPE_WEBHOOK_SECRET": webhookSecret,
		"STRIPE_PRICE_PRO": "price_pro", "STRIPE_PRICE_BUSINESS": "price_business",
	}})

	app.Client(t).Get("/docs/reference").AssertOK()
	app.Client(t).Get("/docs/reference/scalar.js").AssertOK()

	var doc struct {
		OpenAPI string            `json:"openapi"`
		Paths   map[string]any    `json:"paths"`
		Info    map[string]string `json:"info"`
	}
	app.Client(t).GetJSON("/api/v1/openapi.json").AssertOK().JSON(&doc)

	if len(doc.OpenAPI) < 3 || doc.OpenAPI[:3] != "3.1" {
		t.Errorf("openapi = %q, want 3.1.x", doc.OpenAPI)
	}

	for path := range doc.Paths {
		if len(path) < 7 || path[:7] != "/api/v1" {
			t.Errorf("path %q is outside /api/v1", path)
		}
	}

	// The app document stays private: it maps the token surface.
	app.Client(t).GetJSON("/app/v1/openapi.json").AssertStatus(404)
}
