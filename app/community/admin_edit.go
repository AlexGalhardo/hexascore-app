package community

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/app/api/changes"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// AdminEdit writes admin-panel catalogue edits (SDD 04 §7.4 last paragraph): each field in values (suggestion field
// names, JSON values) goes through the same write path as an apply, labelled with the hexascore source, and gets one
// catalog_audit row (source=admin) whose before is read under the target row lock in the same transaction. Fields
// whose value does not change are skipped. It returns the audit ids written.
//
// This replaces the panel BeforeSave re-read of VF-010: orm.Update rewrites every column of the model, which would
// clobber a concurrent apply's provenance, so the panel resources stay read-only and edits use this path.
func (a Applier) AdminEdit(
	ctx context.Context, targetType, targetID string, values map[string]json.RawMessage, actor Actor,
) ([]int64, error) {
	var auditIDs []int64

	err := a.DB.Tx(ctx, func(tx *database.Tx) error {
		auditIDs = nil

		// Same advisory key as submit and apply, so an edit queues behind an apply on the target.
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1::text), hashtext($2::text))`,
			targetType, targetID); err != nil {
			return fmt.Errorf("community: target lock: %w", err)
		}

		itemType, err := lockTarget(ctx, tx, targetType, targetID)
		if err != nil {
			return err
		}

		if err := catalog.EnsureSource(ctx, tx, catalog.SourceHexaScore); err != nil {
			return err
		}

		var fields []string
		for _, field := range slices.Sorted(maps.Keys(values)) {
			before, err := scalarCurrent(ctx, tx, targetType, targetID, field)
			if err != nil {
				return err
			}

			after := values[field]
			if jsonEqual(before, after) {
				continue
			}

			prev, err := catalog.WriteField(ctx, tx, targetType, targetID, field, after, catalog.SourceHexaScore)
			if err != nil {
				return fmt.Errorf("community: admin edit %s: %w", field, err)
			}

			var id int64
			if err := tx.QueryRowContext(
				ctx,
				`INSERT INTO catalog_audit
(actor_id, actor_role, source, target_type, target_id, field, before, after, prev_provenance)
VALUES ($1, 'admin', 'admin', $2, $3, $4, $5::jsonb, $6::jsonb, $7::jsonb) RETURNING id`,
				actor.ID,
				targetType,
				targetID,
				field,
				string(before),
				string(after),
				string(prev),
			).Scan(&id); err != nil {
				return fmt.Errorf("community: audit %s: %w", field, err)
			}

			fields = append(fields, field)
			auditIDs = append(auditIDs, id)
		}

		if len(fields) == 0 {
			return nil
		}

		if targetType == "item" {
			if err := changes.Record(ctx, tx, changes.Change{
				ItemID: targetID, ItemType: itemType, Action: "updated", Fields: apiFields(fields),
			}); err != nil {
				return err
			}
		}

		if err := a.Queue.DispatchTx(ctx, tx, CatalogChanged{
			TargetType: targetType, TargetID: targetID, Fields: fields, AuditIDs: auditIDs,
		}); err != nil {
			return fmt.Errorf("community: dispatch catalog.changed: %w", err)
		}

		return nil
	})

	return auditIDs, err
}
