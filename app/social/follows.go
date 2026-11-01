// Package social is the Phase 7 social layer: follows now, the activity feed next (SDD 14 §12).
package social

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/flags"

	"github.com/AlexGalhardo/hexascore-app/app/notifications"
)

// Flag gates every social route and prop (SDD 14 §12.3, env FLAG_SOCIAL). It stays off until the owner answers
// Q-A5 (lawful basis of the feed) and the privacy policy covers it.
var Flag = flags.Bool("social", false).
	Describe("Follows and the activity feed (SDD 14 §12); keep off until Q-A5 is answered")

// Follow states (SDD 14 §12.4).
const (
	StateAccepted = "accepted"
	StatePending  = "pending"
)

// ErrNotFound answers an unknown, anonymised or banned user, or a missing follow request.
func ErrNotFound() error {
	return errs.NewKey(http.StatusNotFound, "social.not_found", "Not found.", nil).WithExtension("code", "not-found")
}

// ErrSelf answers a follow of your own profile.
func ErrSelf() error {
	return errs.NewKey(http.StatusUnprocessableEntity, "social.cannot_follow_self", "You cannot follow yourself.", nil).
		WithExtension("code", "cannot-follow-self")
}

// Follows is the follow graph. Flags nil means the layer is off.
type Follows struct {
	DB    *database.DB
	Flags *flags.Flags
}

// Enabled reports whether the social layer is on for this request.
func (s *Follows) Enabled(ctx context.Context) bool {
	return s != nil && s.Flags != nil && s.Flags.Enabled(ctx, Flag)
}

// target resolves a nickname the way the profile does; banned users cannot be followed.
func target(ctx context.Context, ex database.Executor, nickname string) (id int64, private bool, err error) {
	var banned bool

	err = ex.QueryRowContext(ctx, `SELECT id, profile_private,
  banned_at IS NOT NULL AND (ban_expires_at IS NULL OR ban_expires_at > now()) FROM users
WHERE username = $1 AND anonymized_at IS NULL`, strings.ToLower(nickname)).Scan(&id, &private, &banned)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && banned) {
		return 0, false, ErrNotFound()
	}

	if err != nil {
		return 0, false, fmt.Errorf("social: load user: %w", err)
	}

	return id, private, nil
}

// Follow makes followerID follow nickname: accepted for a public profile, pending for a private one. Repeating it
// keeps the existing row and sends no second notice.
func (s *Follows) Follow(ctx context.Context, followerID int64, nickname string) (string, error) {
	var state string

	err := s.DB.Tx(ctx, func(tx *database.Tx) error {
		id, private, err := target(ctx, tx, nickname)
		if err != nil {
			return err
		}

		if id == followerID {
			return ErrSelf()
		}

		state = StateAccepted
		if private {
			state = StatePending
		}

		res, err := tx.ExecContext(ctx, `INSERT INTO follows (follower_id, followee_id, state, accepted_at)
VALUES ($1, $2, $3, CASE WHEN $3 = 'accepted' THEN now() END) ON CONFLICT DO NOTHING`, followerID, id, state)
		if err != nil {
			return fmt.Errorf("social: follow: %w", err)
		}

		if n, err := res.RowsAffected(); err != nil || n == 0 {
			if err != nil {
				return fmt.Errorf("social: follow: %w", err)
			}

			return tx.QueryRowContext(ctx, `SELECT state FROM follows WHERE follower_id = $1 AND followee_id = $2`,
				followerID, id).Scan(&state)
		}

		kind := notifications.NewFollower
		if state == StatePending {
			kind = notifications.FollowRequest
		}

		var nickname string
		if err := tx.QueryRowContext(ctx, `SELECT username FROM users WHERE id = $1`, followerID).
			Scan(&nickname); err != nil {
			return fmt.Errorf("social: follower nickname: %w", err)
		}

		_, err = notifications.Send(ctx, tx, id, kind, "follow:"+strconv.FormatInt(followerID, 10),
			map[string]any{"nickname": nickname})

		return err
	})

	return state, err
}

// Unfollow removes the follow or the pending request; no row is not an error.
func (s *Follows) Unfollow(ctx context.Context, followerID int64, nickname string) error {
	var id int64

	err := s.DB.QueryRowContext(ctx, `SELECT id FROM users WHERE username = $1 AND anonymized_at IS NULL`,
		strings.ToLower(nickname)).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound()
	}

	if err != nil {
		return fmt.Errorf("social: unfollow: %w", err)
	}

	if _, err := s.DB.ExecContext(ctx, `DELETE FROM follows WHERE follower_id = $1 AND followee_id = $2`,
		followerID, id); err != nil {
		return fmt.Errorf("social: unfollow: %w", err)
	}

	return nil
}

// Answer accepts or declines the pending request that nickname sent to followeeID. Only the followee's own
// requests match, so anyone else gets ErrNotFound.
func (s *Follows) Answer(ctx context.Context, followeeID int64, nickname string, accept bool) error {
	q := `DELETE FROM follows WHERE followee_id = $1 AND state = 'pending'
  AND follower_id = (SELECT id FROM users WHERE username = $2)`
	if accept {
		q = `UPDATE follows SET state = 'accepted', accepted_at = now() WHERE followee_id = $1 AND state = 'pending'
  AND follower_id = (SELECT id FROM users WHERE username = $2)`
	}

	res, err := s.DB.ExecContext(ctx, q, followeeID, strings.ToLower(nickname))
	if err != nil {
		return fmt.Errorf("social: answer request: %w", err)
	}

	if n, err := res.RowsAffected(); err != nil || n == 0 {
		if err != nil {
			return fmt.Errorf("social: answer request: %w", err)
		}

		return ErrNotFound()
	}

	return nil
}

// visible is the other side of an edge that may be counted: public, not banned, not anonymised (SDD 14 §12.5).
const visible = `NOT u.profile_private AND u.anonymized_at IS NULL
  AND NOT (u.banned_at IS NOT NULL AND (u.ban_expires_at IS NULL OR u.ban_expires_at > now()))`

// ProfileProps is the `social` prop of /u/{nickname}: counts unless the profile is hidden from the viewer, and the
// viewer's follow state when signed in and not the owner. ok is false when the layer is off.
func (s *Follows) ProfileProps(
	ctx context.Context, viewerID, userID int64, hidden bool,
) (props map[string]any, ok bool, err error) {
	if !s.Enabled(ctx) {
		return nil, false, nil
	}

	props = map[string]any{}

	if !hidden {
		var followers, following int

		err := s.DB.QueryRowContext(ctx, `SELECT
  (SELECT count(*) FROM follows f JOIN users u ON u.id = f.follower_id
    WHERE f.followee_id = $1 AND f.state = 'accepted' AND `+visible+`),
  (SELECT count(*) FROM follows f JOIN users u ON u.id = f.followee_id
    WHERE f.follower_id = $1 AND f.state = 'accepted' AND `+visible+`)`, userID).Scan(&followers, &following)
		if err != nil {
			return nil, false, fmt.Errorf("social: counts: %w", err)
		}

		props["followers"], props["following"] = followers, following
	}

	if viewerID != 0 && viewerID != userID {
		var state string

		err := s.DB.QueryRowContext(ctx, `SELECT state FROM follows WHERE follower_id = $1 AND followee_id = $2`,
			viewerID, userID).Scan(&state)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, false, fmt.Errorf("social: state: %w", err)
		}

		props["state"] = state
	}

	return props, true, nil
}
