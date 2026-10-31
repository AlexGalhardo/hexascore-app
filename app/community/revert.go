package community

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/app/api/changes"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// Revert modes (SDD 04 section 7.5).
const (
	RevertSingle  = "single"
	RevertCascade = "cascade"
	RevertForce   = "force"
)

// RevertReasons are the reasons an admin gives for a revert; only incorrect counts against the author.
var RevertReasons = []string{"incorrect", "no_source", "bad_source", "duplicate", "policy", "spam", "stale", "other"}

// ErrAuditNotFound is a catalog_audit id that does not exist.
var ErrAuditNotFound = errors.New("community: audit row not found")

// ErrAdminOnly is a revert by a non-admin actor (owner Q-A15: reverts are admin-only).
var ErrAdminOnly = errors.New("community: only admins revert catalogue changes")

// RevertConflictError is a single revert over later changes of the field (409 revert-conflict). LaterAuditIDs is
// empty when the field changed outside the audit log.
type RevertConflictError struct{ LaterAuditIDs []int64 }

// Error implements error.
func (e *RevertConflictError) Error() string {
	ids := make([]string, len(e.LaterAuditIDs))
	for i, id := range e.LaterAuditIDs {
		ids[i] = strconv.FormatInt(id, 10)
	}

	return "revert-conflict: later audit rows " + strings.Join(ids, ", ")
}

type auditRow struct {
	ID                        int64
	Source, TargetType, Field string
	TargetID                  string
	SuggestionID              sql.NullString
	Before, After, Prev       json.RawMessage
	Reverted                  bool
}

// Revert undoes audit row id through the apply write path: value and exact provenance restored, one revert row per
// field, api_changes, catalog.changed and community.notify in one transaction. A row with a suggestion reverts all
// rows of that suggestion together.
func (a Applier) Revert(ctx context.Context, id int64, mode, reason string, actor Actor) error {
	if actor.Role != "admin" {
		return ErrAdminOnly
	}

	if mode != RevertSingle && mode != RevertCascade && mode != RevertForce {
		return &ValidationError{
			Status: http.StatusUnprocessableEntity,
			Field:  "mode",
			Msg:    "Choose single, cascade or force.",
		}
	}

	if !slices.Contains(RevertReasons, reason) {
		return &ValidationError{Status: http.StatusUnprocessableEntity, Field: "reason", Msg: "Choose a revert reason."}
	}

	return a.DB.Tx(ctx, func(tx *database.Tx) error {
		return a.revert(ctx, tx, id, mode, reason, actor)
	})
}

