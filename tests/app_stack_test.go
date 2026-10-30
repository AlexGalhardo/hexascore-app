package tests

import (
	"net/http"
	"sync"
	"testing"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/vukemtest"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/appapi"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
	"github.com/AlexGalhardo/hexascore-app/routes"
)

func TestAppMinVersion(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: map[string]string{
		"APP_MIN_SUPPORTED_VERSION": "1.2.0",
	}})
	email := "version@example.com"
	signedUp(t, app, email)
	tok := issue(t, app.Client(t).WithHeader(appapi.ClientVersionHeader, "1.2.0"), email, "Pixel")

	ok := app.Client(t).WithToken(tok.Token).WithHeader(appapi.ClientVersionHeader, "1.2.0").
		GetJSON("/app/v1/tokens").AssertOK()
	if ok.Header(appapi.MinVersionHeader) != "1.2.0" {
		t.Fatalf("min header = %q", ok.Header(appapi.MinVersionHeader))
	}

	old := app.Client(t).WithToken(tok.Token).WithHeader(appapi.ClientVersionHeader, "1.1.0")
	res := old.GetJSON("/app/v1/tokens").AssertStatus(http.StatusUpgradeRequired)
	if codeOf(res) != appapi.UpgradeCode || res.Header(appapi.MinVersionHeader) != "1.2.0" {
		t.Fatalf("upgrade = %s header %q", codeOf(res), res.Header(appapi.MinVersionHeader))
	}

	// An old build can still log out.
	old.Delete("/app/v1/tokens/current").AssertNoContent()

	if got := app.Client(t).WithHeader(appapi.ClientVersionHeader, "0.0.1").Get("/up").Status(); got == http.StatusUpgradeRequired {
		t.Fatal("/up answered 426")
	}
}

func TestAppIdempotency(t *testing.T) {
	t.Parallel()

	entered := make(chan struct{})
	release := make(chan struct{})
	var calls int
	var mu sync.Mutex

	app := vukemtest.New(t, bootstrap.New, vukemtest.Options{AppOptions: []vukem.Option{
		vukem.WithRoutes(holdRoute(entered, release, func() {
			mu.Lock()
			calls++
			mu.Unlock()
		})),
	}})
	email := "idem@example.com"
	signedUp(t, app, email)
	tok := issue(t, app.Client(t), email, "Pixel")

	body := map[string]any{"n": 1}
	send := func(key string, payload map[string]any) *vukemtest.Response {
		t.Helper()
		return app.Client(t).WithToken(tok.Token).WithHeader("Idempotency-Key", key).PostJSON("/app/v1/_idempotency/hold", payload)
	}

	done := make(chan int, 1)
	go func() {
		done <- send("flight", body).Status()
	}()

	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the in-flight write")
	}

	if got := send("flight", body).Status(); got != http.StatusConflict {
		t.Fatalf("in flight = %d, want 409", got)
	}

	close(release)

	if got := <-done; got != http.StatusNoContent {
		t.Fatalf("first write = %d, want 204", got)
	}

	replay := send("flight", body).AssertNoContent()
	if replay.Header("Idempotent-Replayed") != "true" {
		t.Fatal("replay missing Idempotent-Replayed")
	}

	send("flight", map[string]any{"n": 2}).AssertStatus(http.StatusUnprocessableEntity)

	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Fatalf("handler calls = %d, want 1", calls)
	}
}

func holdRoute(entered, release chan struct{}, hit func()) func(*router.Router, *vukem.App) error {
	var once sync.Once

	return func(r *router.Router, app *vukem.App) error {
		acc, err := accounts.New(app)
		if err != nil {
			return err
		}

		limiters, err := limits.Factory(app)
		if err != nil {
			return err
		}

		stack, err := routes.AppWriteStack(app, acc, limiters)
		if err != nil {
			return err
		}

		r.Group("/app/v1", func(g *router.Group) {
			g.Post("/_idempotency/hold", func(c *web.Ctx) error {
				hit()
				once.Do(func() { close(entered) })
				<-release
				return c.NoContent()
			}).Use(stack...)
		})

		return nil
	}
}
