package jobs

import (
	"context"
	"time"

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
		p, err := scoring.RecomputeChunk(ctx, db, app.Now().UTC(), scoring.ChunkBudget)
		if err != nil {
			return err
		}
		if !p.Done {
			q, err := app.Queue()
			if err != nil {
				return err
			}

			return q.Dispatch(ctx, RecomputeScores{}, queue.Delay(time.Second))
		}
		app.Log.InfoContext(ctx, "recompute scores",
			"items", p.Items, "chunks", p.Chunks, "duration_ms", app.Now().Sub(p.StartedAt).Milliseconds())

		return nil
	}, queue.Timeout(3*time.Hour), queue.Retries(1))
}