func (a Applier) revert(ctx context.Context, tx *database.Tx, id int64, mode, reason string, actor Actor) error {
	first, err := loadAudit(ctx, tx, `a.id = $1`, id)
	if err != nil {
		return err
	}

	if len(first) == 0 {
		return ErrAuditNotFound
	}

	a0 := first[0]
	if a0.Source == "revert" || a0.Source == "merge" || a0.Source == "create" || a0.Field == "*merge*" {
		return &TransitionError{Status: http.StatusConflict, Code: "not-revertible"}
	}

	if a0.Reverted {
		return &TransitionError{Status: http.StatusConflict, Code: "already-reverted"}
	}

	// Same per-target advisory lock as apply, so a revert and an apply on one target serialise.
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1), hashtext($2))`,
		a0.TargetType, a0.TargetID); err != nil {
		return fmt.Errorf("community: target lock: %w", err)
	}

	itemType, err := lockTarget(ctx, tx, a0.TargetType, a0.TargetID)
	if err != nil {
		return err
	}

	rows := []auditRow{a0}
	if a0.SuggestionID.Valid {
		rows, err = loadAudit(ctx, tx, `a.suggestion_id = $1 AND a.source NOT IN ('revert','merge','create')
  AND NOT EXISTS (SELECT 1 FROM catalog_audit r WHERE r.reverts_audit_id = a.id)`, a0.SuggestionID.String)
		if err != nil {
			return err
		}
	}

	var (
		plan     []auditRow
		conflict []int64
	)

	for _, r := range rows {
		later, err := loadAudit(ctx, tx, `a.target_type = $1 AND a.target_id = $2::uuid AND a.field = $3 AND a.id > $4
  AND a.source NOT IN ('revert','merge') AND NOT EXISTS (SELECT 1 FROM catalog_audit r WHERE r.reverts_audit_id = a.id)`,
			r.TargetType, r.TargetID, r.Field, r.ID)
		if err != nil {
			return err
		}

		switch {
		case len(later) > 0 && mode == RevertSingle:
			for _, l := range later {
				conflict = append(conflict, l.ID)
			}
		case mode == RevertCascade:
			plan = append(plan, later...)
		}

		plan = append(plan, r)
	}

	if len(conflict) > 0 {
		slices.Sort(conflict)

		return &RevertConflictError{LaterAuditIDs: slices.Compact(conflict)}
	}

	// Newest first: each write puts back the value the previous (older) change found.
	slices.SortFunc(plan, func(x, y auditRow) int { return cmp.Compare(y.ID, x.ID) })
	plan = slices.CompactFunc(plan, func(x, y auditRow) bool { return x.ID == y.ID })

	var (
		auditIDs    []int64
		fields      []string
		suggestions []string
	)

	for _, r := range plan {
		rid, err := a.revertRow(ctx, tx, r, mode == RevertForce, reason, actor)
		if err != nil {
			return err
		}

		auditIDs = append(auditIDs, rid)
		if !slices.Contains(fields, r.Field) {
			fields = append(fields, r.Field)
		}

		if r.SuggestionID.Valid && !slices.Contains(suggestions, r.SuggestionID.String) {
			suggestions = append(suggestions, r.SuggestionID.String)
		}
	}

	slices.Sort(fields)

	if a0.TargetType == "item" {
		if err := changes.Record(ctx, tx, changes.Change{
			ItemID: a0.TargetID, ItemType: itemType, Action: "updated", Fields: apiFields(fields),
		}); err != nil {
			return err
		}
	}

	for _, sid := range suggestions {
		if err := a.markReverted(ctx, tx, sid, reason, actor); err != nil {
			return err
		}
	}

	if err := a.Queue.DispatchTx(ctx, tx, CatalogChanged{
		TargetType: a0.TargetType, TargetID: a0.TargetID, Fields: fields, AuditIDs: auditIDs,
	}); err != nil {
		return fmt.Errorf("community: dispatch catalog.changed: %w", err)
	}

	return nil
}

// revertRow writes r.Before back with r's previous provenance and appends the revert row. Without force the live
// value must still be r.After.
func (a Applier) revertRow(
	ctx context.Context, tx *database.Tx, r auditRow, force bool, reason string, actor Actor,
) (int64, error) {
	var current, after, prev json.RawMessage

	if relationFields[r.Field] {
		var err error

		current, after, err = catalog.RestoreRelation(ctx, tx, r.TargetID, r.Field, r.Before, catalog.SourceCommunity)
		if err != nil {
			return 0, fmt.Errorf("community: revert %s: %w", r.Field, err)
		}

		if !force && !relationEqual(current, r.After) {
			return 0, &RevertConflictError{}
		}

		prev = json.RawMessage("null")
	} else {
		var err error

		current, err = scalarCurrent(ctx, tx, r.TargetType, r.TargetID, r.Field)
		if err != nil {
			return 0, err
		}

		if !force && !jsonEqual(current, r.After) {
			return 0, &RevertConflictError{}
		}

		prev, err = catalog.RestoreField(ctx, tx, r.TargetType, r.TargetID, r.Field, r.Before, r.Prev)
		if errors.Is(err, catalog.ErrUnsupportedField) {
			return 0, &TransitionError{Status: http.StatusConflict, Code: "not-revertible", Field: r.Field}
		}

		if err != nil {
			return 0, fmt.Errorf("community: revert %s: %w", r.Field, err)
		}

		after = r.Before
	}

	var id int64
	err := tx.QueryRowContext(ctx, `INSERT INTO catalog_audit
