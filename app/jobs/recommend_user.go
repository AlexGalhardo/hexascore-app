package jobs

import (
	"context"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/recommend/foryou"
)

func handleRecommendUser(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, j foryou.RecommendUser) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		svc := &foryou.Service{DB: db, Now: app.Now}

		return svc.Run(ctx, j)
	}, queue.Retries(2), queue.Timeout(30*time.Second))
}
