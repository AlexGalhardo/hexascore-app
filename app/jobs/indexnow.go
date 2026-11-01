package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/indexnow"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// IndexNowChanges tells IndexNow engines about changed catalogue pages: items created, edited, published, hidden or
// deleted (api_changes) and indexable entities updated since the last run. Scheduled every minute in
// routes/schedule.go, so writes never wait on it; it queues one indexnow.Job per run, which vukem's handler sends
// with retries (https://vukem.dev/docs/digging-deeper/seo-integrations). Off while INDEXNOW_KEY is unset.
type IndexNowChanges struct{}

// Kind names the job on the queue.
func (IndexNowChanges) Kind() string { return "seo.indexnow_changes" }

func handleIndexNowChanges(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ IndexNowChanges) error {
		if !app.Config.IndexNow.Enabled() {
			return nil
		}

		db, err := app.DB()
		if err != nil {
			return err
		}

		q, err := app.Queue()
		if err != nil {
			return err
		}

		return RunIndexNowChanges(
			ctx,
			db,
			app.Config.URL,
			app.Now().UTC(),
			func(ctx context.Context, urls []string) error {
				return q.Dispatch(ctx, indexnow.Job{URLs: urls})
			},
		)
	})
}

// indexNowHandlers is vukem's indexnow.Job handler, or nothing while INDEXNOW_KEY is unset.
func indexNowHandlers(app *vukem.App) []queue.Registration {
	n, err := indexnow.New(app.Config.IndexNow, app.Config.URL, nil)
	if err != nil {
		// ErrDisabled, or a config vukem.New already refused at boot.
		return nil
	}

	return []queue.Registration{n.Handler()}
}

// RunIndexNowChanges reads the changes after indexnow_cursor, hands their URLs to notify and moves the cursor in the
// same transaction, so a failed notify retries the same batch on the next run.
func RunIndexNowChanges(
	ctx context.Context, db *database.DB, siteURL string, now time.Time,
	notify func(context.Context, []string) error,
) error {
	base := strings.TrimSuffix(siteURL, "/")

	return db.Tx(ctx, func(tx *database.Tx) error {
		var c catalog.IndexNowCursor
		if err := tx.QueryRowContext(ctx,
			`SELECT last_change_id, entities_since FROM indexnow_cursor WHERE id = 1 FOR UPDATE`).
			Scan(&c.LastChangeID, &c.EntitiesSince); err != nil {
			return fmt.Errorf("indexnow: read cursor: %w", err)
		}

		urls, next, err := catalog.ChangedURLs(ctx, tx, base, c, now)
		if err != nil {
			return err
		}

		if len(urls) > 0 {
			if err := notify(ctx, urls); err != nil {
				return fmt.Errorf("indexnow: queue %d urls: %w", len(urls), err)
			}
		}

		_, err = tx.ExecContext(ctx, `UPDATE indexnow_cursor SET last_change_id = $1, entities_since = $2 WHERE id = 1`,
			next.LastChangeID, next.EntitiesSince)

		return err
	})
}
