// Package account holds the login, registration, password reset and
// email verification handlers (vukem install:auth), rendered with Inertia
// pages in resources/js/pages/Auth.
package account

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/forms"
	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
)

// Handlers serves the account pages.
type Handlers struct {
	a    *accounts.Accounts
	form forms.Token
}

// New returns the handlers; form signs the sign-up page's render time.
func New(a *accounts.Accounts, form forms.Token) *Handlers { return &Handlers{a: a, form: form} }

// RegisterThrottleKey limits sign-ups per address. It keys on a hash of
// the normalized email, so the limiter store holds no addresses.
func RegisterThrottleKey(r *http.Request) string {
	if r.Method != http.MethodPost {
		return ""
	}

	email := loginEmail(r)
	if email == "" {
		// Unreadable email: a bucket per IP, never one shared by everybody.
		return "register-unparsed:" + ratelimit.ByIP(r)
	}

	sum := sha256.Sum256([]byte(email))

	return "register:" + hex.EncodeToString(sum[:])
}

// loginEmail is the normalized email of the request, shared with the login throttle.
func loginEmail(r *http.Request) string { return auth.RequestEmail(r, "email") }

// ShowLogin renders the login page.
func (*Handlers) ShowLogin(c *web.Ctx) error { return inertia.Render(c, "Auth/Login", nil) }

// Login signs the user in and goes to the page they wanted.
func (h *Handlers) Login(c *web.Ctx) error {
	in, err := web.Parse(c, requests.LoginSchema)
	if err != nil {
		return err
	}

	email, err := h.a.EmailForIdentifier(c.Context(), in.Identifier)
	if err != nil {
		return err
	}

	creds := map[string]string{"email": email, "password": in.Password}

	u, err := h.a.Web.Validate(c, creds)
	if err != nil {
		return err
	}

	// Checked after the password, so a wrong guess reveals nothing (SDD 03 section 7.1).
	if u.IsBanned(h.a.Clock()) {
		return middleware.BannedProblem(c, h.a.ProblemBase, u)
	}

	// Users with two-factor authentication enter a code first (vukem install:auth --2fa).
	if pending, err := h.startTwoFactor(c, u, in.Remember); pending || err != nil {
		return err
	}

	if err := h.a.Web.Login(c, u, auth.Remember(in.Remember)); err != nil {
		return err
	}

	return h.a.RedirectAfterLogin(c, u)
}

// ShowRegister renders the registration page.
func (h *Handlers) ShowRegister(c *web.Ctx) error {
	return inertia.Render(c, "Auth/Signup", inertia.Props{"formToken": h.form.Issue("signup")})
}

// Register answers the same for new and existing emails (no account
// enumeration) and does not log in: the user is asked to check the inbox,
// where the verification link (or an "account exists" notice) waits.
func (h *Handlers) Register(c *web.Ctx) error {
	in, err := web.Parse(c, requests.RegisterSchema)
	if err != nil {
		return err
	}

	// A filled honeypot or a form sent faster than people fill it (or without the signed render time)
	// gets the normal answer and nothing is stored (SDD 13 §7.2).
	if in.Website != "" || !h.form.Valid("signup", in.FormToken) {
		web.SecurityEvent(c.Request(), "security.honeypot", "form", "signup")

		return c.Redirect("/signup/check-email")
	}

	if _, err := h.a.Register(c.Context(), in); err != nil {
		return err
	}

	return c.Redirect("/signup/check-email")
}

// CheckEmail is the page shown after every sign-up attempt.
func (*Handlers) CheckEmail(c *web.Ctx) error { return inertia.Render(c, "Auth/CheckEmail", nil) }

// Logout signs the user out and rotates the history encryption key.
func (h *Handlers) Logout(c *web.Ctx) error {
	if err := h.a.Web.Logout(c); err != nil {
		return err
	}

	inertia.ClearHistory(c)

	return c.Redirect("/")
}

// Dashboard is the first page after login; the user arrives as the shared
// auth.user prop.
func (*Handlers) Dashboard(c *web.Ctx) error { return inertia.Render(c, "Dashboard", nil) }
