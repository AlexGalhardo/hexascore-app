package jobs

import (
	"context"

	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// handleMerged is a placeholder so `catalog.merged` jobs do not dead-letter
// before their subscribers exist (score recompute P1.D.3.2, reindex SDD 07).
func handleMerged() queue.Registration {
	return queue.Handle(func(context.Context, catalog.Merged) error { return nil })
}
