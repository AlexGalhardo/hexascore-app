package scoring

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
)

// Daily caps per user and UTC day (SDD 02 §7.6).
const (
	capNewYoung = 20
	capNewOld   = 200
	capEdits    = 100
	youngAge    = 7 * 24 * time.Hour
	// CodeDailyLimit is the problem code of an exceeded daily cap.
	CodeDailyLimit = "rating-daily-limit"
	// CodeItemLimited is the problem code while an item's ratings are limited.
	CodeItemLimited = "item-ratings-limited"
)

// checkDailyCap counts today's rating_events (not ratings, so delete and
// re-create still counts) and answers 429 with Retry-After until 00:00 UTC.
// Two simultaneous writes on different items can overshoot by the number of
// concurrent requests; shortcut: no per-user lock, the 30/min limiter bounds it.
func checkDailyCap(ctx context.Context, tx *database.Tx, userID int64, isNew bool, now time.Time) error {
	day := now.UTC().Truncate(24 * time.Hour)
	next := day.Add(24 * time.Hour)

	var created time.Time
	var news, edits int
	err := tx.QueryRowContext(ctx, `SELECT u.created_at,
  count(*) FILTER (WHERE e.kind = 'rating_created'), count(*) FILTER (WHERE e.kind = 'rating_updated')
FROM users u LEFT JOIN rating_events e ON e.user_id = u.id AND e.at >= $2 AND e.at < $3
WHERE u.id = $1 GROUP BY u.created_at`, userID, day, next).Scan(&created, &news, &edits)
	if err != nil {
		return err
	}

	limit, used := capEdits, edits
	if isNew {
		limit, used = capNewOld, news
		if now.Sub(created) < youngAge {
			limit = capNewYoung
		}
	}
	if used < limit {
		return nil
	}

	return tooMany("ratings.daily_limit", "Daily rating limit reached.", CodeDailyLimit, next.Sub(now))
}

// checkItemLimiter enforces ratings_limited_until for new ratings: only
// accounts at least 30 days old with a status other than "want" may rate.
func checkItemLimiter(
	ctx context.Context, tx *database.Tx, userID int64, itemID string, until sql.NullTime, now time.Time,
) error {
	if !until.Valid || !now.Before(until.Time) {
		return nil
	}

	var ok bool
	err := tx.QueryRowContext(ctx, `SELECT u.created_at <= $3::timestamptz - interval '30 days'
  AND EXISTS (SELECT 1 FROM item_status s WHERE s.user_id = u.id AND s.item_id = $2 AND s.status <> 'want')
FROM users u WHERE u.id = $1`, userID, itemID, now).Scan(&ok)
	if err != nil || ok {
		return err
	}

	return tooMany(
		"ratings.item_limited",
		"Ratings for this title are temporarily limited.",
		CodeItemLimited,
		until.Time.Sub(now),
	)
}

func tooMany(key, msg, code string, retry time.Duration) error {
	secs := max(int(retry.Seconds())+1, 1)

	return errs.NewKey(http.StatusTooManyRequests, key, msg, nil).WithExtension("code", code).
		WithHeader("Retry-After", strconv.Itoa(secs))
}
