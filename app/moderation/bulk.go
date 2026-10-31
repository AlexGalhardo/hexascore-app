package moderation

import (
	"context"
	"errors"

	"github.com/AlexGalhardo/hexascore-app/app/community"
)

// Bulk actions and limits (SDD 04 section 7.4).
const (
	BulkApprove = "approve"
	BulkReject  = "reject"
	BulkClaim   = "claim"
	BulkRelease = "release"

	// MaxBulk is the most ids one bulk request may carry.
	MaxBulk = 50
)

// ErrBulkSize is an empty bulk request or one above MaxBulk (422).
var ErrBulkSize = errors.New("moderation: bulk takes 1 to 50 ids")

// ErrBulkAction is an unknown bulk action (422).
var ErrBulkAction = errors.New("moderation: unknown bulk action")

// BulkResult is the outcome for one id.
type BulkResult struct {
	ID   string
	OK   bool
	Code string
}

// Bulk runs one action over ids, one transaction per suggestion, and reports each outcome. The version used is the
// live one, so a row that changed under the moderator still fails on its own state (competing, already decided).
func (q Queue) Bulk(
	ctx context.Context,
	action string,
	ids []string,
	reason, note string,
	confirmStale bool,
	actor community.Actor,
) ([]BulkResult, error) {
	if len(ids) == 0 || len(ids) > MaxBulk {
		return nil, ErrBulkSize
	}

	switch action {
	case BulkApprove, BulkClaim, BulkRelease:
	case BulkReject:
		if reason == "" {
			return nil, &InvalidError{Field: "reason", Msg: "Choose a reject reason."}
		}
	default:
		return nil, ErrBulkAction
	}

	out := make([]BulkResult, 0, len(ids))

	for _, id := range ids {
		r := BulkResult{ID: id}

		version, err := q.version(ctx, id)
		if err == nil {
			switch action {
			case BulkApprove:
				err = q.Approve(ctx, id, version, confirmStale, actor)
			case BulkReject:
				err = q.Reject(ctx, id, version, reason, note, actor)
			case BulkClaim:
				err = q.Claim(ctx, id, version, actor)
			case BulkRelease:
				err = q.Release(ctx, id, version, actor)
			}
		}

		r.OK, r.Code = err == nil, code(err)
		out = append(out, r)
	}

	return out, nil
}

func (q Queue) version(ctx context.Context, id string) (int, error) {
	var v int

	if err := q.DB.QueryRowContext(ctx, `SELECT version FROM suggestions WHERE id::text = $1`, id).
		Scan(&v); err != nil {
		return 0, ErrNotFound
	}

	return v, nil
}

func code(err error) string {
	var (
		te *community.TransitionError
		ie *InvalidError
	)

	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrNotFound):
		return "not-found"
	case errors.Is(err, ErrNotAssignee):
		return "not-assignee"
	case errors.Is(err, ErrApplyUnavailable):
		return "apply-unavailable"
	case errors.As(err, &te):
		return te.Code
	case errors.As(err, &ie):
		return "invalid-" + ie.Field
	default:
		return "error"
	}
}
