package community

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// Submit moves one draft to submitted when the author is eligible and still
// under the UTC-day quota. now is the clock. Success is 303. A refusal
// returns a *Problem and writes nothing.
func Submit(ctx context.Context, db *database.DB, id string, authorID int64, version int, now time.Time) (int, error) {
	err := db.Tx(ctx, func(tx *database.Tx) error {
		return submitTx(ctx, tx, id, authorID, version, now)
	})
	if err != nil {
		var problem *Problem
		if errors.As(err, &problem) {
			return problem.Status, problem
		}

		return 0, err
	}

	return http.StatusSeeOther, nil
}

func submitTx(ctx context.Context, tx *database.Tx, id string, authorID int64, version int, now time.Time) error {
	tier, blockedUntil, err := lockStats(ctx, tx, authorID)
	if err != nil {
		return err
	}
	if err := checkEligibility(ctx, tx, authorID, tier, blockedUntil, now); err != nil {
		return err
	}
	used, err := submittedToday(ctx, tx, authorID, now)
	if err != nil {
		return err
	}
	if used >= dailyLimit {
		return quotaExceeded(now, used)
	}
	if err := guardTarget(ctx, tx, id, now); err != nil {
		return err
	}
	var owner sql.NullInt64
	err = tx.QueryRowContext(ctx, `SELECT author_id FROM suggestions WHERE id = $1 FOR UPDATE`, id).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("community: suggestion %s not found", id)
	}
	if err != nil {
		return fmt.Errorf("community: lock suggestion author: %w", err)
	}
	if !owner.Valid || owner.Int64 != authorID {
		return fmt.Errorf("community: suggestion %s is not owned by %d", id, authorID)
	}
	actor := Actor{ID: sql.NullInt64{Int64: authorID, Valid: true}, Role: "author"}
	if err := TransitionTx(ctx, tx, id, version, "submitted", actor); err != nil {
		return err
	}
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE suggestions SET submitted_at = $1 WHERE id = $2`,
		now.UTC(),
		id,
	); err != nil {
		return fmt.Errorf("community: set submitted_at: %w", err)
	}
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE contributor_stats SET submitted_total = submitted_total + 1, updated_at = $2 WHERE user_id = $1`,
		authorID,
		now.UTC(),
	); err != nil {
		return fmt.Errorf("community: increment submitted_total: %w", err)
	}

	return nil
}

func checkEligibility(
	ctx context.Context,
	tx *database.Tx,
	userID int64,
	tier string,
	blockedUntil sql.NullTime,
	now time.Time,
) error {
	var verified sql.NullTime
	var created time.Time
	var banned, banExpires sql.NullTime
	err := tx.QueryRowContext(ctx, `SELECT email_verified_at, created_at, banned_at, ban_expires_at FROM users WHERE id = $1`, userID).
		Scan(&verified, &created, &banned, &banExpires)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("community: user %d not found", userID)
	}
	if err != nil {
		return fmt.Errorf("community: load author: %w", err)
	}
	if !verified.Valid {
		return notEligible(time.Time{}, false)
	}
	eligibleAt := created.UTC().Add(7 * 24 * time.Hour)
	if now.UTC().Before(eligibleAt) {
		return notEligible(eligibleAt, true)
	}
	if banned.Valid && (!banExpires.Valid || banExpires.Time.After(now)) {
		if banExpires.Valid {
			return notEligible(banExpires.Time, true)
		}

		return notEligible(time.Time{}, false)
	}
	if tier == "blocked" && (!blockedUntil.Valid || blockedUntil.Time.After(now)) {
		if blockedUntil.Valid {
			return notEligible(blockedUntil.Time, true)
		}

		return notEligible(time.Time{}, false)
	}

	return nil
}
