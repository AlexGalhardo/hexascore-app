package tests

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/billing"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// inventoryApp boots the full route surface: the Postgres-only routes need TEST_DATABASE_URL, so
// without it only the SQLite surface is walked and stale inventory lines are not reported.
func inventoryApp(t *testing.T) (*vukemtest.App, bool) {
	t.Helper()

	env := map[string]string{
		"STRIPE_KEY": stripeKey, "STRIPE_WEBHOOK_SECRET": webhookSecret,
		"STRIPE_PRICE_PRO": "price_pro", "STRIPE_PRICE_BUSINESS": "price_business",
	}

	full := os.Getenv("TEST_DATABASE_URL") != ""
	if full {
		env["DATABASE_URL"] = pgtest.Database(t, "")
	}

	return vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env}), full
}

func readInventory(t *testing.T) map[string]string {
	t.Helper()

	f, err := os.Open("testdata/route_inventory.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	stacks := map[string]string{}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 3 {
			t.Fatalf("inventory line %q: want \"stack METHOD path\"", line)
		}

		stacks[fields[1]+" "+fields[2]] = fields[0]
	}

	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}

	return stacks
}

// stackViolation checks that a route's path and middleware match its §6.1 stack; "" means it does.
func stackViolation(stack, method, path string, mw []string) string {
	has := func(name string) bool { return slices.Contains(mw, name) }
	write := method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions

	switch stack {
	case "admin":
		// AdminGuard (router-wide) enforces role, 2FA and session age on this prefix.
		if path != "/admin" && !strings.HasPrefix(path, "/admin/") {
			return "admin route outside /admin"
		}
	case "api":
		if !strings.HasPrefix(path, "/api/v1/") {
			return "api route outside /api/v1"
		}

		if path != "/api/v1/openapi.json" && !has("api.Quota") {
			return "api route without the API key quota"
		}
	case "app":
		if !strings.HasPrefix(path, "/app/v1/") {
			return "app route outside /app/v1"
		}

		// Search is optional (SDD 10). Every other /app/v1 route still needs a bearer.
		optionalSearch := path == "/app/v1/search" || strings.HasPrefix(path, "/app/v1/search/")
		if !has("auth.Required") && !has("middleware.(*Challenger).Require") &&
			!(optionalSearch && has("auth.Optional")) {
			return "app route without a bearer token or a challenge"
		}
	case "web":
		if path == "/admin" || strings.HasPrefix(path, "/admin/") || strings.HasPrefix(path, "/api/v1/") ||
			strings.HasPrefix(path, "/app/v1/") {
			return "web route inside another stack's prefix"
		}
	default:
		return "unknown stack " + stack
	}

	if (stack == "web" || stack == "admin") && write && !has("csrf.Middleware") {
		return "session write route without CSRF"
	}

	if !has("limits.Body") {
		return "route without limits.Body"
	}

	return ""
}

// TestSecurity_A01_HexaRouteStackInventory walks the router: every route must be listed in
// testdata/route_inventory.txt under one SDD 13 §6.1 stack, and carry that stack's middleware.
func TestSecurity_A01_HexaRouteStackInventory(t *testing.T) {
	t.Parallel()

	app, full := inventoryApp(t)
	stacks := readInventory(t)
	seen := map[string]bool{}

	for _, rt := range app.Router.Routes() {
		key := rt.Method + " " + rt.Path
		seen[key] = true

		stack, ok := stacks[key]
		if !ok {
			t.Errorf("unclassified route %s: add it to tests/testdata/route_inventory.txt with its stack", key)
			continue
		}

		if v := stackViolation(stack, rt.Method, rt.Path, rt.Middleware); v != "" {
			t.Errorf("%s (%s): %s", key, stack, v)
		}
	}

	if !full {
		return
	}

	for key := range stacks {
		if !seen[key] {
			t.Errorf("stale inventory line %s: the route no longer exists", key)
		}
	}
}

