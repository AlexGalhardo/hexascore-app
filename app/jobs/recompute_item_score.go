package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

func handleRecomputeItemScore(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, j scoring.RecomputeItemScore) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		return scoring.RecomputeItem(ctx, db, j.ItemID, app.Now().UTC())
	})
}
