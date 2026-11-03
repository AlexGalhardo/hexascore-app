package jobs

import (
	"context"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"
)

// PruneTrending deletes trending snapshots older than 90 days, external popularity values older than 35 days (the
// score reads today and 7 days back) and pin/ban overrides that expired more than a year ago; scheduled daily at 04:45 in
// routes/schedule.go.
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

		now := app.Now().UTC()

		ext, err := db.ExecContext(
			ctx,
			`DELETE FROM item_popularity_external WHERE day < $1::date`,
			now.AddDate(0, 0, -35),
		)
		if err != nil {
			return err
		}

		over, err := db.ExecContext(ctx, `DELETE FROM trending_overrides WHERE until IS NOT NULL AND until < $1`,
			now.AddDate(-1, 0, 0))
		if err != nil {
			return err
		}

		e, _ := ext.RowsAffected()
		o, _ := over.RowsAffected()
		app.Log.InfoContext(ctx, "prune trending", "snapshots", n, "external", e, "overrides", o)

		return nil
	}, queue.Timeout(10*time.Minute), queue.Retries(1))
}
