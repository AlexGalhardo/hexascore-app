//go:build integration

package tests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

type keyRig struct {
	t     *testing.T
	app   *vukemtest.App
	acc   *accounts.Accounts
	guard *api.Guard
}

func newKeyRig(t *testing.T) *keyRig {
	t.Helper()

	app := vukemtest.New(t, bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}})

	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	return &keyRig{t: t, app: app, acc: acc, guard: &api.Guard{
		Keys: &api.Keys{DB: acc.DB, Now: app.Now}, User: acc.Users.ByID,
	}}
}

// signIn registers a verified user and returns its Inertia client.
func (r *keyRig) signIn(email string) *vukemtest.Client {
	r.t.Helper()

	c := r.app.Client(r.t).Inertia("")
	register(r.t, r.app, c, email)

	return c
}

// create makes a key through the page and returns the plain key from the one-time flash.
func (r *keyRig) create(c *vukemtest.Client, name string) string {
	r.t.Helper()

	c.PostJSON("/profile/api/keys", map[string]any{"name": name}).AssertRedirect("/profile/api")

	page := c.Get("/profile/api").AssertOK().AssertInertia("Developer/ApiKeys")
	flash, _ := page.Page()["flash"].(map[string]any)

	plain, _ := flash["newKey"].(string)
	if !api.WellFormed(plain) {
		r.t.Fatalf("flash newKey = %q, not a key", plain)
	}

	return plain
}

// authenticate runs the /api/v1 guard on a request carrying bearer.
func (r *keyRig) authenticate(bearer string) error {
	r.t.Helper()

	req := httptest.NewRequestWithContext(r.t.Context(), http.MethodGet, "/api/v1/games", nil)
	req.Header.Set("Authorization", "Bearer "+bearer)

	return r.guard.Authenticate(web.NewCtx(httptest.NewRecorder(), req, nil))
}

func (r *keyRig) count(q string, args ...any) int {
	r.t.Helper()

	var n int
	if err := r.acc.DB.QueryRowContext(r.t.Context(), q, args...).Scan(&n); err != nil {
		r.t.Fatal(err)
	}

	return n
}

func TestKeysWebFlow(t *testing.T) {
	r := newKeyRig(t)
	c := r.signIn("keys@example.com")

	plain := r.create(c, "production server")

	// Shown once: the next visit no longer carries it.
	c.Get("/profile/api").AssertInertia("Developer/ApiKeys").MissingFlash("newKey").
		HasProp("keys", 1).PropEquals("keys.0.name", "production server").
		PropEquals("keys.0.prefix", plain[:api.DisplayPrefixLen]).PropEquals("maxKeys", api.MaxLiveKeys)

	if err := r.authenticate(plain); err != nil {
		t.Fatalf("live key refused: %v", err)
	}

	var id string
	if err := r.acc.DB.QueryRowContext(t.Context(), `SELECT id FROM api_keys`).Scan(&id); err != nil {
		t.Fatal(err)
	}

	// Another user cannot revoke it, and it keeps working.
	r.signIn("other@example.com").Delete("/profile/api/keys/" + id).AssertNotFound()

	if err := r.authenticate(plain); err != nil {
		t.Fatalf("key refused after a foreign revoke: %v", err)
	}

	c.Delete("/profile/api/keys/" + id).AssertRedirect("/profile/api")

	if err := r.authenticate(plain); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("revoked key: err = %v, want ErrUnauthenticated", err)
	}

	c.Get("/profile/api").AssertInertia("Developer/ApiKeys").HasProp("keys.0.revokedAt")
	c.Delete("/profile/api/keys/" + id).AssertNotFound()

	r.app.Client(t).Inertia("").Get("/profile/api").AssertRedirect("/login")
}

func TestKeysNameAndCap(t *testing.T) {
	r := newKeyRig(t)
	c := r.signIn("cap@example.com").WithHeader("Referer", "/profile/api")

	c.PostJSON("/profile/api/keys", map[string]any{"name": "   "}).AssertRedirect("/profile/api")
	c.Get("/profile/api").AssertInertia("Developer/ApiKeys").HasProp("errors.name").HasProp("keys", 0)

	for i := range api.MaxLiveKeys {
		r.create(c, "key "+strconv.Itoa(i))
	}

	c.PostJSON("/profile/api/keys", map[string]any{"name": "one too many"}).AssertRedirect("/profile/api")
	c.Get("/profile/api").AssertInertia("Developer/ApiKeys").PropEquals("errors.name", "limit").
		HasProp("keys", api.MaxLiveKeys)
}

func TestSecurity_A07_ApiKeyHashedAtRest(t *testing.T) {
	r := newKeyRig(t)
	plain := r.create(r.signIn("hash@example.com"), "ci")

	if n := r.count(`SELECT count(*) FROM api_keys WHERE hash = $1 AND prefix = $2`,
		api.HashKey(plain), plain[:api.DisplayPrefixLen]); n != 1 {
		t.Fatalf("%d rows with sha256(key) and the display prefix, want 1", n)
	}

	if n := r.count(`SELECT count(*) FROM api_keys WHERE position($1 in hash || name || prefix) > 0`, plain); n != 0 {
		t.Fatal("plain key stored")
	}

	for _, bad := range []string{"", "hx_live_short", plain[:len(plain)-1] + "-", "hx_live_" + "0123456789abcdefABCDEF0123456789"} {
		if err := r.authenticate(bad); !errors.Is(err, auth.ErrUnauthenticated) {
			t.Errorf("authenticate(%q) = %v, want ErrUnauthenticated", bad, err)
		}
	}
}

func TestSecurity_A01_CrossSurfaceCredentials(t *testing.T) {
	r := newKeyRig(t)
	plain := r.create(r.signIn("cross@example.com"), "ci")
	app := issue(t, r.app.Client(t), "cross@example.com", "Pixel")

	// An app token on /api/v1 is refused by the key guard.
	if err := r.authenticate(app.Token); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("app token on /api/v1: err = %v, want ErrUnauthenticated", err)
	}

	// An API key on /app/v1 is refused by the token guard, while the app token works there.
	r.app.Client(t).WithToken(plain).GetJSON("/app/v1/tokens").AssertUnauthorized()
	r.app.Client(t).WithToken(app.Token).GetJSON("/app/v1/tokens").AssertOK()
}
