package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/webhooks"
)

// PruneWebhookDeliveries deletes delivery-log rows past 30 days. Scheduled
// daily at 04:40 in routes/schedule.go (SDD 05 §5).
type PruneWebhookDeliveries struct{}

// Kind names the job on the queue.
func (PruneWebhookDeliveries) Kind() string { return "webhooks.prune_deliveries" }

func handlePruneWebhookDeliveries(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ PruneWebhookDeliveries) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		n, err := webhooks.PruneDeliveries(ctx, db, app.Now())
		app.Log.InfoContext(ctx, "prune webhook deliveries", "deleted", n)

		return err
	})
}
