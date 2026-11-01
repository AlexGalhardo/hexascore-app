package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

type memSession struct{ data map[string]any }

func (s *memSession) Get(k string) any    { return s.data[k] }
func (s *memSession) Put(k string, v any) { s.data[k] = v }
func (s *memSession) Forget(ks ...string) {
	for _, k := range ks {
		delete(s.data, k)
	}
}
func (s *memSession) Pull(k string) any     { v := s.data[k]; delete(s.data, k); return v }
func (s *memSession) Flash(k string, v any) { s.data[k] = v }
func (s *memSession) Regenerate()           {}
func (s *memSession) Invalidate()           { s.data = map[string]any{} }
func (s *memSession) ID() string            { return "test" }

// guardEnv is one browser: a session, a signed-in user (nil = guest) and a
// clock the test moves.
type guardEnv struct {
	sess      *memSession
	user      *models.User
	twoFactor bool
	now       time.Time
	loggedOut bool
	h         http.Handler
}

func newGuard(t *testing.T, allowlist string) *guardEnv {
	t.Helper()

	list, err := middleware.ParseAllowlist(allowlist)
	if err != nil {
		t.Fatal(err)
	}

	e := &guardEnv{sess: &memSession{data: map[string]any{}}, now: time.Date(2026, 10, 29, 9, 0, 0, 0, time.UTC)}
	e.h = middleware.AdminGuard(middleware.AdminGuardConfig{
		User: func(*web.Ctx) (models.User, error) {
			if e.user == nil {
				return models.User{}, errors.New("guest")
			}

			return *e.user, nil
		},
		TwoFactorEnabled: func(context.Context, int64) (bool, error) { return e.twoFactor, nil },
		Logout: func(*web.Ctx) error {
			e.loggedOut, e.user = true, nil
			e.sess.Invalidate()

			return nil
		},
		Allowlist: list,
		Now:       func() time.Time { return e.now },
		SetupPath: "/user/two-factor",
		LoginPath: "/login",
	})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))

	return e
}

func (e *guardEnv) get(path, remote string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, path, nil)
	r.RemoteAddr = remote
	r = r.WithContext(web.WithSession(r.Context(), e.sess))

	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, r)

	return rec
}

func admin() *models.User { return &models.User{ID: 7, Role: models.RoleAdmin} }

func TestSecurity_A01_HexaAdminNeeds2FA(t *testing.T) {
	t.Parallel()

	e := newGuard(t, "")
	e.user = admin()

	if rec := e.get("/admin/users", "203.0.113.5:1"); rec.Code != http.StatusSeeOther ||
		rec.Header().Get("Location") != "/user/two-factor" {
		t.Fatalf("admin without 2FA on /admin: %d %q, want 303 to setup", rec.Code, rec.Header().Get("Location"))
	}

	// The setup page itself and the rest of the site stay reachable.
	if rec := e.get("/user/two-factor", "203.0.113.5:1"); rec.Code != http.StatusOK {
		t.Fatalf("setup page: %d", rec.Code)
	}

	e.twoFactor = true
	if rec := e.get("/admin", "203.0.113.5:1"); rec.Code != http.StatusOK {
		t.Fatalf("admin with 2FA: %d, want 200", rec.Code)
	}

	// Non-admins pass through; the panel gate answers them 404.
	e2 := newGuard(t, "")
	e2.user = &models.User{ID: 8, Role: "user"}
	if rec := e2.get("/admin", "203.0.113.5:1"); rec.Code != http.StatusOK {
		t.Fatalf("non-admin: %d, want pass-through", rec.Code)
	}
}

func TestSecurity_A07_ModeratorRequires2FA(t *testing.T) {
	t.Parallel()

	e := newGuard(t, "")
	e.user = &models.User{ID: 9, Role: models.RoleModerator}

	if rec := e.get("/admin/moderation", "203.0.113.5:1"); rec.Code != http.StatusSeeOther ||
		rec.Header().Get("Location") != "/user/two-factor" {
		t.Fatalf("moderator without 2FA: %d %q, want 303 to setup", rec.Code, rec.Header().Get("Location"))
	}

	e.twoFactor = true
	if rec := e.get("/admin/moderation", "203.0.113.5:1"); rec.Code != http.StatusOK {
		t.Fatalf("moderator with 2FA: %d, want 200", rec.Code)
	}
}

