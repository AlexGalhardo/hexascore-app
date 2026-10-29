package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// PruneRatingEvents deletes rating events past the 180-day retention; scheduled daily at 04:30 in routes/schedule.go.
type PruneRatingEvents struct{}

// Kind names the job on the queue.
func (PruneRatingEvents) Kind() string { return "scoring.prune_rating_events" }

func handlePruneRatingEvents(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ PruneRatingEvents) error {
		db, err := app.DB()
		if err != nil {
			return err
		}
		n, err := scoring.PruneRatingEvents(ctx, db, app.Now().UTC())
		app.Log.InfoContext(ctx, "prune rating events", "deleted", n)

		return err
	})
}
