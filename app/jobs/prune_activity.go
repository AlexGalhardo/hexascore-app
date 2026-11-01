package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/social"
)

// PruneActivity deletes feed activity events past 180 days (SDD 14 NFR-14-9). Scheduled daily at 04:40 in
// routes/schedule.go; the table exists only on Postgres (VF-033).
type PruneActivity struct{}

// Kind names the job on the queue.
func (PruneActivity) Kind() string { return "social.prune_activity" }

func handlePruneActivity(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ PruneActivity) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		if db.Dialect().Name() != "postgres" {
			return nil
		}

		n, err := social.PruneActivity(ctx, db, app.Now())
		app.Log.InfoContext(ctx, "prune activity events", "deleted", n)

		return err
	})
}
