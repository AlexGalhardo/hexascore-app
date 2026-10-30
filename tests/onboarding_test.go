//go:build integration

package tests

import (
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func TestOnboardingNickname(t *testing.T) {
	app := vukemtest.New(
		t,
		bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}},
	)

	// Unverified: sent to the verification notice.
	ana := app.Client(t)
	register(t, app, ana, "ana@example.com")

	raw := app.Client(t)
	raw.PostJSON("/signup", signUpForm(app, "Bia Souza", "bia@example.com", testPassword))
	raw.PostJSON("/login", map[string]any{"identifier": "bia@example.com", "password": testPassword})
	raw.Get("/onboarding").AssertRedirect("/verify-email")
	raw.PostJSON("/onboarding", map[string]any{"nickname": "bianca"}).AssertForbidden()

	ana.Get("/onboarding").AssertOK()
	ana.PostJSON("/onboarding", map[string]any{"nickname": "ab1"}).AssertUnprocessable()
	ana.PostJSON("/onboarding", map[string]any{"nickname": "  MariaLuiza "}).AssertRedirect("/dashboard")
	// The nickname is kept lowercase and the page is no longer offered.
	ana.Get("/onboarding").AssertRedirect("/dashboard")

	other := app.Client(t)
	register(t, app, other, "carla@example.com")
	other.PostJSON("/onboarding", map[string]any{"nickname": "marialuiza"}).
		AssertUnprocessable().
		AssertSee("nickname-taken")
}
