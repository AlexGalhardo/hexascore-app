package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"
)

// PruneTrending deletes trending snapshots older than 90 days; scheduled daily at 04:45 in routes/schedule.go.
type PruneTrending struct{}

// Kind names the job on the queue.
func (PruneTrending) Kind() string { return "recommend.prune_trending" }

func handlePruneTrending(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ PruneTrending) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		res, err := db.ExecContext(ctx, `DELETE FROM trending_snapshots WHERE computed_at < $1`,
			app.Now().UTC().AddDate(0, 0, -90))
		if err != nil {
			return err
		}

		n, _ := res.RowsAffected()
		app.Log.InfoContext(ctx, "prune trending", "deleted", n)

		return nil
	})
}
