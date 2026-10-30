package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/webhooks"
)

// FanoutWebhooks delivers api_changes to webhook endpoints. Scheduled every
// minute in routes/schedule.go.
type FanoutWebhooks struct{}

// Kind names the job on the queue.
func (FanoutWebhooks) Kind() string { return "webhooks.fanout" }

func handleFanoutWebhooks(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ FanoutWebhooks) error {
		svc, err := webhooks.OpenApp(app)
		if err != nil {
			return err
		}

		q, err := app.Queue()
		if err != nil {
			return err
		}

		return svc.Fanout(ctx, q)
	})
}

func handleWebhookDeliver(app *vukem.App) queue.Registration {
	return webhooks.DeliveryRegistration(app)
}
