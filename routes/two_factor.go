package routes

import (
	"context"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/controllers/account"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	"github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// TwoFactor registers the two-factor settings page and the login
// challenge (vukem install:auth --2fa).
func TwoFactor(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	// RATE_LIMIT_STORE=cache shares the limits between servers (CACHE_STORE=database or Redis).
	limiters, err := limits.Factory(app)
	if err != nil {
		return err
	}

	h := account.NewTwoFactor(acc, app.Config.Name)
	required := auth.Required(acc.Web)
	verified := auth.Verified(acc.Web)
	// Google-only accounts must choose a password first (SDD 03 A07).
	hasPassword := middleware.PasswordSet(
		func(c *web.Ctx) (models.User, error) { return acc.Web.User(c) },
		acc.PasswordSet,
		func(ctx context.Context) string { return acc.T(ctx, "auth.two_factor.password_required") },
	)
	// 5 codes per minute per user and IP: 6-digit codes cannot be guessed.
	// The password checks of the settings page (disable, new recovery
	// codes) share the confirm bucket.
	challenge := ratelimit.Middleware(limits.TwoFactor.From(limiters), account.TwoFactorThrottleKey)
	confirm := ratelimit.Middleware(limits.TwoFactorSet.From(limiters), account.TwoFactorThrottleKey)

	r.Get("/user/two-factor", h.Show).Name("two-factor.show").Use(required, verified, hasPassword)
	r.Post("/user/two-factor", h.Enable).Name("two-factor.enable").Use(required, verified, hasPassword)
	r.Post("/user/two-factor/confirm", h.Confirm).
		Name("two-factor.confirm").
		Use(required, verified, hasPassword, confirm)
	r.Post("/user/two-factor/recovery-codes", h.RegenerateRecoveryCodes).
		Name("two-factor.recovery-codes").
		Use(required, verified, hasPassword, confirm)
	r.Post("/user/two-factor/disable", h.Disable).
		Name("two-factor.disable").
		Use(required, verified, hasPassword, confirm)

	r.Get("/two-factor-challenge", h.ShowChallenge).Name("two-factor.login").Use(auth.Guest(acc.Web))
	r.Post("/two-factor-challenge", h.Challenge).Use(auth.Guest(acc.Web), challenge)

	return nil
}
