package jobs

import (
	"context"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest"
)

// WorkerHeartbeat marks the ingestion worker alive (RFC-001 section 8). The scheduler in the web process dispatches it
// every minute to the ingest queue, so the row is written by whichever process consumes that queue: the worker
// service once it exists, the web process until then. The readiness check of the web reads the row.
type WorkerHeartbeat struct{}

// Kind names the job on the queue.
func (WorkerHeartbeat) Kind() string { return "ingest.worker_heartbeat" }

func handleWorkerHeartbeat(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ WorkerHeartbeat) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		return ingest.WriteWorkerHeartbeat(ctx, db)
	}, queue.OnQueue("ingest"), queue.Timeout(time.Minute), queue.Retries(0))
}