(actor_id, actor_role, source, suggestion_id, target_type, target_id, field, before, after, prev_provenance, reason,
 reverts_audit_id)
VALUES ($1, $2, 'revert', $3, $4, $5, $6, $7::jsonb, $8::jsonb, $9::jsonb, $10, $11) RETURNING id`,
		actor.ID, actor.Role, r.SuggestionID, r.TargetType, r.TargetID, r.Field, string(current), string(after),
		string(prev), reason, r.ID).Scan(&id)

	var s interface{ SQLState() string }
	if errors.As(err, &s) && s.SQLState() == "23505" {
		return 0, &TransitionError{Status: http.StatusConflict, Code: "already-reverted"}
	}

	if err != nil {
		return 0, fmt.Errorf("community: audit revert of %d: %w", r.ID, err)
	}

	return id, nil
}

// markReverted moves an applied suggestion to reverted; reason incorrect counts against its author. A suggestion
// already reverted, or never applied, is left alone.
func (a Applier) markReverted(ctx context.Context, tx *database.Tx, id, reason string, actor Actor) error {
	var (
		status  string
		version int
		author  sql.NullInt64
	)

	if err := tx.QueryRowContext(ctx, `SELECT status::text, version, author_id FROM suggestions WHERE id = $1 FOR UPDATE`,
		id).
		Scan(&status, &version, &author); err != nil {
		return fmt.Errorf("community: lock reverted suggestion: %w", err)
	}

	if status != "applied" {
		return nil
	}

	if err := TransitionTx(ctx, tx, id, version, "reverted", actor); err != nil {
		return err
	}

	if reason == "incorrect" && author.Valid {
		if _, err := tx.ExecContext(ctx, `INSERT INTO contributor_stats (user_id, reverted_total, updated_at)
VALUES ($1, 1, $2) ON CONFLICT (user_id) DO UPDATE SET reverted_total = contributor_stats.reverted_total + 1,
updated_at = EXCLUDED.updated_at`, author.Int64, a.now()); err != nil {
			return fmt.Errorf("community: contributor stats: %w", err)
		}
	}

	if err := a.Queue.DispatchTx(ctx, tx, Notify{SuggestionID: id, State: "reverted"}); err != nil {
		return fmt.Errorf("community: dispatch notify: %w", err)
	}

	return nil
}

// loadAudit reads catalog_audit rows matching where (a constant with bound args), oldest first, in one query.
func loadAudit(ctx context.Context, tx *database.Tx, where string, args ...any) ([]auditRow, error) {
	var raw []byte

	err := tx.QueryRowContext(ctx, `SELECT COALESCE(json_agg(json_build_object(
  'ID', a.id, 'Source', a.source, 'TargetType', a.target_type, 'TargetID', a.target_id, 'Field', a.field,
  'SuggestionID', a.suggestion_id, 'Before', COALESCE(a.before, 'null'::jsonb), 'After', COALESCE(a.after, 'null'::jsonb),
  'Prev', COALESCE(a.prev_provenance, 'null'::jsonb),
  'Reverted', EXISTS (SELECT 1 FROM catalog_audit r WHERE r.reverts_audit_id = a.id)) ORDER BY a.id), '[]')
FROM catalog_audit a WHERE `+where, args...).Scan(&raw)
	if err != nil {
		return nil, fmt.Errorf("community: load audit: %w", err)
	}

	var out []struct {
		auditRow
		SuggestionID *string
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("community: decode audit: %w", err)
	}

	rows := make([]auditRow, len(out))
	for i, o := range out {
		rows[i] = o.auditRow
		if o.SuggestionID != nil {
			rows[i].SuggestionID = sql.NullString{String: *o.SuggestionID, Valid: true}
		}
	}

	return rows, nil
}

// relationEqual compares two relation snapshots by entity and role, ignoring sources (older snapshots lack them).
func relationEqual(x, y json.RawMessage) bool {
	var a, b []catalog.RelationRef
	if json.Unmarshal(x, &a) != nil || json.Unmarshal(y, &b) != nil || len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].ID != b[i].ID || a[i].Role != b[i].Role {
			return false
		}
	}

	return true
}
