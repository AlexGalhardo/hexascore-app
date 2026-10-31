package jobs

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/community"
)

func handleCatalogChanged(app *vukem.App) queue.Registration {
	return queue.Handle(func(ctx context.Context, j community.CatalogChanged) error {
		db, err := app.DB()
		if err != nil {
			return err
		}

		_, err = RunCatalogChanged(ctx, db, j, app.Now().UTC())

		return err
	})
}

// RunCatalogChanged runs the side effects of one `catalog.changed` job (SDD 04 section 6 "Jobs") and reports whether
// any audit id was new. In one transaction it marks the audit ids done, queues the search reindex of the owning item
// or entity (SDD 07 search_dirty) and moves its updated_at, the sitemap lastmod. A replay of the same ids does nothing.
// The api_changes row was written by the apply transaction (SDD 05 section 7.3). Page caches are query-keyed with
// TTLs of a minute or less (render cache, search, home strip), so there is no per-item entry to purge.
func RunCatalogChanged(ctx context.Context, db *database.DB, j community.CatalogChanged, now time.Time) (bool, error) {
	ids := make([]string, len(j.AuditIDs))
	for i, id := range j.AuditIDs {
		ids[i] = strconv.FormatInt(id, 10)
	}

	ran := false

	err := db.Tx(ctx, func(tx *database.Tx) error {
		var fresh int
		if err := tx.QueryRowContext(ctx, `WITH ins AS (
  INSERT INTO catalog_changed_done (audit_id) SELECT unnest($1::bigint[]) ON CONFLICT (audit_id) DO NOTHING RETURNING 1
) SELECT count(*) FROM ins`, "{"+strings.Join(ids, ",")+"}").Scan(&fresh); err != nil {
			return fmt.Errorf("catalog.changed: mark audit ids: %w", err)
		}

		if fresh == 0 {
			return nil
		}

		ran = true

		var q string

		switch j.TargetType {
		case "item":
			q = `WITH t AS (UPDATE items SET updated_at = $2 WHERE id = $1 RETURNING id)
INSERT INTO search_dirty (ref_id, kind) SELECT id, 'item' FROM t ON CONFLICT (ref_id) DO UPDATE SET queued_at = now()`
		case "edition":
			q = `WITH t AS (UPDATE items SET updated_at = $2 WHERE id = (SELECT work_id FROM editions WHERE id = $1) RETURNING id)
INSERT INTO search_dirty (ref_id, kind) SELECT id, 'item' FROM t ON CONFLICT (ref_id) DO UPDATE SET queued_at = now()`
		case "entity":
			q = `WITH t AS (UPDATE entities SET updated_at = $2 WHERE id = $1 RETURNING id)
INSERT INTO search_dirty (ref_id, kind) SELECT id, 'entity' FROM t ON CONFLICT (ref_id) DO UPDATE SET queued_at = now()`
		default:
			return fmt.Errorf("catalog.changed: target type %q", j.TargetType)
		}

		if _, err := tx.ExecContext(ctx, q, j.TargetID, now); err != nil {
			return fmt.Errorf("catalog.changed: reindex %s: %w", j.TargetID, err)
		}

		return nil
	})

	return ran, err
}
