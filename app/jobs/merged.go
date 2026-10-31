package jobs

import (
	"context"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// handleMerged expires the open suggestions on the merged-away item (SDD 04 section 8) and recomputes the
// surviving item after a merge moved the ratings (SDD 02 §7.2); list and status moves (P1.F.1.2) and reindex
// (SDD 07) subscribe separately.
func handleMerged(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, m catalog.Merged) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		q, err := app.Queue()
		if err != nil {
			return err
		}

		if err := community.ExpireTargetGone(ctx, db, q, m.FromID); err != nil {
			return err
		}

		return scoring.RecomputeItem(ctx, db, m.IntoID, app.Now().UTC())
	})
}
