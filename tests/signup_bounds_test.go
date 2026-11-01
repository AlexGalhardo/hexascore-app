//go:build integration

package tests

import (
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// newSignupApp gives each group its own app: sign-up is limited to 10
// requests per IP in 10 minutes and every group below stays under it.
func newSignupApp(t *testing.T) *vukemtest.App {
	t.Helper()

	return vukemtest.New(t, bootstrap.New,
		vukemtest.Options{Env: map[string]string{"DATABASE_URL": pgtest.Database(t, "")}})
}

// postSignup solves the proof-of-work first: the gate turns on after a few sign-ups from one IP.
func postSignup(t *testing.T, c *vukemtest.Client, form map[string]any) *vukemtest.Response {
	t.Helper()

	return c.WithHeader(middlewareHeader, solved(t, c, "signup")).PostJSON("/signup", form)
}

func signUpWithNickname(app *vukemtest.App, email, nickname string) map[string]any {
	f := signUpForm(app, "Ana Silva", email, testPassword)
	f["nickname"] = nickname

	return f
}

func TestSecurity_A07_SignupEmailAndPasswordBounds(t *testing.T) {
	app := newSignupApp(t)
	c := app.Client(t)
	ok := func(f map[string]any) { postSignup(t, c, f).AssertRedirect("/signup/check-email") }
	bad := func(f map[string]any) { postSignup(t, c, f).AssertUnprocessable() }

	// Email 12 to 48 characters ("@example.com" is 12).
	bad(signUpForm(app, "Ana Silva", "abc@exam.io", testPassword))
	ok(signUpForm(app, "Ana Silva", "abcd@exam.io", testPassword))
	ok(signUpForm(app, "Ana Silva", strings.Repeat("a", 36)+"@example.com", testPassword))
	bad(signUpForm(app, "Ana Silva", strings.Repeat("a", 37)+"@example.com", testPassword))

	// Password 8 to 32 characters.
	pad := func(n int) string { return "Aa1!" + strings.Repeat("x", n-4) }
	bad(signUpForm(app, "Ana Silva", "seven@example.com", pad(7)))
	ok(signUpForm(app, "Ana Silva", "eight@example.com", pad(8)))
	ok(signUpForm(app, "Ana Silva", "thirtytwo@example.com", pad(32)))
	bad(signUpForm(app, "Ana Silva", "thirtythree@example.com", pad(33)))
}

func TestSecurity_A07_SignupNicknameRules(t *testing.T) {
	app := newSignupApp(t)
	c := app.Client(t)

	for _, nick := range []string{"abc", strings.Repeat("a", 17), "ab cd", "ab_cd", "abc1", "ação"} {
		postSignup(t, c, signUpWithNickname(app, "x"+strings.ReplaceAll(nick, " ", "")+"@example.com", nick)).
			AssertUnprocessable().
			AssertSee("invalid-nickname")
	}

	postSignup(t, c, signUpWithNickname(app, "four@example.com", "abcd")).AssertRedirect("/signup/check-email")
	postSignup(t, c, signUpWithNickname(app, "sixteen@example.com", strings.Repeat("b", 16))).
		AssertRedirect("/signup/check-email")
}

func TestSecurity_A07_SignupNicknameStoredLowercaseAndEnumerationSafe(t *testing.T) {
	app := newSignupApp(t)
	c := app.Client(t)

	postSignup(t, c, signUpWithNickname(app, "ana@example.com", "  AnaLuiza ")).
		AssertRedirect("/signup/check-email")
	verifyEmail(t, app, "ana@example.com")
	// The nickname is stored lowercase, so the account skips onboarding and answers to it.
	app.Client(t).PostJSON("/login", map[string]any{"identifier": "analuiza", "password": testPassword}).
		AssertRedirect("/dashboard")

	// A taken nickname answers the same for a new and an existing email: no account enumeration.
	postSignup(t, c, signUpWithNickname(app, "new@example.com", "analuiza")).
		AssertUnprocessable().
		AssertSee("nickname-taken")
	postSignup(t, c, signUpWithNickname(app, "ana@example.com", "analuiza")).
		AssertUnprocessable().
		AssertSee("nickname-taken")

	// An existing email with a free nickname gets the normal answer and the nickname is not stored.
	postSignup(t, c, signUpWithNickname(app, "ana@example.com", "otherone")).
		AssertRedirect("/signup/check-email")
	postSignup(t, c, signUpWithNickname(app, "bia@example.com", "otherone")).
		AssertRedirect("/signup/check-email")

	// No nickname (no-JS clients, older forms): the account picks one at onboarding.
	postSignup(t, c, signUpForm(app, "Carla", "carla@example.com", testPassword)).
		AssertRedirect("/signup/check-email")
	verifyEmail(t, app, "carla@example.com")
	app.Client(t).PostJSON("/login", map[string]any{"identifier": "carla@example.com", "password": testPassword}).
		AssertRedirect("/onboarding")
}
