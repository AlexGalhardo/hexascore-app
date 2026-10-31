package community

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

const (
	problemBase = "https://hexascore.app/problems/"
	dailyLimit  = 3
)

// Problem is a community refusal the caller maps onto HTTP.
// Type is the RFC 9457 type URL. EligibleAt is set when a wait has an end.
type Problem struct {
	Status     int
	Code       string
	Type       string
	EligibleAt *time.Time
	RetryAfter int
	Limit      int
	Used       int
	ResetAt    time.Time
	ExistingID string
	Field      string
	Current    string
	LockReason string
}

// Error implements error.
func (e *Problem) Error() string {
	if e == nil {
		return ""
	}

	return e.Code
}

func notEligible(at time.Time, known bool) *Problem {
	p := &Problem{
		Status: http.StatusForbidden,
		Code:   "suggestion-not-eligible",
		Type:   problemBase + "suggestion-not-eligible",
	}
	if known {
		utc := at.UTC()
		p.EligibleAt = &utc
	}

	return p
}

func quotaExceeded(now time.Time, used int) *Problem {
	reset := nextUTCMidnight(now)
	retry := int(math.Ceil(reset.Sub(now.UTC()).Seconds()))
	if retry < 1 {
		retry = 1
	}

	return &Problem{
		Status:     http.StatusTooManyRequests,
		Code:       "suggestion-quota",
		Type:       problemBase + "suggestion-quota",
		RetryAfter: retry,
		Limit:      dailyLimit,
		Used:       used,
		ResetAt:    reset,
	}
}

func nextUTCMidnight(now time.Time) time.Time {
	u := now.UTC()

	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
}

// lockStats inserts the stats row when missing, then locks it. Concurrent
// submits of one user take this lock before any suggestion row.
func lockStats(ctx context.Context, tx *database.Tx, userID int64) (string, sql.NullTime, error) {
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO contributor_stats (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING`,
		userID,
	); err != nil {
		return "", sql.NullTime{}, fmt.Errorf("community: insert contributor stats: %w", err)
	}
	var tier string
	var blocked sql.NullTime
	err := tx.QueryRowContext(ctx, `SELECT tier::text, blocked_until FROM contributor_stats WHERE user_id = $1 FOR UPDATE`, userID).
		Scan(&tier, &blocked)
	if err != nil {
		return "", sql.NullTime{}, fmt.Errorf("community: lock contributor stats: %w", err)
	}

	return tier, blocked, nil
}

func submittedToday(ctx context.Context, tx *database.Tx, userID int64, now time.Time) (int, error) {
	var n int
	err := tx.QueryRowContext(ctx, `
SELECT count(*) FROM suggestions
WHERE author_id = $1 AND submitted_day = ($2 AT TIME ZONE 'UTC')::date`, userID, now.UTC()).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("community: count daily submits: %w", err)
	}

	return n, nil
}
