package jobs

import (
	"context"
	"sync"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/library"
)

// ExpireExports deletes export files past their 24 h link and marks them expired; scheduled hourly in
// routes/schedule.go (SDD 14 §7.4 step 5).
type ExpireExports struct{}

// Kind names the job on the queue.
func (ExpireExports) Kind() string { return "library.expire_exports" }

// exports builds the service once per handler: the local disk holds an open directory handle.
func exports(app *vukem.App) (*library.Exports, error) {
	acc, err := accounts.New(app)
	if err != nil {
		return nil, err
	}

	return library.NewExports(app, acc)
}

// handleLibraryExport builds the ZIP of a requested export (SDD 14 §6.5); after the last attempt the row is
// marked failed so the cooldown does not lock the user out.
func handleLibraryExport(app *vukem.App) queue.Registration {
	get := sync.OnceValues(func() (*library.Exports, error) { return exports(app) })

	return queue.Handle(func(ctx context.Context, j library.ExportJob) error {
		svc, err := get()
		if err != nil {
			return err
		}

		err = svc.Build(ctx, j.ExportID)
		if a, ok := queue.AttemptFrom(ctx); err != nil && (!ok || a.Last()) {
			if ferr := svc.Fail(ctx, j.ExportID); ferr != nil {
				app.Log.ErrorContext(ctx, "mark export failed", "export", j.ExportID, "err", ferr)
			}
		}

		return err
	}, queue.Retries(3), queue.Timeout(5*time.Minute))
}

func handleExpireExports(app *vukem.App) queue.Registration {
	get := sync.OnceValues(func() (*library.Exports, error) { return exports(app) })

	return queue.Handle(func(ctx context.Context, _ ExpireExports) error {
		svc, err := get()
		if err != nil {
			return err
		}

		n, err := svc.Expire(ctx)
		app.Log.InfoContext(ctx, "expire exports", "expired", n)

		return err
	})
}
