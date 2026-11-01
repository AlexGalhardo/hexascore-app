package covers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/storage"
)

const (
	// GCJob names the deferred-delete task (SDD 06 §7.6 "covers.gc").
	GCJob = "covers.gc"
	// GCAfter is how long a released object stays in the bucket (ADR 0014 as amended: it replaces versioning).
	GCAfter = 30 * 24 * time.Hour
	// GCBatch bounds one GC call.
	GCBatch = 1000
)

// GC deletes the objects released before now-GCAfter that no item or edition points at again, and forgets the
// released keys that are referenced again. It returns how many objects it deleted.
//
// shortcut: a key re-referenced between the select and the delete loses its object; that needs the same source
// bytes to come back for the same item exactly 30 days after they left, so the next covers run re-uploads it.
func GC(ctx context.Context, db *database.DB, disk storage.Disk, now time.Time) (int, error) {
	if _, err := db.ExecContext(ctx, `DELETE FROM cover_gc g WHERE
  EXISTS (SELECT 1 FROM items i WHERE i.cover_key = g.key OR i.cover_thumb_key = g.key)
  OR EXISTS (SELECT 1 FROM editions e WHERE e.cover_key = g.key OR e.cover_thumb_key = g.key)`); err != nil {
		return 0, fmt.Errorf("covers gc: forget reused keys: %w", err)
	}

	rows, err := db.QueryContext(ctx, `SELECT key FROM cover_gc WHERE released_at <= $1 ORDER BY released_at LIMIT $2`,
		now.Add(-GCAfter), GCBatch)
	if err != nil {
		return 0, fmt.Errorf("covers gc: list: %w", err)
	}

	var keys []string

	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			_ = rows.Close()

			return 0, err
		}

		keys = append(keys, k)
	}

	_ = rows.Close()

	if err := rows.Err(); err != nil {
		return 0, err
	}

	for i, k := range keys {
		if err := disk.Delete(ctx, k); err != nil && !errors.Is(err, storage.ErrNotFound) {
			return i, fmt.Errorf("covers gc: delete %s: %w", k, err)
		}

		if _, err := db.ExecContext(ctx, `DELETE FROM cover_gc WHERE key = $1`, k); err != nil {
			return i + 1, fmt.Errorf("covers gc: forget %s: %w", k, err)
		}
	}

	return len(keys), nil
}
