package jobs

import (
	"context"
	"fmt"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

func handleItemsChanged(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, j catalog.ItemsChanged) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		return db.Tx(ctx, func(tx *database.Tx) error {
			for _, id := range j.IDs {
				if err := search.MarkDirty(ctx, tx, "item", id); err != nil {
					return fmt.Errorf("catalog.items_changed: reindex %s: %w", id, err)
				}
			}

			return nil
		})
	})
}
