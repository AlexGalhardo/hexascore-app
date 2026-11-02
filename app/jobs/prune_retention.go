package jobs

import (
	"context"
	"fmt"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/config"
)

// PruneBatch is how many rows one DELETE removes, so a large backlog never holds a long lock.
const PruneBatch = 5000

// PruneRetention trims the operational tables of config.Retentions (SDD 11 §5, audit T-17). Scheduled daily at
// 04:55 in routes/schedule.go; the tables exist only on Postgres (VF-033).
type PruneRetention struct{}

// Kind names the job on the queue.
func (PruneRetention) Kind() string { return "ops.prune_retention" }

func handlePruneRetention(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ PruneRetention) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		if db.Dialect().Name() != "postgres" {
			return nil
		}

		windows, err := config.RetentionWindows(app.Source.Get)
		if err != nil {
			return err
		}

		for _, w := range windows {
			n, err := PruneTable(ctx, db, w, PruneBatch)
			app.Log.InfoContext(ctx, "prune retention", "table", w.Table, "days", w.Days, "deleted", n)
			if err != nil {
				return err
			}
		}

		return nil
	}, queue.Timeout(15*time.Minute), queue.Retries(1))
}

// PruneTable deletes the rows of w older than its window in batches of batch and returns how many it removed.
// Table and column names come from config.Retentions, never from input.
func PruneTable(ctx context.Context, db *database.DB, w config.Retention, batch int) (int64, error) {
	where := w.Column + " < now() - make_interval(days => $1)"
	if w.Where != "" {
		where += " AND " + w.Where
	}

	stmt := fmt.Sprintf(`DELETE FROM %[1]s WHERE ctid IN (SELECT ctid FROM %[1]s WHERE %[2]s LIMIT $2)`, w.Table, where)

	var total int64

	for {
		if err := ctx.Err(); err != nil {
			return total, err
		}

		res, err := db.ExecContext(ctx, stmt, w.Days, batch)
		if err != nil {
			return total, fmt.Errorf("prune %s: %w", w.Table, err)
		}

		n, err := res.RowsAffected()
		if err != nil {
			return total, err
		}

		total += n
		if n < int64(batch) {
			return total, nil
		}
	}
}
