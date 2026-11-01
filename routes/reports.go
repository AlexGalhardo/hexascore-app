package routes

import (
	"context"
	"errors"

	"github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/ratelimit"
	"github.com/AlexGalhardo/vukem/router"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/controllers"
	"github.com/AlexGalhardo/hexascore-app/app/http/limits"
	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/config"
)

// ReportService builds the report service shared by the member routes and the admin resolve route.
func ReportService(app *vukem.App, acc *accounts.Accounts) *library.Reports {
	return &library.Reports{
		DB: acc.DB, Events: app.Events, Now: app.Now,
		Ban: func(ctx context.Context, actorID, userID int64, reason string) error {
			err := acc.Ban(ctx, actorID, userID, reason, nil)
			if errors.Is(err, accounts.ErrAlreadyBanned) {
				return nil
			}

			return err
		},
	}
}

// Reports registers POST /reports on the web (10 per hour per user, SDD 14 NFR-14-4) and on /app/v1 (ability
// profile:write), and the public contact form (5 per hour per IP). The tables exist only on PostgreSQL (VF-033).
func Reports(r *router.Router, app *vukem.App) error {
	acc, err := accounts.New(app)
	if err != nil {
		return err
	}

	if acc.DB.Dialect().Name() != "postgres" {
		return nil
	}

	limiters, err := limits.Factory(app)
	if err != nil {
		return err
	}

	svc := ReportService(app, acc)

	sc, err := SocialService(app, acc)
	if err != nil {
		return err
	}

	limit := limits.Reports.From(limiters)

	h := controllers.NewReports(svc, sc, acc.Web.Name(), false)
	r.Post("/reports", h.Create).Name("reports.create").
		Use(auth.Required(acc.Web), auth.Verified(acc.Web), ratelimit.Middleware(limit, ratelimit.ByUser(acc.Web.Name())))

	api := controllers.NewReports(svc, sc, acc.API.Name(), true)
	stack, err := AppWriteStack(app, acc, limiters,
		auth.Ability(acc.API, "profile:write"),
		ratelimit.Middleware(limit, ratelimit.ByUser(acc.API.Name())),
	)
	if err != nil {
		return err
	}
	r.Group("/app/v1", func(g *router.Group) {
		g.Use(stack...)
		g.Post("/reports", api.Create).Name("reports.create")
	}).Name("app.v1.")

	key, err := app.DeriveKey("contact.form", 32)
	if err != nil {
		return err
	}

	contact := controllers.NewContact(&library.Contact{
		DB:     acc.DB,
		Mailer: acc.Mailer,
		Inbox:  config.ContactInbox(app.Source.Get),
		Key:    key,
		Now:    app.Now,
		Log:    app.Log,
	}, func(c *web.Ctx) (models.User, error) { return acc.Web.User(c) })
	byIP := ratelimit.Middleware(limits.Contact.From(limiters), ratelimit.ByIP)
	r.Get("/contact", contact.Index).Name("contact")
	r.Post("/contact", contact.Send).Name("contact.send").Use(byIP)

	return nil
}
