package search

import (
	"context"
	"fmt"

	"github.com/AlexGalhardo/vukem/database"
)

// MarkDirty queues one item or entity (kind "item" or "entity") for the suggest resync that the minute
// search.sync_suggest job drains (SDD 07 §7.5). Re-queuing moves it to the back of the queue.
func MarkDirty(ctx context.Context, ex database.Executor, kind, refID string) error {
	if _, err := ex.ExecContext(ctx, `INSERT INTO search_dirty (ref_id, kind) VALUES ($1, $2)
ON CONFLICT (ref_id) DO UPDATE SET queued_at = now()`, refID, kind); err != nil {
		return fmt.Errorf("search: mark %s dirty: %w", refID, err)
	}

	return nil
}

// IsDirty reports whether refID waits in the suggest resync queue.
func IsDirty(ctx context.Context, ex database.Executor, refID string) (bool, error) {
	var dirty bool
	if err := ex.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM search_dirty WHERE ref_id = $1)`, refID).
		Scan(&dirty); err != nil {
		return false, fmt.Errorf("search: dirty check %s: %w", refID, err)
	}

	return dirty, nil
}
