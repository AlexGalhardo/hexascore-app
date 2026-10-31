package community

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/AlexGalhardo/vukem/database"
)

// Statuses are the suggestion_status values (SDD 04 §7.1). Nine states, so
// every ordered pair is the 81-pair matrix FR-04-4 asks for.
var Statuses = []string{
	"draft", "submitted", "in_review", "approved", "applied",
	"rejected", "reverted", "withdrawn", "expired",
}

// TransitionError is a 409 from the suggestion state machine.
// Code is invalid-transition (From, To) or stale-before (Field, Current).
type TransitionError struct {
	Status  int
	Code    string
	From    string
	To      string
	Field   string
	Current string
}

// Error implements error.
func (e *TransitionError) Error() string {
	if e == nil {
		return ""
	}
	if e.Code == "stale-before" {
		return "stale-before: " + e.Field
	}

	return "invalid-transition: " + e.From + " -> " + e.To
}

// Actor is who performs a transition. Role is author, admin, moderator or system.
type Actor struct {
	ID   sql.NullInt64
	Role string
}

// allowed is SDD 04 §7.1. ADR 0010's stale → expired edge is not here:
// staleness is a badge, and expired is only target_gone from submitted or in_review.
var allowed = map[[2]string]struct{}{
	{"draft", "submitted"}:     {},
	{"draft", "withdrawn"}:     {},
	{"draft", "expired"}:       {},
	{"submitted", "in_review"}: {},
	{"in_review", "submitted"}: {},
	{"submitted", "approved"}:  {},
	{"in_review", "approved"}:  {},
	{"submitted", "applied"}:   {},
	{"in_review", "applied"}:   {},
	{"approved", "applied"}:    {},
	{"approved", "in_review"}:  {},
	{"submitted", "rejected"}:  {},
	{"in_review", "rejected"}:  {},
	{"submitted", "withdrawn"}: {},
	{"in_review", "withdrawn"}: {},
	{"submitted", "expired"}:   {},
	{"in_review", "expired"}:   {},
	{"applied", "reverted"}:    {},
}

func transitionAllowed(from, to string) bool {
	_, ok := allowed[[2]string{from, to}]

	return ok
}

type rowQuery interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Transition moves one suggestion along a single §7.1 edge and writes one
// suggestion_events row in the same transaction. version is the caller's
// optimistic lock.
func Transition(ctx context.Context, db *database.DB, id string, version int, to string, actor Actor) error {
	return db.Tx(ctx, func(tx *database.Tx) error {
		return TransitionTx(ctx, tx, id, version, to, actor)
	})
}

// TransitionTx is Transition on a transaction the caller owns. A rollback
// of that transaction undoes the status change and the event together.
func TransitionTx(ctx context.Context, tx *database.Tx, id string, version int, to string, actor Actor) error {
	return transition(ctx, tx, id, version, to, actor)
}

// Approve records the moderator decision. deferred keeps the suggestion
// approved (cover processing or an unapplied depends_on). Otherwise the
// same transaction records approved and then applied.
func Approve(ctx context.Context, db *database.DB, id string, version int, deferred bool, actor Actor) error {
	return db.Tx(ctx, func(tx *database.Tx) error {
		return approveTx(ctx, tx, id, version, deferred, actor)
	})
}

func transition(ctx context.Context, q rowQuery, id string, version int, to string, actor Actor) error {
	from, current, err := lockSuggestion(ctx, q, id)
	if err != nil {
		return err
	}
	if current != version {
		return stale(from, to, current)
	}
	if !transitionAllowed(from, to) {
		return &TransitionError{Status: 409, Code: "invalid-transition", From: from, To: to}
	}

	return writeStep(ctx, q, id, from, to, actor)
}

func approveTx(ctx context.Context, q rowQuery, id string, version int, deferred bool, actor Actor) error {
	from, current, err := lockSuggestion(ctx, q, id)
	if err != nil {
		return err
	}
	if current != version {
		return stale(from, "approved", current)
	}
	if !deferred && (from == "submitted" || from == "in_review") {
		if err := writeStep(ctx, q, id, from, "approved", actor); err != nil {
			return err
		}

		return writeStep(ctx, q, id, "approved", "applied", actor)
	}
	to := "approved"
	if !deferred {
		to = "applied"
	}
	if !transitionAllowed(from, to) {
		return &TransitionError{Status: 409, Code: "invalid-transition", From: from, To: to}
	}

	return writeStep(ctx, q, id, from, to, actor)
}

func lockSuggestion(ctx context.Context, q rowQuery, id string) (string, int, error) {
	var status string
	var version int
	err := q.QueryRowContext(ctx, `SELECT status, version FROM suggestions WHERE id = $1 FOR UPDATE`, id).
		Scan(&status, &version)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, fmt.Errorf("community: suggestion %s not found", id)
	}
	if err != nil {
		return "", 0, fmt.Errorf("community: lock suggestion: %w", err)
	}

	return status, version, nil
}

func writeStep(ctx context.Context, q rowQuery, id, from, to string, actor Actor) error {
	res, err := q.ExecContext(
		ctx,
		`UPDATE suggestions SET status = $1, version = version + 1, updated_at = now() WHERE id = $2`,
		to,
		id,
	)
	if err != nil {
		return fmt.Errorf("community: update suggestion: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("community: update suggestion: %w", err)
	}
	if n != 1 {
		return fmt.Errorf("community: suggestion %s not updated", id)
	}
	_, err = q.ExecContext(
		ctx,
		`INSERT INTO suggestion_events (suggestion_id, from_status, to_status, actor_id, actor_role)
VALUES ($1, $2, $3, $4, $5)`,
		id,
		from,
		to,
		actor.ID,
		actor.Role,
	)
	if err != nil {
		return fmt.Errorf("community: insert suggestion event: %w", err)
	}

	return nil
}

func stale(from, to string, version int) *TransitionError {
	return &TransitionError{
		Status:  409,
		Code:    "stale-before",
		From:    from,
		To:      to,
		Field:   "version",
		Current: strconv.Itoa(version),
	}
}
