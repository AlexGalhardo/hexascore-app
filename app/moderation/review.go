package moderation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/app/community"
)

// ErrAdminOnly is an apply or final reject by a non-admin actor (owner Q-A15, SDD 04 FR-04-17). The route gate
// refuses it first; the service refuses it again so no other caller can skip the rule.
var ErrAdminOnly = errors.New("moderation: only admins apply or reject suggestions")

// Recommendations a moderator can record.
const (
	RecommendApprove = "approve"
	RecommendReject  = "reject"
)

// Review is one reviewer's recommendation on a suggestion.
type Review struct {
	ReviewerID     int64     `json:"reviewer_id"`
	Recommendation string    `json:"recommendation"`
	Note           string    `json:"note"`
	Version        int       `json:"version"`
	At             time.Time `json:"at"`
}

func requireAdmin(actor community.Actor) error {
	if actor.Role != "admin" {
		return ErrAdminOnly
	}

	return nil
}

// Recommend records the actor's recommendation on an open suggestion, replacing their previous one. It is not a
// state transition: the suggestion's version does not move, so it never makes an admin's open form stale.
func (q Queue) Recommend(
	ctx context.Context, id string, version int, recommendation, note string, actor community.Actor,
) error {
	if recommendation != RecommendApprove && recommendation != RecommendReject {
		return &InvalidError{Field: "recommendation", Msg: "Choose approve or reject."}
	}

	if len([]rune(note)) > maxNote {
		return &InvalidError{Field: "note", Msg: "The note is too long."}
	}

	return q.DB.Tx(ctx, func(tx *database.Tx) error {
		var (
			status  string
			current int
		)

		err := tx.QueryRowContext(ctx, `SELECT status::text, version FROM suggestions WHERE id = $1 FOR SHARE`, id).
			Scan(&status, &current)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}

		if err != nil {
			return fmt.Errorf("moderation: load suggestion: %w", err)
		}

		if status != "submitted" && status != "in_review" {
			return &community.TransitionError{
				Status: http.StatusConflict, Code: "invalid-transition", From: status, To: status,
			}
		}

		if current != version {
			return &community.TransitionError{Status: http.StatusConflict, Code: "stale-before", Field: "version"}
		}

		_, err = tx.ExecContext(ctx, `
INSERT INTO suggestion_reviews (suggestion_id, reviewer_id, recommendation, note, suggestion_version, created_at, updated_at)
VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $6)
ON CONFLICT (suggestion_id, reviewer_id) DO UPDATE SET recommendation = EXCLUDED.recommendation,
  note = EXCLUDED.note, suggestion_version = EXCLUDED.suggestion_version, updated_at = EXCLUDED.updated_at`,
			id, actor.ID, recommendation, note, version, q.now())
		if err != nil {
			return fmt.Errorf("moderation: record recommendation: %w", err)
		}

		return nil
	})
}

// reviews loads a suggestion's recommendations, newest first, in one query.
func (q Queue) reviews(ctx context.Context, id string) ([]Review, error) {
	var raw []byte

	err := q.DB.QueryRowContext(ctx, `
SELECT COALESCE(json_agg(json_build_object('reviewer_id', reviewer_id, 'recommendation', recommendation,
  'note', COALESCE(note, ''), 'version', suggestion_version, 'at', updated_at) ORDER BY updated_at DESC), '[]')
FROM suggestion_reviews WHERE suggestion_id = $1`, id).Scan(&raw)
	if err != nil {
		return nil, fmt.Errorf("moderation: load reviews: %w", err)
	}

	var out []Review
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("moderation: decode reviews: %w", err)
	}

	return out, nil
}
