package tests

import (
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
			signed.PostJSON("/login", map[string]any{"email": "ana@example.com", "password": testPassword})
			signed.Get(l.prefix+"/verify-email").AssertOK().AssertInertia("Auth/VerifyNotice").
				PropEquals("locale", l.locale)

			verifyEmail(t, app, "ana@example.com")
			signed.Get(l.prefix+"/onboarding").AssertOK().AssertInertia("Auth/Onboarding").
				PropEquals("locale", l.locale)
		})
	}
}
