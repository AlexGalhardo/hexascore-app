package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// RecomputeScores rebuilds the priors then recomputes every rated item; scheduled daily at 04:00 in routes/schedule.go.
type RecomputeScores struct{}

// Kind names the job on the queue.
func (RecomputeScores) Kind() string { return "scoring.recompute_scores" }

func handleRecomputeScores(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ RecomputeScores) error {
		db, err := app.DB()
		if err != nil {
			return err
		}
		now := app.Now().UTC()
		if err := scoring.RebuildPriors(ctx, db, now); err != nil {
			return err
		}
		n, err := scoring.RecomputeAll(ctx, db, now)
		app.Log.InfoContext(ctx, "recompute scores", "items", n)

		return err
	})
}
