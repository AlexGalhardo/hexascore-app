package community

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/queue"

	"github.com/AlexGalhardo/hexascore-app/app/api/changes"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/metrics"
)

// ErrNotFound is a suggestion id that does not exist.
var ErrNotFound = errors.New("community: suggestion not found")

// ErrApplyDeadlock is an apply that deadlocked twice (SDD 04 section 8: rolled back, retried once, then shown).
var ErrApplyDeadlock = errors.New("community: apply deadlocked, try again")

// CatalogChanged is the `catalog.changed` job (SDD 04 section 6 "Jobs"): cache purge, reindex and sitemap
// handlers run from it, idempotent by AuditIDs.
type CatalogChanged struct {
	TargetType string   `json:"target_type"`
	TargetID   string   `json:"target_id"`
	Fields     []string `json:"fields"`
	AuditIDs   []int64  `json:"audit_ids"`
}

// Kind names the job on the queue.
func (CatalogChanged) Kind() string { return "catalog.changed" }

// Notify is the `community.notify` job: one author notification per suggestion state (SDD 04 FR-04-14).
type Notify struct {
	SuggestionID string `json:"suggestion_id"`
	State        string `json:"state"`
}

// Kind names the job on the queue.
func (Notify) Kind() string { return "community.notify" }

// Applier runs the SDD 04 section 7.4 apply transaction. It satisfies moderation.Applier.
type Applier struct {
	DB    *database.DB
	Queue *queue.Queue
	Now   func() time.Time

	// failAt lets tests inject a failure at a named step of an attempt ("end": after every write and dispatch).
	failAt func(step string, attempt int) error
}

func (a Applier) now() time.Time {
	if a.Now != nil {
		return a.Now().UTC()
	}

	return time.Now().UTC()
}

// ApproveOptions are the moderator's confirmations on approve.
type ApproveOptions struct {
	// ConfirmStale applies over a live value that changed since the suggestion was made.
	ConfirmStale bool
	// ConfirmTitle is the typed title of the item a merge deletes (SDD 04 section 7.6).
	ConfirmTitle string
}

// Approve applies suggestion id at version. A deadlock is rolled back and retried once. Approved suggestions that
// waited on this one (depends_on) are applied next.
func (a Applier) Approve(ctx context.Context, id string, version int, opts ApproveOptions, actor Actor) error {
	var err error
	for attempt := range 2 {
		err = a.DB.Tx(ctx, func(tx *database.Tx) error {
			return a.apply(ctx, tx, id, version, opts, actor, attempt)
		})
		if !isDeadlock(err) {
			break
		}
	}

	if isDeadlock(err) {
		return fmt.Errorf("%w: %w", ErrApplyDeadlock, err)
	}

	if err != nil {
		return err
	}

	a.unblock(ctx, id, actor)

	return nil
}

// unblock applies the approved suggestions that depended on id, each in its own transaction. The parent is already
// committed: a dependent that fails (stale, locked, its own cover) stays approved in the queue for a moderator, so
// its error is not the parent's.
func (a Applier) unblock(ctx context.Context, id string, actor Actor) {
	var raw []byte
	if err := a.DB.QueryRowContext(ctx, `SELECT COALESCE(json_agg(json_build_object('id', id::text, 'version', version)
ORDER BY id), '[]') FROM suggestions WHERE depends_on = $1 AND status = 'approved'`, id).Scan(&raw); err != nil {
		return
	}

	var deps []struct {
		ID      string `json:"id"`
		Version int    `json:"version"`
	}
	if json.Unmarshal(raw, &deps) != nil {
		return
	}

	for _, d := range deps {
		_ = a.Approve(ctx, d.ID, d.Version, ApproveOptions{}, actor)
	}
}

func isDeadlock(err error) bool {
	var s interface{ SQLState() string }

	return errors.As(err, &s) && s.SQLState() == "40P01"
}

type applyRow struct {
	kind, targetType, targetID, status, domain string
	version                                    int
	author                                     sql.NullInt64
	changes                                    map[string]storedChange
	pendingDep                                 bool
}

