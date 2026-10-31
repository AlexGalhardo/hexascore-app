package trending

import (
	"context"
	"fmt"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/events"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
)

// batchSize is the number of item ids per queued rebuild.
const batchSize = 500

// Rebuild is the queued trending_rebuild: it re-rolls the days from FromDay (YYYY-MM-DD) to today for ItemIDs only.
type Rebuild struct {
	ItemIDs []string `json:"item_ids"`
	FromDay string   `json:"from_day"`
}

// Kind names the job on the queue.
func (Rebuild) Kind() string { return "recommend.trending_rebuild" }

// Run re-rolls the job's items; the next ComputeTrending picks the result up.
func (r Rebuild) Run(ctx context.Context, db *database.DB, now time.Time) error {
	from, err := time.Parse(time.DateOnly, r.FromDay)
	if err != nil {
		return fmt.Errorf("trending rebuild from_day: %w", err)
	}

	days := int(now.UTC().Truncate(24*time.Hour).Sub(from)/(24*time.Hour)) + 1

	return RollupDays(ctx, db, now, max(days, 1), r.ItemIDs)
}

// userItems lists the items the user's window activity reached: ratings, statuses, list entries and suggestions.
const userItems = `SELECT item_id::text FROM (
  SELECT item_id FROM rating_events WHERE user_id = $1 AND at >= $2
  UNION SELECT item_id FROM item_status WHERE user_id = $1 AND updated_at >= $2
  UNION SELECT li.item_id FROM list_items li JOIN lists l ON l.id = li.list_id WHERE l.user_id = $1 AND li.added_at >= $2
  UNION SELECT s.target_id FROM suggestions s WHERE s.author_id = $1 AND s.target_type = 'item'
    AND s.target_id IS NOT NULL AND s.submitted_at >= $2
) x ORDER BY 1`

// Listeners queues a rebuild of the items a banned user touched in the window (SDD 08 §6 "Jobs"). An unban
// reverses the weights the same way.
func Listeners(app *vukem.App) []events.Registration {
	return []events.Registration{
		events.On(func(ctx context.Context, e accounts.UserBanned) error { return enqueue(ctx, app, e.UserID) }),
		events.On(func(ctx context.Context, e accounts.UserUnbanned) error { return enqueue(ctx, app, e.UserID) }),
	}
}

func enqueue(ctx context.Context, app *vukem.App, userID int64) error {
	db, err := app.DB()
	if err != nil {
		return err
	}

	q, err := app.Queue()
	if err != nil {
		return err
	}

	from := app.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -(WindowDays - 1))

	rows, err := db.QueryContext(ctx, userItems, userID, from)
	if err != nil {
		return fmt.Errorf("trending rebuild items of user %d: %w", userID, err)
	}

	defer func() { _ = rows.Close() }()

	var ids []string

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}

		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for len(ids) > 0 {
		n := min(batchSize, len(ids))
		if err := q.Dispatch(ctx, Rebuild{ItemIDs: ids[:n], FromDay: from.Format(time.DateOnly)}); err != nil {
			return fmt.Errorf("trending rebuild dispatch: %w", err)
		}

		ids = ids[n:]
	}

	return nil
}
