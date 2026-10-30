package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/privacy"
)

// RetainPersonalData strips session IP and user agent past 30 days, prunes client_reports past 7 days, and
// expires export ZIPs; scheduled daily at 03:15 in routes/schedule.go (SDD 13 §7.6).
type RetainPersonalData struct{}

// Kind names the job on the queue.
func (RetainPersonalData) Kind() string { return "privacy.retain" }

func handleRetainPersonalData(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ RetainPersonalData) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		acc, err := accounts.New(app)
		if err != nil {
			return err
		}

		svc, err := library.NewExports(app, acc)
		if err != nil {
			return err
		}

		res, err := privacy.Run(ctx, db, svc, app.Now())
		app.Log.InfoContext(ctx, "privacy retention",
			"sessions", res.SessionsCleared, "reports", res.ReportsDeleted, "exports", res.ExportsExpired)

		return err
	})
}