func TestSecurity_A07_HexaAdminSessionExpiry(t *testing.T) {
	t.Parallel()

	t.Run("idle 31 min logs out", func(t *testing.T) {
		t.Parallel()

		e := newGuard(t, "")
		e.user, e.twoFactor = admin(), true

		if rec := e.get("/admin", "203.0.113.5:1"); rec.Code != http.StatusOK {
			t.Fatalf("first request: %d", rec.Code)
		}

		e.now = e.now.Add(29 * time.Minute)
		if rec := e.get("/", "203.0.113.5:1"); rec.Code != http.StatusOK {
			t.Fatalf("29 min idle: %d, want 200", rec.Code)
		}

		e.now = e.now.Add(31 * time.Minute)

		rec := e.get("/admin", "203.0.113.5:1")
		if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/login" || !e.loggedOut {
			t.Fatalf("31 min idle: %d %q loggedOut=%v, want 303 /login and logout",
				rec.Code, rec.Header().Get("Location"), e.loggedOut)
		}
	})

	t.Run("absolute 12 h logs out even when active", func(t *testing.T) {
		t.Parallel()

		e := newGuard(t, "")
		e.user, e.twoFactor = admin(), true

		for range 12 * 3 {
			if rec := e.get("/admin", "203.0.113.5:1"); rec.Code != http.StatusOK {
				t.Fatalf("active at %s: %d", e.now, rec.Code)
			}

			e.now = e.now.Add(20 * time.Minute)
		}

		e.now = e.now.Add(time.Minute)

		if rec := e.get("/admin", "203.0.113.5:1"); rec.Code != http.StatusSeeOther || !e.loggedOut {
			t.Fatalf("after 12 h: %d loggedOut=%v, want 303 and logout", rec.Code, e.loggedOut)
		}
	})

	t.Run("another account's clock in the session does not count", func(t *testing.T) {
		t.Parallel()

		e := newGuard(t, "")
		e.sess.data["admin.uid"], e.sess.data["admin.since"], e.sess.data["admin.seen"] = "99", int64(0), int64(0)
		e.user, e.twoFactor = admin(), true

		if rec := e.get("/admin", "203.0.113.5:1"); rec.Code != http.StatusOK {
			t.Fatalf("fresh admin with a stale foreign clock: %d, want 200", rec.Code)
		}
	})
}

func TestSecurity_A01_HexaAdminIPAllowlist(t *testing.T) {
	t.Parallel()

	e := newGuard(t, "198.51.100.0/24, 2001:db8::/32, 192.0.2.7")
	e.user, e.twoFactor = admin(), true

	cases := []struct {
		path, remote string
		want         int
	}{
		{"/admin", "198.51.100.20:1", http.StatusOK},
		{"/admin/users", "[2001:db8::1]:1", http.StatusOK},
		{"/admin", "192.0.2.7:1", http.StatusOK},
		{"/admin", "203.0.113.5:1", http.StatusNotFound},
		{"/admin/users", "192.0.2.8:1", http.StatusNotFound},
		{"/", "203.0.113.5:1", http.StatusOK},
		{"/administrator", "203.0.113.5:1", http.StatusOK},
	}
	for _, c := range cases {
		if rec := e.get(c.path, c.remote); rec.Code != c.want {
			t.Errorf("%s from %s: %d, want %d", c.path, c.remote, rec.Code, c.want)
		}
	}

	// A guest outside the allowlist also gets 404, not the login redirect.
	e.user = nil
	if rec := e.get("/admin", "203.0.113.5:1"); rec.Code != http.StatusNotFound {
		t.Errorf("guest outside allowlist: %d, want 404", rec.Code)
	}

	if _, err := middleware.ParseAllowlist("10.0.0.0/8, nonsense"); err == nil {
		t.Error("a bad allowlist entry was accepted")
	}

	if list, err := middleware.ParseAllowlist(" , "); err != nil || len(list) != 0 {
		t.Errorf("blank allowlist = %v, %v; want off", list, err)
	}
}
