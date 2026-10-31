package moderation

import (
	"context"
	"fmt"
	"iter"
	"strconv"
	"time"
)

// AuditExportHeader is the CSV header of AuditRecord.
var AuditExportHeader = []string{
	"id", "occurred_at", "actor_id", "actor_role", "source", "suggestion_id", "target_type", "target_id", "field",
	"before", "after", "reason", "reverts_audit_id", "reverted",
}

// AuditRecord is one export row as CSV cells, in AuditExportHeader order.
func AuditRecord(r AuditRow) []string {
	return []string{
		strconv.FormatInt(r.ID, 10), r.OccurredAt.UTC().Format(time.RFC3339Nano), strconv.FormatInt(r.ActorID, 10),
		r.ActorRole, r.Source, r.SuggestionID, r.TargetType, r.TargetID, r.Field, r.Before, r.After, r.Reason,
		strconv.FormatInt(r.RevertsAuditID, 10), strconv.FormatBool(r.Reverted),
	}
}

// AuditRows streams every catalog_audit row matching f, newest first, one row at a time (FR-04-15 export): memory
// stays flat whatever the row count. The cursor in f is honoured; there is no page limit.
func (q Queue) AuditRows(ctx context.Context, f AuditFilter) iter.Seq2[AuditRow, error] {
	return func(yield func(AuditRow, error) bool) {
		cond, args, err := auditWhere(f)
		if err != nil {
			yield(AuditRow{}, err)
			return
		}

		rows, err := q.DB.QueryContext(ctx, `SELECT a.id, a.occurred_at, COALESCE(a.actor_id, 0),
  a.actor_role::text, a.source::text, COALESCE(a.suggestion_id::text, ''), a.target_type::text, a.target_id::text,
  a.field, COALESCE(a.before::text, 'null'), COALESCE(a.after::text, 'null'), COALESCE(a.reason, ''),
  COALESCE(a.reverts_audit_id, 0), EXISTS (SELECT 1 FROM catalog_audit r WHERE r.reverts_audit_id = a.id)
FROM catalog_audit a `+cond+` ORDER BY a.id DESC`, args...)
		if err != nil {
			yield(AuditRow{}, fmt.Errorf("moderation: export audit: %w", err))
			return
		}

		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var r AuditRow
			if err := rows.Scan(&r.ID, &r.OccurredAt, &r.ActorID, &r.ActorRole, &r.Source, &r.SuggestionID,
				&r.TargetType, &r.TargetID, &r.Field, &r.Before, &r.After, &r.Reason, &r.RevertsAuditID,
				&r.Reverted); err != nil {
				yield(AuditRow{}, fmt.Errorf("moderation: scan audit: %w", err))
				return
			}

			if !yield(r, nil) {
				return
			}
		}

		if err := rows.Err(); err != nil {
			yield(AuditRow{}, fmt.Errorf("moderation: export audit: %w", err))
		}
	}
}
