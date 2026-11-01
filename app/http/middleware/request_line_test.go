package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	vmw "github.com/AlexGalhardo/vukem/middleware"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/logging"
)

var logKey = bytes.Repeat([]byte{7}, 32)

// stack mirrors production: vukem's RequestID and Logger, then the route
// attribute the vukem router adds on each matched route.
func stack(t *testing.T) (http.Handler, *slog.Logger, *bytes.Buffer) {
	t.Helper()

	var buf bytes.Buffer

	log := slog.New(logging.NewHandler(slog.NewJSONHandler(&buf, nil), logKey))
	mux := http.NewServeMux()
	route := func(pattern string, h http.HandlerFunc) {
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			web.AddLogAttrs(r.Context(), slog.String("route", r.Pattern))
			h(w, r)
		})
	}

	route("GET /items/{slug}", func(_ http.ResponseWriter, r *http.Request) {
		logging.SetUser(r.Context(), "42")
	})
	route("POST /login", func(w http.ResponseWriter, r *http.Request) {
		log.WarnContext(r.Context(), "login failed for ana@example.com",
			"email", "ana@example.com", "password", "hunter2",
			slog.Group("headers", "Authorization", "Bearer abc.def.ghi"),
			"detail", "key hx_live_aZ09aZ09aZ09aZ09aZ09aZ09aZ09aZ09 rejected",
			"err", errStr("GET /search?q=my+secret+title&page=2 failed"))
		w.WriteHeader(http.StatusUnauthorized)
	})
	route("GET /search", func(_ http.ResponseWriter, r *http.Request) {
		log.InfoContext(r.Context(), "search", "q", r.URL.Query().Get("q"))
	})
	route("GET /up", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("fail") {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	return vmw.RequestID()(vmw.Logger(log)(mux)), log, &buf
}

type errStr string

func (e errStr) Error() string { return string(e) }

func do(h http.Handler, method, target, remote string) {
	r := httptest.NewRequest(method, target, nil)
	r.RemoteAddr = remote
	h.ServeHTTP(httptest.NewRecorder(), r)
}

func lines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	var out []map[string]any

	for l := range strings.SplitSeq(strings.TrimSpace(buf.String()), "\n") {
		if l == "" {
			continue
		}

		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("not JSON: %s", l)
		}

		out = append(out, m)
	}

	return out
}

func TestLog_RequestLineShape(t *testing.T) {
	t.Parallel()

	h, _, buf := stack(t)

	do(h, http.MethodGet, "/items/zelda", "203.0.113.77:5000")
	do(h, http.MethodGet, "/nope", "[2001:db8:abcd:12::1]:443")

	got := lines(t, buf)
	if len(got) != 2 {
		t.Fatalf("lines = %d, want 2:\n%s", len(got), buf)
	}

	item := got[0]
	for _, k := range []string{"level", "request_id", "route", "method", "status", "duration_ms", "user_hash"} {
		if _, ok := item[k]; !ok {
			t.Errorf("request line lacks %q: %v", k, item)
		}
	}

	checks := []struct{ key, want string }{
		{"route", "GET /items/{slug}"},
		{"user_hash", logging.UserHash(logKey, "42")},
		{"ip", "203.0.113.0"},
		{"peer", "203.0.113.0"},
	}
	for _, c := range checks {
		if item[c.key] != c.want {
			t.Errorf("%s = %v, want %s", c.key, item[c.key], c.want)
		}
	}

	if len(logging.UserHash(logKey, "42")) != 16 {
		t.Error("user_hash is not 16 hex chars")
	}

	if _, ok := item["duration"]; ok {
		t.Error("raw duration kept next to duration_ms")
	}

	miss := got[1]
	if miss["route"] != "unmatched" || miss["ip"] != "2001:db8:abcd::" || miss["user_hash"] != nil {
		t.Errorf("unmatched IPv6 line = %v", miss)
	}
}

func TestLog_SamplingKeepsErrorsAndOneInHundred(t *testing.T) {
	t.Parallel()

	h, _, buf := stack(t)

	for range 200 {
		do(h, http.MethodGet, "/up", "10.0.0.1:1")
	}

	do(h, http.MethodGet, "/up?fail=1", "10.0.0.1:1")

	got := lines(t, buf)
	if len(got) != 3 {
		t.Fatalf("lines = %d, want 2 sampled successes + 1 failure", len(got))
	}

	if got[2]["status"] != float64(http.StatusServiceUnavailable) {
		t.Errorf("last line = %v, want the 503", got[2])
	}
}

func TestSecurity_A09_HexaLogsRedacted(t *testing.T) {
	t.Parallel()

	h, _, buf := stack(t)

	r := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("email=ana%40example.com&password=hunter2"))
	r.Header.Set("Authorization", "Bearer abc.def.ghi")
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(httptest.NewRecorder(), r)
	do(h, http.MethodGet, "/search?q=my+secret+title", "198.51.100.9:1")

	out := buf.String()
	for _, leak := range []string{"hunter2", "ana@example.com", "abc.def.ghi", "aZ09aZ09", "my secret title", "my+secret+title", "198.51.100.9"} {
		if strings.Contains(out, leak) {
			t.Errorf("log contains %q:\n%s", leak, out)
		}
	}

	if !strings.Contains(out, logging.Redacted) {
		t.Errorf("nothing marked redacted:\n%s", out)
	}
}
