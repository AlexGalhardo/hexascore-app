package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/similar"
)

// ComputeSimilar refreshes item_similar for the site variant; scheduled daily at 02:00 in routes/schedule.go.
type ComputeSimilar struct{}

// Kind names the job on the queue.
func (ComputeSimilar) Kind() string { return "recommend.compute_similar" }

func handleComputeSimilar(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ ComputeSimilar) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		n, err := similar.Compute(ctx, db, similar.Options{Now: app.Now().UTC()})
		app.Log.InfoContext(ctx, "compute similar", "written", n)

		return err
	})
}
