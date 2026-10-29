package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// RefreshEntityStats recomputes entity_stats; scheduled hourly in routes/schedule.go.
type RefreshEntityStats struct{}

// Kind names the job on the queue.
func (RefreshEntityStats) Kind() string { return "catalog.refresh_entity_stats" }

func handleRefreshEntityStats(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ RefreshEntityStats) error {
		db, err := app.DB()
		if err != nil {
			return err
		}
		return catalog.RefreshEntityStats(ctx, db)
	})
}
