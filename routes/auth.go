package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/controllers/account"
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
	limiters, err := app.RateLimiters()
	if err != nil {
		return err
	}

	h := account.New(acc)
	guest := auth.Guest(acc.Web)
	required := auth.Required(acc.Web)
	// Signed-in pages also need a verified email; unverified users are sent
	// to the verification notice.
	verified := auth.Verified(acc.Web)
	loginIP, loginAccount := auth.ThrottleKeys("email")
	throttle := ratelimit.Middleware(limiters("login", 5, time.Minute), loginIP)
	perAccount := ratelimit.Middleware(limiters("login.account", 20, 10*time.Minute), loginAccount)
	mails := ratelimit.Middleware(limiters("mails", 3, time.Minute), ratelimit.ByIP)
	// Sign-up queues a mail: limit it per IP and per address. The limits count
	// requests, not accounts, so they answer alike for new and existing emails.
	signupIP := ratelimit.Middleware(limiters("register", 10, 10*time.Minute), ratelimit.ByIP)
	signupEmail := ratelimit.Middleware(limiters("register.email", 5, time.Hour), account.RegisterThrottleKey)

	r.Use(shareAuth(acc))

	r.Get("/login", h.ShowLogin).Name("login").Use(guest)
	r.Post("/login", h.Login).Use(guest, throttle, perAccount)
	r.Get("/signup", h.ShowRegister).Name("register").Use(guest)
	r.Post("/signup", h.Register).Use(guest, signupIP, signupEmail)
	r.Get("/signup/check-email", h.CheckEmail).Name("register.check").Use(guest)
	r.Post("/logout", h.Logout).Name("logout").Use(required)

	r.Get("/recover-password", h.ShowForgot).Name("password.request").Use(guest)
	r.Post("/recover-password", h.SendResetLink).Name("password.email").Use(guest, mails)
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
		Use(required, ratelimit.Middleware(limiters("deletion", 5, time.Hour), deleteKey))
	r.Delete("/profile/deletion", h.CancelDeletion).Name("deletion.cancel").Use(required)
	r.Get("/help/delete-account", h.HelpDeleteAccount).Name("help.delete-account")

	r.Get("/dashboard", h.Dashboard).Name("dashboard").Use(required, verified)

	return nil
}

// shareAuth shares the signed-in user (or null) with every page as
// auth.user; it is loaded only when a page is rendered.
func shareAuth(acc *accounts.Accounts) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := web.NewCtx(w, r, web.ServicesFrom(r.Context()))

			inertia.Share(r.Context(), "auth", func(context.Context) (any, error) {
				u, err := acc.Web.User(c)
				if err != nil {
					return map[string]any{"user": nil}, nil //nolint:nilerr // signed out
				}

				return map[string]any{"user": u}, nil
			})

			next.ServeHTTP(w, r)
		})
	}
}
