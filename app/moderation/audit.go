package moderation

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/uuid"
)

// AuditPageSize is the catalog-audit list page size.
const AuditPageSize = 50

// AuditFilter narrows /admin/catalog-audit. Empty fields do not filter; Before is the keyset cursor (an audit id).
type AuditFilter struct {
	TargetType, TargetID, Field, Source string
	Before                              int64
}

// AuditRow is one catalog_audit entry as listed.
type AuditRow struct {
	ID             int64     `json:"id"`
	OccurredAt     time.Time `json:"occurred_at"`
	ActorID        int64     `json:"actor_id"`
	ActorRole      string    `json:"actor_role"`
	Source         string    `json:"source"`
	SuggestionID   string    `json:"suggestion_id"`
	TargetType     string    `json:"target_type"`
	TargetID       string    `json:"target_id"`
	Field          string    `json:"field"`
	Before         string    `json:"before"`
	After          string    `json:"after"`
	Reason         string    `json:"reason"`
	RevertsAuditID int64     `json:"reverts_audit_id"`
	Reverted       bool      `json:"reverted"`
}

// Revertible is false for rows SDD 04 section 7.5 and 7.6 never revert: reverts, merges, creates, and rows already
// reverted.
func (r AuditRow) Revertible() bool {
	return !r.Reverted && r.Source != "revert" && r.Source != "merge" && r.Source != "create"
}

// AuditPage is one keyset page, newest first; Next is the cursor of the following page or 0.
type AuditPage struct {
	Rows []AuditRow
	Next int64
}

// Audit lists the catalogue audit log, newest first.
func (q Queue) Audit(ctx context.Context, f AuditFilter) (AuditPage, error) {
	var (
		where []string
		args  []any
	)

	arg := func(v any) string {
		args = append(args, v)

		return "$" + strconv.Itoa(len(args))
	}

	if f.TargetType != "" {
		where = append(where, "a.target_type::text = "+arg(f.TargetType))
	}

	if f.TargetID != "" {
		if _, err := uuid.Parse(f.TargetID); err != nil {
			return AuditPage{}, &InvalidError{Field: "target_id", Msg: "The target id must be a uuid."}
		}

		where = append(where, "a.target_id = "+arg(f.TargetID)+"::uuid")
	}

	if f.Field != "" {
		where = append(where, "a.field = "+arg(f.Field))
	}

	if f.Source != "" {
		where = append(where, "a.source::text = "+arg(f.Source))
	}

	if f.Before > 0 {
		where = append(where, "a.id < "+arg(f.Before))
	}

	cond := ""
	if len(where) > 0 {
		cond = "WHERE " + strings.Join(where, " AND ")
	}

	var raw []byte

	err := q.DB.QueryRowContext(ctx, `SELECT COALESCE(json_agg(row_to_json(t) ORDER BY t.id DESC), '[]') FROM (
SELECT a.id, a.occurred_at, COALESCE(a.actor_id, 0) AS actor_id, a.actor_role::text AS actor_role,
  a.source::text AS source, COALESCE(a.suggestion_id::text, '') AS suggestion_id, a.target_type::text AS target_type,
  a.target_id::text AS target_id, a.field, COALESCE(a.before::text, 'null') AS before,
  COALESCE(a.after::text, 'null') AS after, COALESCE(a.reason, '') AS reason,
  COALESCE(a.reverts_audit_id, 0) AS reverts_audit_id,
  EXISTS (SELECT 1 FROM catalog_audit r WHERE r.reverts_audit_id = a.id) AS reverted
FROM catalog_audit a `+cond+` ORDER BY a.id DESC LIMIT `+strconv.Itoa(AuditPageSize+1)+`) t`, args...).Scan(&raw)
	if err != nil {
		return AuditPage{}, fmt.Errorf("moderation: list audit: %w", err)
	}

	var page AuditPage
	if err := json.Unmarshal(raw, &page.Rows); err != nil {
		return AuditPage{}, fmt.Errorf("moderation: decode audit: %w", err)
	}

	if len(page.Rows) > AuditPageSize {
		page.Rows = page.Rows[:AuditPageSize]
		page.Next = page.Rows[AuditPageSize-1].ID
	}

	return page, nil
}
