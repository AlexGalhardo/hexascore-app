package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/trending"
)

// ComputeTrending runs the rollup, the per-type snapshot and the cache invalidation; scheduled at minute 5 of every
// hour in routes/schedule.go. On failure the readers keep the last snapshot.
type ComputeTrending struct{}

// Kind names the job on the queue.
func (ComputeTrending) Kind() string { return "recommend.compute_trending" }

func handleComputeTrending(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ ComputeTrending) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		now := app.Now().UTC()
		if err := trending.Rollup(ctx, db, now); err != nil {
			return err
		}

		if err := trending.Snapshot(ctx, db, now); err != nil {
			return err
		}

		store, err := app.Cache()
		if err != nil {
			return err
		}

		for _, typ := range trending.Types {
			for _, locale := range trending.Locales {
				if err := store.Forget(ctx, trending.CacheKey(typ, locale)); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// TrendingRebuild runs the queued trending_rebuild (kind shared with app/recommend/trending.Rebuild).
func handleTrendingRebuild(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, j trending.Rebuild) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		return j.Run(ctx, db, app.Now().UTC())
	})
}
