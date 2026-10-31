package moderation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/community"
)

// RejectReasons are the reject_reason enum values (SDD 04 section 5).
var RejectReasons = []string{
	"incorrect", "no_source", "bad_source", "duplicate", "policy", "spam", "stale", "target_gone", "other",
}

const maxNote = 1000

// ErrApplyUnavailable means no Applier is wired (routes/admin.go sets community.Applier).
var ErrApplyUnavailable = errors.New("moderation: apply service not available")

// ErrNoAuthor is a clone of a suggestion whose author was deleted.
var ErrNoAuthor = errors.New("moderation: suggestion has no author")

// InvalidError is a 422 on one input field.
type InvalidError struct{ Field, Msg string }

// Error implements error.
func (e *InvalidError) Error() string { return e.Field + ": " + e.Msg }

// Applier applies an approved suggestion to the catalogue (SDD 04 section 7.4 steps 1 to 6). The queue never writes
// the catalogue itself.
type Applier interface {
	Approve(ctx context.Context, id string, version int, opts community.ApproveOptions, actor community.Actor) error
}

// Detail is the moderation page of one suggestion.
type Detail struct {
	ID, Kind, TargetType, TargetID, Domain, Locale, Status, Note string
	RejectReason, DecisionNote                                   string
	SourceURLs                                                   []string
	Version                                                      int
	AuthorID, AssignedTo                                         int64
	SubmittedAt                                                  time.Time
	Age                                                          string
	Competing, Stale, Unreachable                                bool
	Diffs                                                        []FieldDiff
}

// Detail loads one suggestion with its per-field diff against the live catalogue.
func (q Queue) Detail(ctx context.Context, id string) (Detail, error) {
	if _, err := uuid.Parse(id); err != nil {
		return Detail{}, ErrNotFound
	}

	var (
		d                                 Detail
		raw                               []byte
		urls, target, note, reason, dNote sql.NullString
		submitted                         sql.NullTime
	)

	err := q.DB.QueryRowContext(ctx, `
SELECT id::text, kind::text, target_type::text, target_id::text, target_domain, locale, status::text, note,
  reject_reason::text, decision_note, changes, version, COALESCE(author_id, 0), COALESCE(assigned_to, 0), submitted_at,
  COALESCE((meta ->> 'competing')::boolean, false), COALESCE((meta ->> 'stale')::boolean, false),
  COALESCE(meta ->> 'source_status' = 'unreachable', false), array_to_json(source_urls)::text
FROM suggestions WHERE id = $1`, id).Scan(&d.ID, &d.Kind, &d.TargetType, &target, &d.Domain, &d.Locale, &d.Status,
		&note, &reason, &dNote, &raw, &d.Version, &d.AuthorID, &d.AssignedTo, &submitted,
		&d.Competing, &d.Stale, &d.Unreachable, &urls)
	if errors.Is(err, sql.ErrNoRows) {
		return Detail{}, ErrNotFound
	}

	if err != nil {
		return Detail{}, fmt.Errorf("moderation: load suggestion: %w", err)
	}

	d.TargetID, d.Note, d.RejectReason, d.DecisionNote = target.String, note.String, reason.String, dNote.String
	if urls.Valid {
		_ = json.Unmarshal([]byte(urls.String), &d.SourceURLs)
	}

	if submitted.Valid {
		d.SubmittedAt = submitted.Time
		d.Age = AgeBadge(submitted.Time, q.now())
	}

	d.Diffs, err = BuildDiffs(raw, func(field string) json.RawMessage {
		if d.TargetID == "" {
			return nil
		}

		cur, err := community.CurrentValue(ctx, q.DB, d.TargetType, d.TargetID, field)
		if err != nil {
			return nil
		}

		return cur
	})
	if err != nil {
		return Detail{}, err
	}

	return d, nil
}

// Approve hands the decision to the Applier (community.Applier in production).
func (q Queue) Approve(
	ctx context.Context, id string, version int, opts community.ApproveOptions, actor community.Actor,
) error {
	if q.Applier == nil {
		return ErrApplyUnavailable
	}

	err := q.Applier.Approve(ctx, id, version, opts, actor)
	if errors.Is(err, community.ErrNotFound) {
		return ErrNotFound
	}

	return err
}

// Reject records reason and note and moves the suggestion to rejected, all in one transaction.
func (q Queue) Reject(ctx context.Context, id string, version int, reason, note string, actor community.Actor) error {
	if !slices.Contains(RejectReasons, reason) {
		return &InvalidError{Field: "reason", Msg: "Choose a reject reason."}
	}

	if len([]rune(note)) > maxNote {
		return &InvalidError{Field: "note", Msg: "The note is too long."}
	}

	return q.DB.Tx(ctx, func(tx *database.Tx) error {
		if _, err := assignee(ctx, tx, id); err != nil {
			return err
		}

		if err := community.TransitionTx(ctx, tx, id, version, "rejected", actor); err != nil {
			return err
		}

		_, err := tx.ExecContext(ctx, `UPDATE suggestions SET reject_reason = $2, decision_note = NULLIF($3, ''),
decided_by = $4, decided_at = $5, assigned_to = NULL, claimed_at = NULL WHERE id = $1`,
			id, reason, note, actor.ID, q.now())
		if err != nil {
			return fmt.Errorf("moderation: record rejection: %w", err)
		}

		return nil
	})
}

// CloneAsDraft copies a terminal suggestion into a new draft owned by its author, without using quota (SDD 04 7.1).
func (q Queue) CloneAsDraft(ctx context.Context, id string, actor community.Actor) (string, error) {
	if _, err := uuid.Parse(id); err != nil {
		return "", ErrNotFound
	}

	newID := uuid.NewV7().String()

	err := q.DB.Tx(ctx, func(tx *database.Tx) error {
		var (
			status string
			author sql.NullInt64
		)

		err := tx.QueryRowContext(ctx, `SELECT status::text, author_id FROM suggestions WHERE id = $1 FOR UPDATE`, id).
			Scan(&status, &author)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}

		if err != nil {
			return fmt.Errorf("moderation: load suggestion: %w", err)
		}

		if !slices.Contains([]string{"rejected", "reverted", "withdrawn", "expired"}, status) {
			return &community.TransitionError{Status: 409, Code: "invalid-transition", From: status, To: "draft"}
		}

		if !author.Valid {
			return ErrNoAuthor
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO suggestions
(id, author_id, kind, target_type, target_id, target_domain, locale, changes, note, source_urls)
SELECT $2::uuid, author_id, kind, target_type, target_id, target_domain, locale, changes, note, source_urls
FROM suggestions WHERE id = $1`, id, newID); err != nil {
			return fmt.Errorf("moderation: clone suggestion: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO suggestion_events
(suggestion_id, from_status, to_status, actor_id, actor_role, reason, meta)
VALUES ($1::uuid, NULL, 'draft', $2, $3, 'cloned', jsonb_build_object('cloned_from', $4::text))`,
			newID, actor.ID, actor.Role, id); err != nil {
			return fmt.Errorf("moderation: clone event: %w", err)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return newID, nil
}
