package jobs

import (
	"cmp"
	"context"
	"sync"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/app/community"
)

// CommunityHourly expires stale drafts, refreshes contributor tiers and releases idle claims.
type CommunityHourly struct{}

// Kind names the job on the queue.
func (CommunityHourly) Kind() string { return "community.hourly" }

func communityCovers(app *vukem.App) (*storage.Local, error) {
	dir := cmp.Or(app.Source.Get("COMMUNITY_COVERS_DIR"), "storage/community")
	disk, err := storage.NewLocal(dir, "")
	if err != nil {
		return nil, err
	}
	app.OnShutdown(func(context.Context) error { return disk.Close() })

	return disk, nil
}

func handleCommunityHourly(app *vukem.App) queue.Registration {
	get := sync.OnceValues(func() (*storage.Local, error) { return communityCovers(app) })

	return queue.Handle(func(ctx context.Context, _ CommunityHourly) error {
		db, err := app.DB()
		if err != nil {
			return err
		}
		disk, err := get()
		if err != nil {
			return err
		}

		return community.Hourly(ctx, db, app.Now(), disk)
	})
}
