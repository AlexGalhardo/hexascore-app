// Package community is the suggestion quota, the contributor tiers, and the hourly maintenance job.
package community

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/storage"
)

const (
	draftTTL = 14 * 24 * time.Hour
	claimTTL = 2 * time.Hour
	blockFor = 30 * 24 * time.Hour
)

// cooldownExceeded is 429 suggestion-cooldown. RetryAfter is the whole seconds until until.
func cooldownExceeded(until, now time.Time) *Problem {
	utc := until.UTC()
	retry := max(int(math.Ceil(utc.Sub(now.UTC()).Seconds())), 1)

	return &Problem{
		Status:        http.StatusTooManyRequests,
		Code:          "suggestion-cooldown",
		Type:          problemBase + "suggestion-cooldown",
		RetryAfter:    retry,
		CooldownUntil: &utc,
	}
}

// rejectCooldown is the submit check. A future cooldown_until refuses the submit and consumes nothing.
func rejectCooldown(ctx context.Context, tx *database.Tx, userID int64, now time.Time) error {
	var until sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT cooldown_until FROM contributor_stats WHERE user_id = $1`, userID).
		Scan(&until); err != nil {
		return fmt.Errorf("community: load cooldown: %w", err)
	}
	if until.Valid && until.Time.After(now) {
		return cooldownExceeded(until.Time, now)
	}

	return nil
}

// Hourly expires drafts idle for 14 days, releases claims idle for 2 hours, and rebuilds tiers.
// An expired draft's uploaded cover (community/covers/…) is deleted. An https cover is not ours.
func Hourly(ctx context.Context, db *database.DB, now time.Time, disk storage.Disk) error {
	if err := expireDrafts(ctx, db, now, disk); err != nil {
		return err
	}
	if err := releaseClaims(ctx, db, now); err != nil {
		return err
	}

	return RebuildStats(ctx, db, now)
}

func expireDrafts(ctx context.Context, db *database.DB, now time.Time, disk storage.Disk) error {
	return db.Tx(ctx, func(tx *database.Tx) error {
		rows, err := tx.QueryContext(ctx, `SELECT id::text, version, changes
FROM suggestions WHERE status = 'draft' AND updated_at <= $1 FOR UPDATE`, now.UTC().Add(-draftTTL))
		if err != nil {
			return fmt.Errorf("community: list expired drafts: %w", err)
		}
		type draft struct {
			id      string
			version int
			changes []byte
		}
		var drafts []draft
		for rows.Next() {
			var d draft
			if err := rows.Scan(&d.id, &d.version, &d.changes); err != nil {
				if closeErr := finishRows(rows); closeErr != nil {
					err = errors.Join(err, closeErr)
				}

				return fmt.Errorf("community: list expired drafts: %w", err)
			}
			drafts = append(drafts, d)
		}
		if err := finishRows(rows); err != nil {
			return fmt.Errorf("community: list expired drafts: %w", err)
		}

		var keys []string
		actor := Actor{Role: "system"}
		for _, d := range drafts {
			if err := TransitionTx(ctx, tx, d.id, d.version, "expired", actor); err != nil {
				return err
			}
			if key := uploadedCoverKey(d.changes); key != "" {
				keys = append(keys, key)
			}
		}
		for _, key := range keys {
			if disk == nil {
				return errors.New("community: cover disk is not configured")
			}
			if err := disk.Delete(ctx, key); err != nil && !errors.Is(err, storage.ErrNotFound) {
				return fmt.Errorf("community: delete cover: %w", err)
			}
		}

		return nil
	})
}

func releaseClaims(ctx context.Context, db *database.DB, now time.Time) error {
	return db.Tx(ctx, func(tx *database.Tx) error {
		rows, err := tx.QueryContext(ctx, `SELECT id::text, version FROM suggestions
WHERE status = 'in_review' AND claimed_at IS NOT NULL AND claimed_at <= $1 FOR UPDATE`, now.UTC().Add(-claimTTL))
		if err != nil {
			return fmt.Errorf("community: list idle claims: %w", err)
		}
		type claim struct {
			id      string
			version int
		}
		var claims []claim
		for rows.Next() {
			var c claim
			if err := rows.Scan(&c.id, &c.version); err != nil {
				if closeErr := finishRows(rows); closeErr != nil {
					err = errors.Join(err, closeErr)
				}

				return fmt.Errorf("community: list idle claims: %w", err)
			}
			claims = append(claims, c)
		}
		if err := finishRows(rows); err != nil {
			return fmt.Errorf("community: list idle claims: %w", err)
		}

		actor := Actor{Role: "system"}
		for _, c := range claims {
			if err := TransitionTx(ctx, tx, c.id, c.version, "submitted", actor); err != nil {
				return err
			}
			if _, err := tx.ExecContext(
				ctx,
				`UPDATE suggestions SET assigned_to = NULL, claimed_at = NULL WHERE id = $1`,
				c.id,
			); err != nil {
				return fmt.Errorf("community: release claim: %w", err)
			}
		}

		return nil
	})
}

// uploadedCoverKey returns the stored object key for a cover change. https values stay; they are not uploads.
func uploadedCoverKey(raw []byte) string {
	var changes map[string]json.RawMessage
	if err := json.Unmarshal(raw, &changes); err != nil {
		return ""
	}
	cover, ok := changes["cover"]
	if !ok {
		return ""
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(cover, &obj); err != nil {
		return ""
	}
	after, ok := obj["after"]
	if !ok {
		return ""
	}
	var key string
	if err := json.Unmarshal(after, &key); err != nil {
		return ""
	}
	if strings.HasPrefix(key, "https://") || strings.HasPrefix(key, "http://") {
		return ""
	}
	if !strings.HasPrefix(key, "community/covers/") || strings.Contains(key, "..") {
		return ""
	}

	return key
}

// RebuildStats rewrites contributor_stats from suggestions. A second run on an already correct table
// changes nothing, including updated_at. Accuracy is the generated Laplace column and is not written.
// trusted_eligible is stored and read by nothing in v1.
func RebuildStats(ctx context.Context, db *database.DB, now time.Time) error {
	return db.Tx(ctx, func(tx *database.Tx) error {
		if err := writeTotals(ctx, tx, now); err != nil {
			return err
		}
		if err := writeTiers(ctx, tx, now); err != nil {
			return err
		}

		return writeCooldowns(ctx, tx, now)
	})
}

func writeTotals(ctx context.Context, tx *database.Tx, now time.Time) error {
	_, err := tx.ExecContext(ctx, `WITH agg AS (
  SELECT author_id AS user_id,
    count(*) FILTER (WHERE submitted_at IS NOT NULL)::int AS submitted_total,
    count(*) FILTER (WHERE status IN ('applied', 'reverted'))::int AS approved_total,
    count(*) FILTER (WHERE status = 'rejected' AND reject_reason IS NOT NULL
      AND reject_reason NOT IN ('duplicate', 'stale', 'target_gone'))::int AS rejected_total,
    count(*) FILTER (WHERE status = 'reverted' AND reject_reason = 'incorrect')::int AS reverted_total,
    count(*) FILTER (WHERE status = 'withdrawn')::int AS withdrawn_total
  FROM suggestions
  WHERE author_id IS NOT NULL
  GROUP BY author_id
)
INSERT INTO contributor_stats (user_id, submitted_total, approved_total, rejected_total, reverted_total, withdrawn_total, updated_at)
SELECT user_id, submitted_total, approved_total, rejected_total, reverted_total, withdrawn_total, $1 FROM agg
ON CONFLICT (user_id) DO UPDATE SET
  submitted_total = EXCLUDED.submitted_total,
  approved_total = EXCLUDED.approved_total,
  rejected_total = EXCLUDED.rejected_total,
  reverted_total = EXCLUDED.reverted_total,
  withdrawn_total = EXCLUDED.withdrawn_total,
  updated_at = EXCLUDED.updated_at
WHERE contributor_stats.submitted_total IS DISTINCT FROM EXCLUDED.submitted_total
  OR contributor_stats.approved_total IS DISTINCT FROM EXCLUDED.approved_total
  OR contributor_stats.rejected_total IS DISTINCT FROM EXCLUDED.rejected_total
  OR contributor_stats.reverted_total IS DISTINCT FROM EXCLUDED.reverted_total
  OR contributor_stats.withdrawn_total IS DISTINCT FROM EXCLUDED.withdrawn_total`, now.UTC())
	if err != nil {
		return fmt.Errorf("community: rebuild totals: %w", err)
	}
	_, err = tx.ExecContext(ctx, `UPDATE contributor_stats s SET
  submitted_total = 0, approved_total = 0, rejected_total = 0, reverted_total = 0, withdrawn_total = 0, updated_at = $1
WHERE NOT EXISTS (SELECT 1 FROM suggestions g WHERE g.author_id = s.user_id)
  AND (s.submitted_total <> 0 OR s.approved_total <> 0 OR s.rejected_total <> 0 OR s.reverted_total <> 0 OR s.withdrawn_total <> 0)`, now.UTC())
	if err != nil {
		return fmt.Errorf("community: rebuild empty totals: %w", err)
	}

	return nil
}

func writeTiers(ctx context.Context, tx *database.Tx, now time.Time) error {
	// A future blocked_until stays, so an admin block (or the 30-day accuracy block) is not cleared early.
	// The accuracy rule sets blocked_until only when it is empty or already past.
	_, err := tx.ExecContext(ctx, `UPDATE contributor_stats AS s SET
  tier = v.tier,
  blocked_until = v.blocked_until,
  trusted_eligible = v.trusted_eligible,
  updated_at = $1
FROM (
  SELECT user_id,
    CASE
      WHEN (approved_total + rejected_total + reverted_total) >= 20 AND accuracy < 0.20 THEN 'blocked'::contributor_tier
      WHEN blocked_until > $1 THEN 'blocked'::contributor_tier
      WHEN approved_total = 0 THEN 'new'::contributor_tier
      WHEN (approved_total + rejected_total + reverted_total) >= 10 AND accuracy < 0.40 THEN 'low'::contributor_tier
      ELSE 'standard'::contributor_tier
    END AS tier,
    CASE
      WHEN (approved_total + rejected_total + reverted_total) >= 20 AND accuracy < 0.20
        THEN COALESCE(CASE WHEN blocked_until > $1 THEN blocked_until END, $1 + interval '30 days')
      WHEN blocked_until > $1 THEN blocked_until
      ELSE NULL
    END AS blocked_until,
    (approved_total >= 50 AND accuracy >= 0.90 AND NOT EXISTS (
      SELECT 1 FROM suggestions g
      WHERE g.author_id = contributor_stats.user_id
        AND g.status = 'rejected' AND g.reject_reason = 'spam'
        AND g.decided_at > $1 - interval '90 days'
    )) AS trusted_eligible
  FROM contributor_stats
) AS v
WHERE s.user_id = v.user_id
  AND (s.tier IS DISTINCT FROM v.tier
    OR s.blocked_until IS DISTINCT FROM v.blocked_until
    OR s.trusted_eligible IS DISTINCT FROM v.trusted_eligible)`, now.UTC())
	if err != nil {
		return fmt.Errorf("community: rebuild tiers: %w", err)
	}

	return nil
}

func writeCooldowns(ctx context.Context, tx *database.Tx, now time.Time) error {
	// Three qualifying rejections inside 24 h cool the author for 48 h from the third.
	// Any spam rejection cools them for 7 days from that decision. The later end wins.
	_, err := tx.ExecContext(ctx, `WITH rej AS (
  SELECT author_id, decided_at, reject_reason::text AS reason
  FROM suggestions
  WHERE status = 'rejected' AND author_id IS NOT NULL AND decided_at IS NOT NULL
    AND reject_reason IN ('incorrect', 'no_source', 'bad_source', 'policy', 'spam')
), marked AS (
  SELECT author_id, decided_at, reason,
    count(*) OVER (
      PARTITION BY author_id ORDER BY decided_at
      RANGE BETWEEN INTERVAL '24 hours' PRECEDING AND CURRENT ROW
    ) AS in_window
  FROM rej
), ends AS (
  SELECT author_id, NULLIF(GREATEST(
    COALESCE(max(decided_at) FILTER (WHERE reason = 'spam') + interval '7 days', '-infinity'::timestamptz),
    COALESCE(max(decided_at) FILTER (WHERE in_window >= 3) + interval '48 hours', '-infinity'::timestamptz)
  ), '-infinity'::timestamptz) AS until
  FROM marked GROUP BY author_id
), next AS (
  SELECT s.user_id, CASE WHEN e.until > $1 THEN e.until ELSE NULL END AS until
  FROM contributor_stats s
  LEFT JOIN ends e ON e.author_id = s.user_id
)
UPDATE contributor_stats AS s SET cooldown_until = n.until, updated_at = $1
FROM next n
WHERE s.user_id = n.user_id AND s.cooldown_until IS DISTINCT FROM n.until`, now.UTC())
	if err != nil {
		return fmt.Errorf("community: rebuild cooldowns: %w", err)
	}

	return nil
}
