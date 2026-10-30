package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/api/changes"
)

// PruneAPIChanges deletes api_changes rows past the 30-day retention; scheduled daily at 04:50 in routes/schedule.go.
type PruneAPIChanges struct{}

// Kind names the job on the queue.
func (PruneAPIChanges) Kind() string { return "api.prune_changes" }

func handlePruneAPIChanges(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ PruneAPIChanges) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		n, err := changes.Prune(ctx, db, app.Now())
		app.Log.InfoContext(ctx, "prune api changes", "deleted", n)

		return err
	})
}
