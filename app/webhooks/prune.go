package webhooks

import (
	"context"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// DeliveryRetain is how long webhook_deliveries are kept (SDD 05 §5). A row
// at exactly this age stays.
const DeliveryRetain = 30 * 24 * time.Hour

// PruneDeliveries deletes delivery-log rows older than DeliveryRetain.
// attempted_at is Unix milliseconds.
func PruneDeliveries(ctx context.Context, db *database.DB, now time.Time) (int64, error) {
	if now.IsZero() {
		now = time.Now()
	}

	cutoff := now.UTC().Add(-DeliveryRetain).UnixMilli()
	res, err := db.ExecContext(ctx, `DELETE FROM webhook_deliveries WHERE attempted_at < $1`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("webhook: prune deliveries: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("webhook: prune deliveries count: %w", err)
	}

	return n, nil
}
