package routes

import (
	"context"
	"net/http"

	"github.com/AlexGalhardo/vukem"
	vadmin "github.com/AlexGalhardo/vukem/admin"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/authz"
	"github.com/AlexGalhardo/vukem/csrf"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/admin"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// Admin mounts the admin panel at /admin (vukem install:admin). Admins are
// users with role admin (SDD 03 section 6.3); resources live in app/admin;
// feature flags are at /admin/_flags.
func Admin(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	// The panel has its own gate, so its policies never mix with the app's.
	gate := authz.New(func(ctx context.Context) (*models.User, error) {
		u, err := auth.UserFrom[models.User](ctx, acc.Web.Name())
		return &u, err
	})
	isAdmin := admin.Gate(app.Now)
	gate.DefineResult(vadmin.AbilityAdmin, isAdmin)
	gate.DefineResult(vadmin.AbilityAudit, isAdmin)
	gate.DefineResult(vadmin.AbilityFlags, isAdmin)

	if err := gate.Register(admin.Policies(app.Now)...); err != nil {
		return err
	}

	// The flags registered with vukem.WithFlags (vukem install:flags) are
	// toggled at /admin/_flags.
	ff, err := app.Flags()
	if err != nil {
		return err
	}

	panel, err := vadmin.New(vadmin.Config{
		DB:         acc.DB,
		Guard:      acc.Web,
		Authorizer: gate,
		Title:      app.Config.Name + " admin",
		Flags:      ff,
	})
	if err != nil {
		return err
	}

	if err := panel.Register(admin.Resources(acc.DB.Dialect().Name() == "postgres")...); err != nil {
		return err
	}

	if err := panel.Mount(r); err != nil {
		return err
	}

	csrfMW, err := csrf.Middleware(csrf.Config{TokenFallback: true})
	if err != nil {
		return err
	}

	bans := admin.BanHandlers{
		Accounts:  acc,
		Authorize: func(c *web.Ctx) error { return gate.Authorize(c.Context(), vadmin.AbilityAdmin, nil) },
	}
	r.Post("/admin/users/{id}/ban", bans.Ban).Name("admin.users.ban").Use(auth.Required(acc.Web), csrfMW)
	r.Post("/admin/users/{id}/unban", bans.Unban).Name("admin.users.unban").Use(auth.Required(acc.Web), csrfMW)

	scoringH := admin.ScoringHandlers{
		Accounts:  acc,
		Writer:    &scoring.Writer{DB: acc.DB, Events: app.Events, Now: app.Now},
		Authorize: bans.Authorize,
	}
	moderator := []func(http.Handler) http.Handler{auth.Required(acc.Web), csrfMW}
	r.Post("/admin/scoring/bursts/{id}/approve", scoringH.ApproveBurst).
		Name("admin.scoring.bursts.approve").
		Use(moderator...)
	r.Post("/admin/scoring/bursts/{id}/exclude", scoringH.ExcludeBurst).
		Name("admin.scoring.bursts.exclude").
		Use(moderator...)
	r.Post("/admin/scoring/flags/{user}/{item}/restore", scoringH.RestoreFlag).
		Name("admin.scoring.flags.restore").
		Use(moderator...)

	return nil
}