func (a Applier) apply(
	ctx context.Context, tx *database.Tx, id string, version int, opts ApproveOptions, actor Actor, attempt int,
) error {
	// The per-target advisory lock (same key as submit) comes before any row lock: two applies on one target queue
	// here instead of deadlocking when each flags the other's suggestion stale.
	if _, err := tx.ExecContext(
		ctx,
		`SELECT pg_advisory_xact_lock(hashtext(target_type::text), hashtext(target_id::text))
FROM suggestions WHERE id = $1 AND target_id IS NOT NULL`,
		id,
	); err != nil {
		return fmt.Errorf("community: target lock: %w", err)
	}

	// Step 1: suggestion lock, status and version.
	s, err := lockForApply(ctx, tx, id)
	if err != nil {
		return err
	}

	if s.version != version {
		return stale(s.status, "applied", s.version)
	}

	if s.status != "submitted" && s.status != "in_review" && s.status != "approved" {
		return &TransitionError{Status: http.StatusConflict, Code: "invalid-transition", From: s.status, To: "applied"}
	}

	fields := slices.Sorted(maps.Keys(s.changes))
	if s.targetID != "" {
		if err := fieldLocked(ctx, tx, s.targetType, s.targetID, fields, a.now()); err != nil {
			return err
		}
	}

	// Covers finish in the SDD 06 pipeline and dependencies apply first: approve now, write later.
	if _, cover := s.changes["cover"]; cover || s.pendingDep {
		if err := approveTx(ctx, tx, id, version, true, actor); err != nil {
			return err
		}

		if err := a.decided(ctx, tx, id, actor); err != nil {
			return err
		}

		if err := a.Queue.DispatchTx(ctx, tx, Notify{SuggestionID: id, State: "approved"}); err != nil {
			return fmt.Errorf("community: dispatch notify: %w", err)
		}

		return a.fail("end", attempt)
	}

	switch s.kind {
	case "create":
		return a.applyCreate(ctx, tx, s, id, fields, actor, attempt)
	case "merge":
		return a.applyMerge(ctx, tx, s, id, opts, actor, attempt)
	}

	// Step 2: target row lock (one row for an edit, so the id order of legacy mergeItems holds trivially).
	itemType, err := lockTarget(ctx, tx, s.targetType, s.targetID)
	if err != nil {
		return err
	}

	if err := catalog.EnsureSource(ctx, tx, catalog.SourceCommunity); err != nil {
		return err
	}

	// Steps 3 and 4: per field, stale check, write, audit row with the real previous value.
	auditIDs := make([]int64, 0, len(fields))
	for _, field := range fields {
		auditID, err := a.applyField(ctx, tx, s, id, field, opts.ConfirmStale, actor, "community")
		if err != nil {
			return err
		}

		auditIDs = append(auditIDs, auditID)
	}

	if s.targetType == "item" {
		if err := changes.Record(ctx, tx, changes.Change{
			ItemID: s.targetID, ItemType: itemType, Action: "updated", Fields: apiFields(fields),
		}); err != nil {
			return err
		}
	}

	// Step 5: rivals flagged stale, then suggestion, events and contributor stats in finish.
	if _, err := tx.ExecContext(ctx, `UPDATE suggestions SET meta = meta || '{"stale": true}'
WHERE target_type = $1 AND target_id = $2 AND id <> $3 AND status IN ('submitted','in_review','approved')
  AND changes ?| $4::text[]`, s.targetType, s.targetID, id, "{"+strings.Join(fields, ",")+"}"); err != nil {
		return fmt.Errorf("community: flag stale rivals: %w", err)
	}

	return a.finish(ctx, tx, s, id, actor, attempt, CatalogChanged{
		TargetType: s.targetType, TargetID: s.targetID, Fields: fields, AuditIDs: auditIDs,
	})
}

// finish is steps 5 and 6 shared by edit, create and merge: suggestion to applied with its events, the decision,
// the author's approved count, and the jobs, which commit with the transaction or not at all.
func (a Applier) finish(
	ctx context.Context, tx *database.Tx, s applyRow, id string, actor Actor, attempt int, job CatalogChanged,
) error {
	if err := approveTx(ctx, tx, id, s.version, false, actor); err != nil {
		return err
	}

	if err := a.decided(ctx, tx, id, actor); err != nil {
		return err
	}

	if s.author.Valid {
		if _, err := tx.ExecContext(ctx, `INSERT INTO contributor_stats (user_id, approved_total, updated_at)
VALUES ($1, 1, $2) ON CONFLICT (user_id) DO UPDATE SET approved_total = contributor_stats.approved_total + 1,
updated_at = EXCLUDED.updated_at`, s.author.Int64, a.now()); err != nil {
			return fmt.Errorf("community: contributor stats: %w", err)
		}
	}

	if err := a.Queue.DispatchTx(ctx, tx, job); err != nil {
		return fmt.Errorf("community: dispatch catalog.changed: %w", err)
	}

	if err := a.Queue.DispatchTx(ctx, tx, Notify{SuggestionID: id, State: "applied"}); err != nil {
		return fmt.Errorf("community: dispatch notify: %w", err)
	}

	return a.fail("end", attempt)
}

func (a Applier) fail(step string, attempt int) error {
	if a.failAt == nil {
		return nil
	}

	return a.failAt(step, attempt)
}

func (a Applier) decided(ctx context.Context, tx *database.Tx, id string, actor Actor) error {
	now := a.now()

	var submitted sql.NullTime
	if err := tx.QueryRowContext(ctx, `UPDATE suggestions SET decided_by = $2, decided_at = $3, assigned_to = NULL,
claimed_at = NULL WHERE id = $1 RETURNING submitted_at`, id, actor.ID, now).Scan(&submitted); err != nil {
		return fmt.Errorf("community: record decision: %w", err)
	}

	tx.AfterCommit(func(ctx context.Context) {
		metrics.RecordSuggestionDecision(ctx, "approved", now.Sub(submitted.Time))
	})

	return nil
}

