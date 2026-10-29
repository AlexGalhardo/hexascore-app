package jobs

import (
	"context"
	"errors"
	"sync"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/library"
)

// DeletionPurge anonymises accounts whose 30-day grace period ended; scheduled daily at 03:00 in routes/schedule.go.
type DeletionPurge struct{}

// Kind names the job on the queue.
func (DeletionPurge) Kind() string { return "accounts.deletion_purge" }

func handleDeletionPurge(app *vukem.App) queue.Registration {
	svcOnce := sync.OnceValues(func() (*library.Exports, error) { return exports(app) })

	return queue.Handle(func(ctx context.Context, _ DeletionPurge) error {
		acc, err := accounts.New(app)
		if err != nil {
			return err
		}

		ids, err := acc.PurgeDue(ctx)
		app.Log.InfoContext(ctx, "deletion purge", "anonymised", len(ids))

		svc, xerr := svcOnce()
		if xerr != nil {
			return errors.Join(err, xerr)
		}

		for _, id := range ids {
			err = errors.Join(err, svc.Purge(ctx, id))
		}

		return err
	})
}
