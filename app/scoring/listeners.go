package scoring

import (
	"context"
	"fmt"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/events"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
)

// RecomputeItemScore is the queued per-item recompute (kind shared with app/jobs.RecomputeItemScore).
type RecomputeItemScore struct {
	ItemID string `json:"item_id"`
}

// Kind names the job on the queue.
func (RecomputeItemScore) Kind() string { return "scoring.recompute_item_score" }

// Listeners subscribes to SDD 03's ban and unban events (open issue 02-8). Register them from a provider:
// they need the app's database and queue, which exist only after boot.
func Listeners(app *vukem.App) []events.Registration {
	return []events.Registration{
		events.On(
			func(ctx context.Context, e accounts.UserBanned) error { return enqueueUserItems(ctx, app, e.UserID) },
		),
		events.On(
			func(ctx context.Context, e accounts.UserUnbanned) error { return enqueueUserItems(ctx, app, e.UserID) },
		),
	}
}

// enqueueUserItems dispatches RecomputeItemScore for every item the user rated, in batches of 500 ids ascending,
// one transaction per batch.
func enqueueUserItems(ctx context.Context, app *vukem.App, userID int64) error {
	db, err := app.DB()
	if err != nil {
		return err
	}
	q, err := app.Queue()
	if err != nil {
		return err
	}

	last := nilUUID
	for {
		ids, err := userItemPage(ctx, db, userID, last)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		if err := dispatchBatch(ctx, db, q, ids); err != nil {
			return fmt.Errorf("enqueue recomputes for user %d: %w", userID, err)
		}
		last = ids[len(ids)-1]
	}
}

func userItemPage(ctx context.Context, db *database.DB, userID int64, after string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT item_id::text FROM ratings WHERE user_id = $1 AND item_id > $2::uuid
ORDER BY item_id LIMIT $3`, userID, after, BatchSize)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func dispatchBatch(ctx context.Context, db *database.DB, q *queue.Queue, ids []string) error {
	return db.Tx(ctx, func(tx *database.Tx) error {
		for _, id := range ids {
			if err := q.DispatchTx(ctx, tx, RecomputeItemScore{ItemID: id}); err != nil {
				return err
			}
		}

		return nil
	})
}