func (a Applier) applyField(
	ctx context.Context, tx *database.Tx, s applyRow, id, field string, confirmStale bool, actor Actor, source string,
) (int64, error) {
	ch := s.changes[field]

	var before, after, prev json.RawMessage

	if len(ch.Add)+len(ch.Remove) > 0 {
		if s.targetType != "item" {
			return 0, &ValidationError{
				Status: http.StatusUnprocessableEntity,
				Field:  field,
				Msg:    "relations need an item",
			}
		}

		var err error

		before, after, err = catalog.WriteRelation(ctx, tx, s.targetID, field, refs(ch.Add), refs(ch.Remove),
			catalog.SourceCommunity)
		if err != nil {
			return 0, fmt.Errorf("community: apply %s: %w", field, err)
		}

		prev = json.RawMessage("null")
	} else {
		current, err := scalarCurrent(ctx, tx, s.targetType, s.targetID, field)
		if err != nil {
			return 0, err
		}

		if source != "create" && ch.Before != nil && !jsonEqual(ch.Before, current) && !confirmStale {
			return 0, &TransitionError{
				Status: http.StatusConflict, Code: "stale-before", Field: field, Current: string(current),
			}
		}

		after = ch.After
		if after == nil {
			after = json.RawMessage("null")
		}

		prev, err = catalog.WriteField(ctx, tx, s.targetType, s.targetID, field, after, catalog.SourceCommunity)
		if err != nil {
			return 0, fmt.Errorf("community: apply %s: %w", field, err)
		}

		before = current
	}

	if source == "create" {
		before = json.RawMessage("null")
	}

	var auditID int64
	if err := tx.QueryRowContext(ctx, `INSERT INTO catalog_audit
(actor_id, actor_role, source, suggestion_id, target_type, target_id, field, before, after, prev_provenance)
VALUES ($1, $2, $10, $3, $4, $5, $6, $7::jsonb, $8::jsonb, $9::jsonb) RETURNING id`,
		actor.ID, actor.Role, id, s.targetType, s.targetID, field, string(before), string(after), string(prev),
		source).Scan(&auditID); err != nil {
		return 0, fmt.Errorf("community: audit %s: %w", field, err)
	}

	return auditID, nil
}

func refs(ops []RelationOp) []catalog.RelationRef {
	out := make([]catalog.RelationRef, len(ops))
	for i, op := range ops {
		out[i] = catalog.RelationRef{ID: op.ID, Role: op.Role}
	}

	return out
}

func lockForApply(ctx context.Context, tx *database.Tx, id string) (applyRow, error) {
	var (
		s      applyRow
		target sql.NullString
		raw    []byte
	)

	err := tx.QueryRowContext(ctx, `SELECT s.kind::text, s.target_type::text, s.target_id::text, s.status::text,
  s.target_domain, s.version, s.author_id, s.changes,
  COALESCE(d.status::text NOT IN ('applied'), false)
FROM suggestions s LEFT JOIN suggestions d ON d.id = s.depends_on
WHERE s.id = $1 FOR UPDATE OF s`, id).
		Scan(&s.kind, &s.targetType, &target, &s.status, &s.domain, &s.version, &s.author, &raw, &s.pendingDep)
	if errors.Is(err, sql.ErrNoRows) {
		return s, ErrNotFound
	}

	if err != nil {
		return s, fmt.Errorf("community: lock suggestion: %w", err)
	}

	s.targetID = target.String
	if err := json.Unmarshal(raw, &s.changes); err != nil {
		return s, fmt.Errorf("community: decode changes: %w", err)
	}

	return s, nil
}

// lockTarget locks the edited row and returns the item type (empty for editions and entities).
func lockTarget(ctx context.Context, tx *database.Tx, targetType, targetID string) (string, error) {
	table := map[string]string{"item": "items", "edition": "editions", "entity": "entities"}[targetType]
	if table == "" {
		return "", &TransitionError{Status: http.StatusConflict, Code: "invalid-transition", To: "applied"}
	}

	typeCol := "''"
	if table == "items" {
		typeCol = "type"
	}

	var itemType string

	err := tx.QueryRowContext(ctx, `SELECT `+typeCol+` FROM `+table+` WHERE id = $1 FOR UPDATE`, targetID).
		Scan(&itemType)
	if errors.Is(err, sql.ErrNoRows) {
		return "", &TransitionError{Status: http.StatusConflict, Code: "target-gone", To: "applied"}
	}

	if err != nil {
		return "", fmt.Errorf("community: lock target: %w", err)
	}

	return itemType, nil
}

// apiFields maps suggestion fields to the public API names of the change feed (changes.Visible drops the rest).
func apiFields(fields []string) []string {
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		switch {
		case f == "title" || strings.HasPrefix(f, "translations."):
			out = append(out, "title")
		case strings.HasPrefix(f, "synopsis."):
			out = append(out, "synopsis")
		case f == "release_date":
			out = append(out, "releaseDate")
		case f == "external_ids.wikidata":
			out = append(out, "wikidata", "externalIds")
		case strings.HasPrefix(f, "external_ids."):
			out = append(out, "externalIds")
		case relationFields[f]:
			out = append(out, "entities")
		}
	}

	return out
}

var relationFields = map[string]bool{
	"genres": true, "creators": true, "organizations": true, "franchise": true, "platforms": true,
}
