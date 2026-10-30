package changes

import (
	"context"
	"fmt"
	"time"
)

// Prune deletes api_changes rows older than Retain. A row at exactly now minus Retain stays.
func Prune(ctx context.Context, ex Exec, now time.Time) (int64, error) {
	res, err := ex.ExecContext(ctx, `DELETE FROM api_changes WHERE changed_at < $1`, now.UTC().Add(-Retain))
	if err != nil {
		return 0, fmt.Errorf("api changes: prune: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("api changes: prune count: %w", err)
	}

	return n, nil
}
