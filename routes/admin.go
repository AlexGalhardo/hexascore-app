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
	"github.com/AlexGalhardo/vukem/view"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/admin"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/community"
	appmiddleware "github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/moderation"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/entities"
)

// Admin mounts the admin panel at /admin (vukem install:admin). Admins are
// users with role admin (SDD 03 section 6.3); resources live in app/admin;
// feature flags are at /admin/_flags.
func Admin(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	allowlist, err := appmiddleware.ParseAllowlist(app.Source.Get("ADMIN_IP_ALLOWLIST"))
	if err != nil {
		return err
	}

	// Router-wide: the 2FA gate and allowlist act on /admin, the session limits
	// on every request of an admin (SDD 13 §7.5).
	r.Use(appmiddleware.AdminGuard(appmiddleware.AdminGuardConfig{
		User: acc.Web.User,
		TwoFactorEnabled: func(ctx context.Context, id int64) (bool, error) {
			tf, err := acc.TwoFactor(ctx, id)
			return tf.Enabled(), err
		},
		Logout:    acc.Web.Logout,
		Allowlist: allowlist,
		Now:       app.Now,
		SetupPath: "/user/two-factor",
		LoginPath: "/login",
	}))

	// The panel has its own gate, so its policies never mix with the app's.
	gate := authz.New(func(ctx context.Context) (*models.User, error) {
		u, err := auth.UserFrom[models.User](ctx, acc.Web.Name())
		return &u, err
	})
	isAdmin := admin.Gate(app.Now)
	gate.DefineResult(vadmin.AbilityAdmin, isAdmin)
	gate.DefineResult(vadmin.AbilityAudit, isAdmin)
	gate.DefineResult(vadmin.AbilityFlags, isAdmin)

	admin.DefineModeration(gate, app.Now, func(ctx context.Context) bool {
		ff, err := app.Flags()

		return err == nil && ff.Enabled(ctx, community.TrustedModeratorsFlag)
	})

	if err := gate.Register(admin.Policies(app.Now)...); err != nil {
		return err
	}

	// The flags registered with vukem.WithFlags (vukem install:flags) are
	// toggled at /admin/_flags.
	ff, err := app.Flags()
	if err != nil {
		return err
	}

	views, err := admin.ViewsFS()
	if err != nil {
		return err
	}

	panel, err := vadmin.New(vadmin.Config{
		DB:         acc.DB,
		Guard:      acc.Web,
		Authorizer: gate,
		Title:      app.Config.Name + " admin",
		Flags:      ff,
		Views:      views,
	})
	if err != nil {
		return err
	}

	if err := panel.Register(admin.Resources(acc, acc.DB.Dialect().Name() == "postgres")...); err != nil {
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

	reportH := admin.ReportHandlers{Accounts: acc, Reports: ReportService(app, acc), Authorize: bans.Authorize}
	r.Post("/admin/reports/{id}/resolve", reportH.Resolve).
		Name("admin.reports.resolve").
		Use(moderator...)

	q, err := app.Queue()
	if err != nil {
		return err
	}

	engine, err := view.New(view.Config{FS: views, DefaultLayout: "admin"})
	if err != nil {
		return err
	}

	catalogH := admin.CatalogHandlers{Accounts: acc, Queue: q, Views: engine, Authorize: bans.Authorize}
	if len(ff.All()) > 0 {
		catalogH.FlagsURL = "/admin/_flags"
	}
	r.Get("/admin/items/{id}/manage", catalogH.Manage).Name("admin.items.manage").Use(auth.Required(acc.Web), csrfMW)
	r.Post("/admin/items/{id}/affiliate-links", catalogH.SaveAffiliateLink).
		Name("admin.items.affiliate.save").Use(moderator...)
	r.Post("/admin/items/{id}/affiliate-links/{store}/delete", catalogH.DeleteAffiliateLink).
		Name("admin.items.affiliate.delete").Use(moderator...)
	r.Post("/admin/items/{id}/merge", catalogH.Merge).Name("admin.items.merge").Use(moderator...)
	applier := community.Applier{DB: acc.DB, Queue: q, Now: app.Now}
	catalogH.Editor = &applier
	catalogH.AuthorizeApply = func(c *web.Ctx) error {
		return gate.Authorize(c.Context(), admin.AbilityApplySuggestion, nil)
	}
	r.Post("/admin/catalog/{type}/{id}/fields", catalogH.EditFields).Name("admin.catalog.fields").Use(moderator...)
	moderationH := admin.ModerationHandlers{
		Pages:    catalogH,
		Queue:    moderation.Queue{DB: acc.DB, Now: app.Now, Applier: applier, Jobs: q},
		Reverter: applier,
		AuthorizeAudit: func(c *web.Ctx) error {
			return gate.Authorize(c.Context(), admin.AbilityViewCatalogAudit, nil)
		},
		AuthorizeRevert: func(c *web.Ctx) error {
			return gate.Authorize(c.Context(), admin.AbilityRevertCatalog, nil)
		},
		AuthorizeLock: func(c *web.Ctx) error {
			return gate.Authorize(c.Context(), admin.AbilityLockCatalog, nil)
		},
		Authorize: func(c *web.Ctx) error { return gate.Authorize(c.Context(), admin.AbilityModerate, nil) },
		AuthorizeApply: func(c *web.Ctx) error {
			return gate.Authorize(c.Context(), admin.AbilityApplySuggestion, nil)
		},
	}
	r.Get("/admin/moderation", moderationH.Index).Name("admin.moderation").Use(auth.Required(acc.Web), csrfMW)
	modMW := []func(http.Handler) http.Handler{auth.Required(acc.Web), csrfMW}
	r.Get("/admin/moderation/metrics", moderationH.Metrics).Name("admin.moderation.metrics").Use(modMW...)
	r.Post("/admin/moderation/{id}/claim", moderationH.Claim).Name("admin.moderation.claim").Use(modMW...)
	r.Post("/admin/moderation/{id}/release", moderationH.Release).Name("admin.moderation.release").Use(modMW...)
	r.Get("/admin/moderation/{id}", moderationH.Detail).Name("admin.moderation.detail").Use(modMW...)
	r.Post("/admin/moderation/{id}/recommend", moderationH.Recommend).
		Name("admin.moderation.recommend").Use(modMW...)
	r.Post("/admin/moderation/{id}/approve", moderationH.Approve).Name("admin.moderation.approve").Use(modMW...)
	r.Post("/admin/moderation/{id}/reject", moderationH.Reject).Name("admin.moderation.reject").Use(modMW...)
	r.Post("/admin/moderation/{id}/clone", moderationH.Clone).Name("admin.moderation.clone").Use(modMW...)
	r.Post("/admin/moderation/bulk", moderationH.Bulk).Name("admin.moderation.bulk").Use(modMW...)
	r.Get("/admin/catalog-audit", moderationH.Audit).Name("admin.catalog_audit").Use(modMW...)
	r.Post("/admin/catalog-audit/{id}/revert", moderationH.Revert).Name("admin.catalog_audit.revert").Use(modMW...)
	r.Get("/admin/catalog-audit/export.csv", moderationH.AuditExport("csv")).
		Name("admin.catalog_audit.export_csv").Use(modMW...)
	r.Get("/admin/catalog-audit/export.ndjson", moderationH.AuditExport("ndjson")).
		Name("admin.catalog_audit.export_ndjson").Use(modMW...)
	r.Get("/admin/locks", moderationH.Locks).Name("admin.locks").Use(modMW...)
	r.Post("/admin/locks", moderationH.CreateLock).Name("admin.locks.create").Use(modMW...)
	r.Post("/admin/locks/{id}/delete", moderationH.DeleteLock).Name("admin.locks.delete").Use(modMW...)
	reviewsH := admin.IngestReviewHandlers{
		Pages:   catalogH,
		Reviews: entities.ReviewQueue{DB: acc.DB, KindOf: catalog.RoleKind},
		Authorize: func(c *web.Ctx) error {
			return gate.Authorize(c.Context(), admin.AbilityRunIngestion, nil)
		},
	}
	r.Get("/admin/ingest/reviews", reviewsH.List).Name("admin.ingest.reviews").Use(modMW...)
	r.Post("/admin/ingest/reviews/{id}", reviewsH.Resolve).Name("admin.ingest.reviews.resolve").Use(modMW...)
	ingestH := admin.IngestHandlers{
		Pages: catalogH, DB: acc.DB, Queue: q, Flags: ff, Now: app.Now,
		AuthorizeView: func(c *web.Ctx) error {
			return gate.Authorize(c.Context(), admin.AbilityViewIngestion, nil)
		},
		AuthorizeRun: func(c *web.Ctx) error {
			return gate.Authorize(c.Context(), admin.AbilityRunIngestion, nil)
		},
		AuthorizeRollback: func(c *web.Ctx) error {
			return gate.Authorize(c.Context(), admin.AbilityRollbackIngestion, nil)
		},
	}
	r.Get("/admin/ingest", ingestH.Index).Name("admin.ingest").Use(modMW...)
	r.Post("/admin/ingest/runs", ingestH.Trigger).Name("admin.ingest.runs").Use(modMW...)
	r.Post("/admin/ingest/runs/{id}/cancel", ingestH.Cancel).Name("admin.ingest.runs.cancel").Use(modMW...)
	r.Post("/admin/ingest/runs/{id}/rollback", ingestH.Rollback).Name("admin.ingest.runs.rollback").Use(modMW...)
	r.Get("/admin/ingestion", catalogH.Ingestion).Name("admin.ingestion").Use(auth.Required(acc.Web), csrfMW)

	return nil
}
