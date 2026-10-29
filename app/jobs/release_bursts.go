package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// ReleaseBursts auto-resolves bursts unreviewed for 14 days; scheduled hourly in routes/schedule.go.
type ReleaseBursts struct{}

// Kind names the job on the queue.
func (ReleaseBursts) Kind() string { return "scoring.release_bursts" }

func handleReleaseBursts(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ ReleaseBursts) error {
		db, err := app.DB()
		if err != nil {
			return err
		}
		n, err := scoring.ReleaseBursts(ctx, db, app.Now().UTC())
		app.Log.InfoContext(ctx, "release bursts", "resolved", n)

		return err
	})
}
