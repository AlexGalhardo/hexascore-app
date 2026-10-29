package tests

import (
	"strings"
	"testing"

	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

// TestInertiaPages checks that every Auth page renders its component in each
// locale and receives that locale's catalogue.
func TestInertiaPages(t *testing.T) {
	locales := []struct{ prefix, locale string }{
		{"", "en"},
		{"/pt-br", "pt-BR"},
		{"/es", "es"},
	}

	for _, l := range locales {
		t.Run("Home/"+l.locale, func(t *testing.T) {
			app := vukemtest.New(t, bootstrap.New)
			guest := app.Client(t).Inertia("")
			guest.Get(l.prefix+"/").AssertOK().AssertInertia("Home/Index").PropEquals("locale", l.locale).
				HasProp("sections", 4)

			signed := app.Client(t).Inertia("")
			signed.PostJSON("/signup", signUpForm("Bia Souza", "bia@example.com", testPassword))
			signed.PostJSON("/login", map[string]any{"identifier": "bia@example.com", "password": testPassword})
			signed.Get(l.prefix+"/").AssertOK().AssertInertia("Home/Index").HasProp("sections", 4)
		})

		t.Run("Auth/"+l.locale, func(t *testing.T) {
			app := vukemtest.New(t, bootstrap.New)
			guest := app.Client(t).Inertia("")

			for path, component := range map[string]string{
				"/login":              "Auth/Login",
				"/signup":             "Auth/Signup",
				"/signup/check-email": "Auth/CheckEmail",
				"/recover-password":   "Auth/Recover",
				"/reset-password/tok": "Auth/Reset",
			} {
				guest.Get(l.prefix+path).AssertOK().AssertInertia(component).PropEquals("locale", l.locale).
					HasProp("messages")
			}

			signed := app.Client(t).Inertia("")
			signed.PostJSON("/signup", signUpForm("Ana Silva", "ana@example.com", testPassword))
			signed.PostJSON("/login", map[string]any{"identifier": "ana@example.com", "password": testPassword})
			signed.Get(l.prefix+"/verify-email").AssertOK().AssertInertia("Auth/VerifyNotice").
				PropEquals("locale", l.locale)

			verifyEmail(t, app, "ana@example.com")
			signed.Get(l.prefix+"/onboarding").AssertOK().AssertInertia("Auth/Onboarding").
				PropEquals("locale", l.locale)

			email := "twofa-" + strings.ToLower(l.locale) + "@example.com"
			owner := app.Client(t).Inertia("")
			enableTwoFactor(t, app, owner, email)
			owner.PostJSON("/logout", nil).AssertRedirect("/")

			pending := app.Client(t).Inertia("")
			passwordStep(pending, email, false).AssertRedirect("/two-factor-challenge")
			pending.Get(l.prefix+"/two-factor-challenge").AssertOK().AssertInertia("Auth/TwoFactorChallenge").
				PropEquals("locale", l.locale).HasProp("messages")
		})
	}
}
