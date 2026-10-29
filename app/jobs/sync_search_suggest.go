package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// SyncSearchSuggest drains search_dirty into search_suggest; scheduled every minute in routes/schedule.go.
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
	})
}
