package legacyimport

import (
	"context"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/app/scoring"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// finalise is the last stage (SDD 11 §7.6 step 5): score recompute with the
// nightly job's order (priors, then every item), then the search suggest
// rebuild (search.Rebuild). Every step is idempotent,
// so a rerun after a failure starts over safely. It returns the items
// recomputed.
func finalise(ctx context.Context, db *database.DB, now time.Time) (int64, error) {
	if err := scoring.RebuildPriors(ctx, db, now); err != nil {
		return 0, fmt.Errorf("rebuild priors: %w", err)
	}

	n, err := scoring.RecomputeAll(ctx, db, now)
	if err != nil {
		return 0, fmt.Errorf("recompute scores: %w", err)
	}

	if err := search.Rebuild(ctx, db); err != nil {
		return 0, fmt.Errorf("search rebuild: %w", err)
	}

	return int64(n), nil
}
