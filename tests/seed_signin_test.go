package tests

import (
	"net/url"
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/social"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func markSeed(t *testing.T, app *vukemtest.App, email string) int64 {
	t.Helper()

	db, err := app.DB()
	if err != nil {
		t.Fatal(err)
	}

	ph := db.Dialect().Placeholder

	if _, err := db.ExecContext(
		t.Context(),
		"UPDATE users SET is_seed = "+ph(1)+" WHERE email = "+ph(2),
		true,
		email,
	); err != nil {
		t.Fatal(err)
	}

	var id int64
	if err := db.QueryRowContext(t.Context(), "SELECT id FROM users WHERE email = "+ph(1), email).
		Scan(&id); err != nil {
		t.Fatal(err)
	}

	return id
}

// A seed user (users.is_seed) can never sign in, whatever the way in; every refusal is the generic one.
func TestSecurity_A07_SeedUserCannotSignIn(t *testing.T) {
	t.Parallel()

	app := vukemtest.New(t, bootstrap.New)
	const email = "seed1@example.test"

	register(t, app, app.Client(t).Inertia(""), email)

	// Remember-me session and app token exist before the account is flagged.
	remembered := app.Client(t).Inertia("")
	remembered.PostJSON("/login", map[string]any{"identifier": email, "password": testPassword, "remember": true}).
		AssertRedirect("/onboarding")
	remembered.Get("/onboarding").AssertOK()

	tok := issue(t, app.Client(t), email, "Pixel")

	// A reset link requested while the account was still normal.
	app.Client(t).Inertia("").PostJSON("/recover-password", map[string]any{"email": email}).
		AssertRedirect("/recover-password")
	reset := link(t, app, email)
	ru, _ := url.Parse(reset)
	token := strings.TrimPrefix(ru.Path, "/reset-password/")

	id := markSeed(t, app, email)

	// Password login answers like a wrong password.
	seedLogin := app.Client(t).
		Inertia("").
		PostJSON("/login", map[string]any{"identifier": email, "password": testPassword})
	wrong := app.Client(t).
		Inertia("").
		PostJSON("/login", map[string]any{"identifier": "nobody@example.test", "password": testPassword})

	if seedLogin.Status() != wrong.Status() || seedLogin.Header("Location") != wrong.Header("Location") {
		t.Errorf("seed login = %d %q, wrong password = %d %q",
			seedLogin.Status(), seedLogin.Header("Location"), wrong.Status(), wrong.Header("Location"))
	}

	// The remember-me session and the browser session end.
	remembered.Get("/onboarding").AssertRedirect("/login")

	// App tokens: issue, use and refresh all fail.
	body := map[string]any{"identifier": email, "password": testPassword, "deviceName": "x"}
	if got := codeOf(
		app.Client(t).PostJSON("/app/v1/tokens", body).AssertUnprocessable(),
	); got != "invalid-credentials" {
		t.Errorf("seed token issue code = %q", got)
	}

	app.Client(t).WithToken(tok.Token).GetJSON("/app/v1/tokens").AssertUnauthorized()
	app.Client(t).WithToken(tok.Token).PostJSON("/app/v1/tokens/rotate", map[string]any{}).AssertUnauthorized()

	// Google: neither an existing link nor a verified email match signs the seed user in.
	acc, err := accounts.New(app.App)
	if err != nil {
		t.Fatal(err)
	}

	db := acc.DB
	ph := db.Dialect().Placeholder

	if _, err := db.ExecContext(t.Context(),
		"INSERT INTO social_accounts (user_id, provider, provider_id, created_at, updated_at) VALUES ("+
			ph(1)+", 'google', 'g-seed', "+ph(2)+", "+ph(3)+")", id, acc.Clock(), acc.Clock()); err != nil {
		t.Fatal(err)
	}

	for _, su := range []social.User{
		{Provider: "google", ID: "g-seed", Email: email, EmailVerified: true},
		{Provider: "google", ID: "g-other", Email: email, EmailVerified: true},
	} {
		if _, err := acc.FindOrCreateFromSocial(t.Context(), su); err == nil {
			t.Errorf("google sign-in of a seed user succeeded: %+v", su)
		}
	}

	vukemtest.AssertDatabaseMissing(t, app, "social_accounts", map[string]any{"provider_id": "g-other"})

	// Password reset: no new link, and the link issued earlier is dead.
	before := len(app.Mails().SentTo(email))
	app.Client(t).Inertia("").PostJSON("/recover-password", map[string]any{"email": email}).
		AssertRedirect("/recover-password")

	if got := len(app.Mails().SentTo(email)); got != before {
		t.Errorf("reset mail sent to a seed user: %d -> %d", before, got)
	}

	const newPassword = "Another horse 7 battery!"

	app.Client(t).Inertia("").WithHeader("Referer", reset).PostJSON("/reset-password", map[string]any{
		"token": token, "email": email, "password": newPassword, "password_confirmation": newPassword,
	}).AssertRedirect(reset)

	u, err := acc.Users.ByCredentials(t.Context(), map[string]string{"email": email})
	if err == nil {
		t.Errorf("provider returned the seed user: %+v", u.ID)
	}
}