func TestSecurity_A01_HexaRouteStackInventory_Rules(t *testing.T) {
	t.Parallel()

	ok := []string{"limits.Body", "csrf.Middleware"}
	if v := stackViolation("app", "GET", "/app/v1/search", []string{"limits.Body", "auth.Optional"}); v != "" {
		t.Errorf("optional search: %s", v)
	}

	for _, c := range []struct{ stack, method, path string }{
		{"admin", "POST", "/admin/users"},
		{"web", "POST", "/login"},
		{"web", "GET", "/api/search/suggest"},
	} {
		if v := stackViolation(c.stack, c.method, c.path, ok); v != "" {
			t.Errorf("%v: %s", c, v)
		}
	}

	for _, c := range []struct {
		stack, method, path string
		mw                  []string
	}{
		{"web", "GET", "/admin/x", ok},
		{"web", "POST", "/login", []string{"limits.Body"}},
		{"app", "GET", "/app/v1/me", ok},
		{"app", "GET", "/app/v1/me", []string{"limits.Body", "auth.Optional"}},
		{"api", "GET", "/api/v1/search", ok},
		{"web", "GET", "/", nil},
		{"webhook", "POST", "/x", ok},
	} {
		if stackViolation(c.stack, c.method, c.path, c.mw) == "" {
			t.Errorf("%s %s %s %v: violation not reported", c.stack, c.method, c.path, c.mw)
		}
	}
}

// TestSecurity_A06_HexaBodyLimits checks the SDD 13 §7.3 table, the per-stack choice, and the limits
// the running app enforces on a form route and on the Stripe webhook.
func TestSecurity_A06_HexaBodyLimits(t *testing.T) {
	t.Parallel()

	for name, c := range map[string]struct{ got, want int64 }{
		"forms":          {limits.FormBody, 64 << 10},
		"suggestion":     {limits.SuggestionBody, 256 << 10},
		"cover":          {limits.CoverBody, 5 << 20},
		"stripe webhook": {limits.StripeWebhookBody, 512 << 10},
		"client-errors":  {limits.ClientErrorsBody, 16 << 10},
		"client-metrics": {limits.ClientMetricsBody, 4 << 10},
		"other":          {limits.DefaultBody, 1 << 20},
	} {
		if c.got != c.want {
			t.Errorf("%s limit = %d, want %d", name, c.got, c.want)
		}
	}

	read := limits.Body(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				return
			}

			t.Error(err)
		}
	}))

	for _, c := range []struct {
		path  string
		limit int64
	}{
		{"/login", limits.FormBody},
		{"/csp-reports", limits.CSPReportsBody},
		{"/pt-br/lists", limits.FormBody},
		{"/admin/users", limits.FormBody},
		{"/api/v1/search", limits.DefaultBody},
		{"/app/v1/lists", limits.DefaultBody},
	} {
		for size, want := range map[int64]int{c.limit: http.StatusOK, c.limit + 1: http.StatusRequestEntityTooLarge} {
			rec := httptest.NewRecorder()
			read.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, c.path, bytes.NewReader(make([]byte, size))))

			if rec.Code != want {
				t.Errorf("%s with %d bytes: status %d, want %d", c.path, size, rec.Code, want)
			}
		}
	}

	app, _ := inventoryApp(t)
	h := app.Handler()

	post := func(contentType, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(body))
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Accept", "application/json")

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		return rec
	}

	pad := strings.Repeat("a", int(limits.FormBody))

	if rec := post("application/json", `{"email":"`+pad+`"}`); rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("JSON form over %d bytes: status %d, want 413", limits.FormBody, rec.Code)
	}

	// vukem v0.34.0 (VF-061) keeps the MaxBytesError from MethodOverride, so an oversized
	// urlencoded form is 413. The server-level limit in routes/web.go is still required:
	// MethodOverride parses the body before the router middleware runs.
	if rec := post("application/x-www-form-urlencoded", "email=a%40b.co&pad="+pad); rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("urlencoded form over %d bytes: status %d, want 413", limits.FormBody, rec.Code)
	}

	small := post("application/x-www-form-urlencoded", "email=a%40b.co&pad=a")
	if small.Code == http.StatusRequestEntityTooLarge {
		t.Fatal("small urlencoded form was rejected as too large")
	}
	var parsed struct{ Errors map[string]any }
	_ = json.Unmarshal(small.Body.Bytes(), &parsed)
	if _, missing := parsed.Errors["email"]; missing {
		t.Error("small urlencoded form was not parsed")
	}

	for size, tooLarge := range map[int64]bool{limits.StripeWebhookBody: false, limits.StripeWebhookBody + 1: true} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, billing.WebhookPath, bytes.NewReader(make([]byte, size))))

		if (rec.Code == http.StatusRequestEntityTooLarge) != tooLarge {
			t.Errorf("webhook with %d bytes: status %d", size, rec.Code)
		}
	}
}
