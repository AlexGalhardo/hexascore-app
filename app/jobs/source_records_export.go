package jobs

import (
	"cmp"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"
	"github.com/AlexGalhardo/vukem/storage"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/run"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

// ExportSourceHistory is the daily cold-tier job of ADR 0025: it creates the history partitions ahead, writes every
// closed month of source_record_history to the bucket as gzip NDJSON, and drops the partitions that are both
// exported and older than the retention window. Scheduled in routes/schedule.go; dark behind the ingest master flag.
type ExportSourceHistory struct{}

// Kind names the job on the queue.
func (ExportSourceHistory) Kind() string { return "source_records.export" }

// historyExportKey is the object of one month.
func historyExportKey(p store.HistoryPartition) string {
	return "exports/source_record_history/" + p.Month() + ".ndjson.gz"
}

// ExportResult says what a run did.
type ExportResult struct {
	Exported, Dropped []string
}

// RunExportSourceHistory exports and drops the closed months. Months with no rows write no object. An object that
// already exists is not written again, so a rerun after a crash (or a second run the same day) is a no-op, and a
// partition is dropped only when its object exists or it was empty.
func RunExportSourceHistory(
	ctx context.Context,
	db database.Executor,
	disk storage.Disk,
	now time.Time,
) (ExportResult, error) {
	var res ExportResult

	if err := store.EnsureHistoryPartitions(ctx, db, now, 3); err != nil {
		return res, err
	}

	parts, err := store.HistoryPartitions(ctx, db)
	if err != nil {
		return res, err
	}

	cutoff := now.UTC().AddDate(0, 0, -run.HistoryRetention)

	for _, p := range parts {
		if p.To.After(now.UTC()) {
			continue // the month is still open
		}

		var rows int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM `+p.Name).Scan(&rows); err != nil {
			return res, fmt.Errorf("source history %s: %w", p.Month(), err)
		}

		if rows > 0 {
			key := historyExportKey(p)

			ok, err := disk.Exists(ctx, key)
			if err != nil {
				return res, err
			}

			if !ok {
				if err := exportPartition(ctx, db, disk, p, key); err != nil {
					return res, fmt.Errorf("source history %s: %w", p.Month(), err)
				}

				res.Exported = append(res.Exported, p.Month())
			}
		}

		if !p.To.After(cutoff) {
			if err := store.DropHistoryPartition(ctx, db, p); err != nil {
				return res, err
			}

			res.Dropped = append(res.Dropped, p.Month())
		}
	}

	return res, nil
}

func exportPartition(
	ctx context.Context,
	db database.Executor,
	disk storage.Disk,
	p store.HistoryPartition,
	key string,
) error {
	pr, pw := io.Pipe()
	done := make(chan error, 1)

	go func() {
		err := disk.Put(ctx, key, pr)
		_ = pr.CloseWithError(err)
		done <- err
	}()

	err := writeHistoryLines(ctx, db, pw, p)
	_ = pw.CloseWithError(err)

	if perr := <-done; perr != nil {
		return cmp.Or(err, perr)
	}

	return err
}

// writeHistoryLines streams one JSON object per row; p.Name was validated by store.HistoryPartitions.
func writeHistoryLines(ctx context.Context, db database.Executor, w io.Writer, p store.HistoryPartition) error {
	rows, err := db.QueryContext(ctx, `SELECT to_jsonb(h)::text FROM `+p.Name+
		` h ORDER BY superseded_at, source, external_id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	zw := gzip.NewWriter(w)

	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return err
		}

		if _, err := io.WriteString(zw, line+"\n"); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return zw.Close()
}

func handleExportSourceHistory(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, _ ExportSourceHistory) error {
		ff, err := app.Flags()
		if err != nil {
			return err
		}

		if !run.Allowed(ctx, ff, "") {
			return nil
		}

		db, err := app.DB()
		if err != nil {
			return err
		}

		disk, err := idsDisk(app)
		if err != nil {
			return err
		}

		res, err := RunExportSourceHistory(ctx, db, disk, app.Now())
		if err == nil && (len(res.Exported) > 0 || len(res.Dropped) > 0) {
			app.Log.InfoContext(ctx, "source history exported", "exported", res.Exported, "dropped", res.Dropped)
		}

		return err
	}, queue.Retries(2), queue.Timeout(10*time.Minute))
}
