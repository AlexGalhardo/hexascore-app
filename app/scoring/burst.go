package scoring

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
)

// Burst detector thresholds (SDD 02 §7.5); untuned until the 90-day review (open issue 02-7).
const (
	burstVolume24h  = 30
	burstVolume1h   = 15
	burstVolumeMult = 8
	burstHourMult   = 3
	burstSkewMin    = 20
	burstSkewPoints = 25.0
	burstSkewBefore = 10
	baselineDays    = 28.0
	burstLimitedFor = 48 * time.Hour
	burstUnusualFor = 7 * 24 * time.Hour
	holdWeight      = 0.25
)

// ErrBurstResolved means the burst already has a resolution.
var ErrBurstResolved = errs.NewKey(http.StatusConflict, "scoring.burst_resolved", "The burst is already resolved.", nil)

// detectBurst runs after a rating insert, inside the item-locked
// transaction. On a trigger it opens (or extends) the item burst, limits
// ratings and flags the window low-trust ratings. It returns the id of the
// score_actions row written for a newly opened burst, 0 otherwise.
func detectBurst(ctx context.Context, tx *database.Tx, itemID string, now time.Time) (int64, error) {
	var n1h, n24h, nBase int
	if err := tx.QueryRowContext(ctx, `SELECT
  count(*) FILTER (WHERE at >= $2::timestamptz - interval '1 hour'),
  count(*) FILTER (WHERE at >= $2::timestamptz - interval '24 hours'),
  count(*) FILTER (WHERE at < $2::timestamptz - interval '48 hours')
FROM rating_events WHERE item_id = $1 AND kind = 'rating_created' AND at >= $2::timestamptz - interval '30 days'`,
		itemID, now).Scan(&n1h, &n24h, &nBase); err != nil {
		return 0, err
	}

	b := math.Max(1, float64(nBase)/baselineDays)
	trigger := ""
	switch {
	case float64(n24h) >= math.Max(burstVolume24h, burstVolumeMult*b) ||
		float64(n1h) >= math.Max(burstVolume1h, burstHourMult*b):
		trigger = "volume"
	case n24h >= burstSkewMin:
		skew, err := skewTriggered(ctx, tx, itemID, now)
		if err != nil {
			return 0, err
		}
		if skew {
			trigger = "skew"
		}
	}
	if trigger == "" {
		return 0, nil
	}

	var burstID int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM score_bursts WHERE item_id = $1 AND resolution IS NULL FOR UPDATE`, itemID).
		Scan(&burstID)
	opened := errors.Is(err, sql.ErrNoRows)
	if err != nil && !opened {
		return 0, err
	}

	limitedUntil := now.Add(burstLimitedFor)
	var action int64
	if opened {
		before, err := aggregateJSON(ctx, tx, itemID)
		if err != nil {
			return 0, err
		}
		stats := `{"n1h":` + strconv.Itoa(n1h) + `,"n24h":` + strconv.Itoa(n24h) + `,"baseline":` +
			strconv.FormatFloat(b, 'f', 3, 64) + `}`
		if err := tx.QueryRowContext(ctx, `INSERT INTO score_bursts (item_id, started_at, limited_until, trigger, stats)
VALUES ($1, $2, $3, $4, $5::jsonb) RETURNING id`, itemID, now, limitedUntil, trigger, stats).Scan(&burstID); err != nil {
			return 0, err
		}
		if err := tx.QueryRowContext(ctx, `INSERT INTO score_actions (at, kind, item_id, burst_id, before)
VALUES ($1, 'burst_open', $2, $3, $4::jsonb) RETURNING id`, now, itemID, burstID, before).Scan(&action); err != nil {
			return 0, err
		}
	} else if _, err := tx.ExecContext(ctx, `UPDATE score_bursts SET limited_until = $2 WHERE id = $1`, burstID, limitedUntil); err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE items SET ratings_limited_until = $2, unusual_activity_until = $3 WHERE id = $1`,
		itemID,
		limitedUntil,
		now.Add(burstUnusualFor),
	); err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO rating_flags (user_id, item_id, shadow_weight, reason, burst_id)
SELECT r.user_id, r.item_id, $4::numeric, 'burst_hold', $2::bigint
FROM ratings r JOIN users u ON u.id = r.user_id
LEFT JOIN item_status s ON s.user_id = r.user_id AND s.item_id = r.item_id
WHERE r.item_id = $1 AND r.created_at >= $3::timestamptz - interval '24 hours'
  AND (u.created_at > $3::timestamptz - interval '30 days' OR s.user_id IS NULL)
ON CONFLICT (user_id, item_id) DO NOTHING`, itemID, burstID, now, holdWeight)

	return action, err
}

// skewTriggered compares the mean overall of the last 24 h of inserts with
// the mean before the window, needing 10 earlier ratings. shortcut: unweighted
// means instead of the weighted ones of SDD 02 §7.5, upgrade with the 02-7 tuning.
func skewTriggered(ctx context.Context, tx *database.Tx, itemID string, now time.Time) (bool, error) {
	var win, before sql.NullFloat64
	var nBefore int
	if err := tx.QueryRowContext(ctx, `SELECT
  avg(overall) FILTER (WHERE created_at >= $2::timestamptz - interval '24 hours'),
  avg(overall) FILTER (WHERE created_at < $2::timestamptz - interval '24 hours'),
  count(*) FILTER (WHERE created_at < $2::timestamptz - interval '24 hours')
