package routes

import (
	"net/http"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/controllers/account"
	"github.com/AlexGalhardo/hexascore-app/app/forms"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Auth registers login, registration, password reset and email
// verification routes (vukem install:auth).
func Auth(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	// RATE_LIMIT_STORE=cache shares the limits between servers (CACHE_STORE=database or Redis).
	limiters, err := limits.Factory(app)
	if err != nil {
		return err
	}

	pow, err := newChallenger(app, acc.ProblemBase)
	if err != nil {
		return err
	}

	formKey, err := app.DeriveKey("signup.form", 32)
	if err != nil {
		return err
	}

	h := account.New(acc, forms.Token{Key: formKey, Now: app.Now})
	guest := auth.Guest(acc.Web)
	required := auth.Required(acc.Web)
	// Signed-in pages also need a verified email; unverified users are sent
	// to the verification notice.
	verified := auth.Verified(acc.Web)
	loginIP, loginAccount := auth.ThrottleKeys("identifier")
	throttle := ratelimit.Middleware(limits.Login.From(limiters), loginIP)
	perAccount := ratelimit.Middleware(limits.LoginAccount.From(limiters), loginAccount)
	// Caps one client across every identifier (email, nickname, case variants)
	// without resolving accounts before authentication: a resolved key would
	// leak which nickname belongs to which email and let a public nickname
	// lock out the email login (security review, P1.C.2.1).
	anyIdentifier := ratelimit.Middleware(limits.LoginIP.From(limiters), ratelimit.ByIP)
	mails := ratelimit.Middleware(limits.Mails.From(limiters), ratelimit.ByIP)
	// Sign-up queues a mail: limit it per IP and per address. The limits count
	// requests, not accounts, so they answer alike for new and existing emails.
	signupIP := ratelimit.Middleware(limits.Signup.From(limiters), ratelimit.ByIP)
	signupEmail := ratelimit.Middleware(limits.SignupEmail.From(limiters), account.RegisterThrottleKey)

	r.Get("/login", h.ShowLogin).Name("login").Use(guest)
	r.Post("/login", h.Login).Use(guest, anyIdentifier, throttle, perAccount, pow.Require("login"))
	r.Get("/signup", h.ShowRegister).Name("register").Use(guest)
	r.Post("/signup", h.Register).Use(guest, signupIP, signupEmail, pow.Require("signup"))
	r.Get("/signup/check-email", h.CheckEmail).Name("register.check").Use(guest)
	r.Post("/logout", h.Logout).Name("logout").Use(required)

	r.Get("/recover-password", h.ShowForgot).Name("password.request").Use(guest)
	// RegisterThrottleKey hashes the posted email, which is also the reset form's field.
	forgotIP := ratelimit.Middleware(limits.Forgot.From(limiters), ratelimit.ByIP)
	forgotEmail := ratelimit.Middleware(limits.ForgotEmail.From(limiters), account.RegisterThrottleKey)
	r.Post("/recover-password", h.SendResetLink).Name("password.email").
		Use(guest, forgotIP, forgotEmail, pow.Require("reset"))
	r.Get("/reset-password/{token}", h.ShowReset).Name("password.reset").Use(guest)
	r.Post("/reset-password", h.Reset).Name("password.update").Use(guest)

	r.Get("/verify-email", h.Notice).Name("verification.notice").Use(required)
	// The signed link needs the account's login: signed out, Required keeps
	// the link as the intended URL and the login comes back to it. Verify
	// checks the signature after the login, so an expired link can point to
	// the resend page.
	r.Get("/verify-email/confirm/{id}", h.Verify).Name("verification.verify").Use(required)
	r.Post("/verify-email/resend", h.Resend).Name("verification.send").Use(required, mails)

	r.Get("/onboarding", h.ShowOnboarding).Name("onboarding").Use(required, verified)
	r.Post("/onboarding", h.Onboarding).Use(required, verified)

	// 5 schedules per hour per user (SDD 03 section 4); cancelling is not limited.
	deleteKey := func(r *http.Request) string {
		u, err := auth.UserFrom[models.User](r.Context(), "web")
		if err != nil {
			return ratelimit.ByIP(r)
		}

		return "deletion:" + u.AuthID()
	}
	r.Post("/profile/deletion", h.ScheduleDeletion).Name("deletion.schedule").
		Use(required, ratelimit.Middleware(limits.Deletion.From(limiters), deleteKey))
	r.Delete("/profile/deletion", h.CancelDeletion).Name("deletion.cancel").Use(required)
	r.Get("/help/delete-account", h.HelpDeleteAccount).Name("help.delete-account")

	r.Get("/dashboard", h.Dashboard).Name("dashboard").Use(required, verified)

	return nil
}
