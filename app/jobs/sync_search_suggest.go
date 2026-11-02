package jobs

import (
	"context"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// SyncSearchSuggest drains the suggest dirty queue into the suggest table; scheduled every minute in routes/schedule.go.
type SyncSearchSuggest struct{}

// Kind names the job on the queue.
func (SyncSearchSuggest) Kind() string { return "search.sync_suggest" }

func handleSyncSearchSuggest(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ SyncSearchSuggest) error {
		db, err := app.DB()
		if err != nil {
			return err
		}
		return search.DrainSuggest(ctx, db)
	}, queue.Timeout(10*time.Minute), queue.Retries(1))
}