FROM ratings WHERE item_id = $1`, itemID, now).Scan(&win, &before, &nBefore); err != nil {
		return false, err
	}

	return win.Valid && before.Valid && nBefore >= burstSkewBefore &&
		math.Abs(win.Float64-before.Float64) >= burstSkewPoints, nil
}

func aggregateJSON(ctx context.Context, tx *database.Tx, itemID string) (string, error) {
	var j string
	err := tx.QueryRowContext(ctx, `SELECT jsonb_build_object('score', score, 'rating_count', rating_count,
'rating_weight', rating_weight)::text FROM items WHERE id = $1`, itemID).Scan(&j)

	return j, err
}

// setActionAfter stores the recomputed aggregate on a score_actions row.
func setActionAfter(ctx context.Context, tx *database.Tx, actionID int64, itemID string) error {
	after, err := aggregateJSON(ctx, tx, itemID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE score_actions SET after = $2::jsonb WHERE id = $1`, actionID, after)

	return err
}

// ResolveBurst applies an admin decision to an open burst: approve deletes
// its flags, exclude sets them to shadow 0. The item is recomputed in the
// same transaction and the action is audited with the before/after aggregate.
func (w *Writer) ResolveBurst(ctx context.Context, actor, burstID int64, approve bool) error {
	return w.DB.Tx(ctx, func(tx *database.Tx) error {
		now := w.Now().UTC()

		var itemID string
		err := tx.QueryRowContext(ctx, `SELECT item_id FROM score_bursts WHERE id = $1`, burstID).Scan(&itemID)
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrNotFound
		}
		if err != nil {
			return err
		}

		var typ string
		if err := tx.QueryRowContext(ctx, `SELECT type FROM items WHERE id = $1 FOR UPDATE`, itemID).
			Scan(&typ); err != nil {
			return err
		}

		resolution, kind := "excluded", "burst_exclude"
		if approve {
			resolution, kind = "approved", "burst_approve"
		}

		res, err := tx.ExecContext(ctx, `UPDATE score_bursts SET resolution = $2, resolved_at = $3, resolved_by = $4
WHERE id = $1 AND resolution IS NULL`, burstID, resolution, now, actor)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return ErrBurstResolved
		}

		before, err := aggregateJSON(ctx, tx, itemID)
		if err != nil {
			return err
		}

		if approve {
			_, err = tx.ExecContext(ctx, `DELETE FROM rating_flags WHERE burst_id = $1`, burstID)
		} else {
			_, err = tx.ExecContext(ctx,
				`UPDATE rating_flags SET shadow_weight = 0, reason = 'admin_exclude' WHERE burst_id = $1`, burstID)
		}
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(
			ctx,
			`UPDATE items SET ratings_limited_until = NULL, unusual_activity_until = NULL WHERE id = $1`,
			itemID,
		); err != nil {
			return err
		}

		return finishAction(ctx, tx, kind, actor, itemID, 0, burstID, before, typ, now)
	})
}

// RestoreFlag removes one rating flag (admin restore) and recomputes the item.
func (w *Writer) RestoreFlag(ctx context.Context, actor, userID int64, itemID string) error {
	return w.DB.Tx(ctx, func(tx *database.Tx) error {
		now := w.Now().UTC()

		var typ string
		err := tx.QueryRowContext(ctx, `SELECT type FROM items WHERE id = $1 FOR UPDATE`, itemID).Scan(&typ)
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrNotFound
		}
		if err != nil {
			return err
		}

		before, err := aggregateJSON(ctx, tx, itemID)
		if err != nil {
			return err
		}

		res, err := tx.ExecContext(ctx, `DELETE FROM rating_flags WHERE user_id = $1 AND item_id = $2`, userID, itemID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return errs.ErrNotFound
		}

		return finishAction(ctx, tx, "flag_restore", actor, itemID, userID, 0, before, typ, now)
	})
}

// finishAction recomputes the item and writes the score_actions audit row.
func finishAction(
	ctx context.Context, tx *database.Tx, kind string, actor int64, itemID string, userID, burstID int64,
	before, typ string, now time.Time,
) error {
	if _, err := recomputeItem(ctx, tx, itemID, typ, now); err != nil {
		return err
	}
	after, err := aggregateJSON(ctx, tx, itemID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO score_actions (at, actor, kind, item_id, user_id, burst_id, before, after)
VALUES ($1, $2, $3, $4, NULLIF($5::bigint, 0), NULLIF($6::bigint, 0), $7::jsonb, $8::jsonb)`,
		now, actor, kind, itemID, userID, burstID, before, after)

	return err
}
