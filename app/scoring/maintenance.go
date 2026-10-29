package scoring

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

const (
	// BatchSize is the page of item ids per pass and per queued recompute batch (SDD 02 §6, open issue 02-8).
	BatchSize = 500

	priorMinItems   = 20
	priorMinRatings = 5
	burstReviewDays = 14
	eventRetention  = 180 * 24 * time.Hour
	pruneBatch      = 10000
	nilUUID         = "00000000-0000-0000-0000-000000000000"
)

// RecomputeItem recomputes one item under its row lock, the same order as the write path (SDD 02 §7.2).
// A vanished item (merged away) is not an error.
func RecomputeItem(ctx context.Context, db *database.DB, itemID string, now time.Time) error {
	return db.Tx(ctx, func(tx *database.Tx) error {
		var typ string
		err := tx.QueryRowContext(ctx, `SELECT type FROM items WHERE id = $1 FOR UPDATE`, itemID).Scan(&typ)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		_, err = recomputeItem(ctx, tx, itemID, typ, now)

		return err
	})
}

// RecomputeAll recomputes every item that has ratings or a stale counter, one transaction per item,
// ascending by id (SDD 02 §7.2), and returns how many it touched.
func RecomputeAll(ctx context.Context, db *database.DB, now time.Time) (int, error) {
	total, last := 0, nilUUID
	for {
		ids, err := itemPage(ctx, db, last)
		if err != nil {
			return total, err
		}
		for _, id := range ids {
			if err := RecomputeItem(ctx, db, id, now); err != nil {
				return total, fmt.Errorf("recompute item %s: %w", id, err)
			}
		}
		total += len(ids)
		if len(ids) < BatchSize {
			return total, nil
		}
		last = ids[len(ids)-1]
	}
}

func itemPage(ctx context.Context, db *database.DB, after string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT id::text FROM items WHERE id > $1::uuid
AND (rating_count > 0 OR EXISTS (SELECT 1 FROM ratings r WHERE r.item_id = items.id))
ORDER BY id LIMIT $2`, after, BatchSize)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, rows.Err()
}

// itemTypes are the values of items.type (a text column with a CHECK).
var itemTypes = []string{"game", "movie", "series", "book"}

// RebuildPriors stores C_{t,k} per item type for the active formula version (SDD 02 §7.1): the mean over
// items with N >= 5 of their weighted raw means, or 60 while fewer than 20 items qualify.
func RebuildPriors(ctx context.Context, db *database.DB, now time.Time) error {
	var version int
	if err := db.QueryRowContext(ctx, `SELECT version FROM score_formula_versions
WHERE activated_at <= $1 ORDER BY version DESC LIMIT 1`, now).Scan(&version); err != nil {
		return fmt.Errorf("active formula version: %w", err)
	}

	for _, typ := range itemTypes {
		if err := rebuildPrior(ctx, db, typ, version, now); err != nil {
			return fmt.Errorf("prior %s: %w", typ, err)
		}
	}

	return nil
}

func rebuildPrior(ctx context.Context, db *database.DB, typ string, version int, now time.Time) error {
	var n int
	var m [Categories]sql.NullFloat64
	err := db.QueryRowContext(ctx, weightedRatingsSQL+`
  WHERE i.type = $1
), per AS (
  SELECT item_id, sum(weight * c1) / sum(weight) AS m1, sum(weight * c2) / sum(weight) AS m2,
    sum(weight * c3) / sum(weight) AS m3, sum(weight * c4) / sum(weight) AS m4,
    sum(weight * c5) / sum(weight) AS m5, sum(weight * c6) / sum(weight) AS m6
  FROM w WHERE weight > 0 GROUP BY item_id HAVING count(*) >= `+strconv.Itoa(priorMinRatings)+`
)
SELECT count(*), avg(m1), avg(m2), avg(m3), avg(m4), avg(m5), avg(m6) FROM per`, typ, now).
		Scan(&n, &m[0], &m[1], &m[2], &m[3], &m[4], &m[5])
	if err != nil {
		return err
	}

	prior := [Categories]float64{}
	for i := range prior {
		prior[i] = defaultPrior
		if n >= priorMinItems && m[i].Valid {
			prior[i] = m[i].Float64
		}
	}

	_, err = db.ExecContext(
		ctx,
		`INSERT INTO score_priors (item_type, formula_version, prior_c, items_used, computed_at)
VALUES ($1, $2, $3::numeric[], $4, $5)
ON CONFLICT (item_type, formula_version) DO UPDATE
SET prior_c = EXCLUDED.prior_c, items_used = EXCLUDED.items_used, computed_at = EXCLUDED.computed_at`,
		typ,
		version,
		fmt.Sprintf("{%g,%g,%g,%g,%g,%g}", prior[0], prior[1], prior[2], prior[3], prior[4], prior[5]),
		n,
		now,
	)

	return err
}

// ReleaseBursts resolves bursts left unreviewed for 14 days (SDD 02 §7.5): auto_released when the skew trigger
// no longer holds, else kept_held. It returns the number of bursts resolved.
func ReleaseBursts(ctx context.Context, db *database.DB, now time.Time) (int, error) {
	open, err := openBursts(ctx, db, now)
	if err != nil {
		return 0, err
	}

	resolved := 0
	for _, b := range open {
		ok, err := autoResolve(ctx, db, b.id, b.item, now)
		if err != nil {
			return resolved, fmt.Errorf("burst %d: %w", b.id, err)
		}
		if ok {
			resolved++
		}
	}

	return resolved, nil
}

type burst struct {
	id   int64
	item string
}

func openBursts(ctx context.Context, db *database.DB, now time.Time) ([]burst, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, item_id::text FROM score_bursts
WHERE resolution IS NULL AND started_at < $1::timestamptz - make_interval(days => $2) ORDER BY id`, now, burstReviewDays)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var open []burst
	for rows.Next() {
		var b burst
		if err := rows.Scan(&b.id, &b.item); err != nil {
			return nil, err
		}
		open = append(open, b)
	}

	return open, rows.Err()
}

func autoResolve(ctx context.Context, db *database.DB, burstID int64, itemID string, now time.Time) (bool, error) {
	done := false
	err := db.Tx(ctx, func(tx *database.Tx) error {
		var typ string
		err := tx.QueryRowContext(ctx, `SELECT type FROM items WHERE id = $1 FOR UPDATE`, itemID).Scan(&typ)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}

		// Fresh triggering activity (volume or skew) extends limited_until, so
		// a burst still limited stays open and is retried on a later run
		// (security review, P1.D.3.2).
		var limitedUntil time.Time
		if err := tx.QueryRowContext(ctx, `SELECT limited_until FROM score_bursts WHERE id = $1`, burstID).
			Scan(&limitedUntil); err != nil {
			return err
		}
		if limitedUntil.After(now) {
			return nil
		}

		skew, err := skewTriggered(ctx, tx, itemID, now)
		if err != nil {
			return err
		}
		resolution, kind := "auto_released", "burst_auto_release"
		if skew {
			resolution, kind = "kept_held", "burst_kept_held"
		}

		res, err := tx.ExecContext(ctx, `UPDATE score_bursts SET resolution = $2, resolved_at = $3
WHERE id = $1 AND resolution IS NULL`, burstID, resolution, now)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return nil
		}

		before, err := aggregateJSON(ctx, tx, itemID)
		if err != nil {
			return err
		}
		if !skew {
			if _, err := tx.ExecContext(ctx, `DELETE FROM rating_flags WHERE burst_id = $1`, burstID); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx,
				`UPDATE items SET ratings_limited_until = NULL, unusual_activity_until = NULL
WHERE id = $1 AND (ratings_limited_until IS NULL OR ratings_limited_until <= $2)`,
				itemID, now); err != nil {
				return err
			}
		}
		done = true

		return finishAction(ctx, tx, kind, 0, itemID, 0, burstID, before, typ, now)
	})

	return done, err
}

// PruneRatingEvents deletes rating_events older than 180 days (SDD 02 §5) and returns the rows removed.
func PruneRatingEvents(ctx context.Context, db *database.DB, now time.Time) (int64, error) {
	var total int64
	for {
		res, err := db.ExecContext(ctx, `DELETE FROM rating_events WHERE id IN
(SELECT id FROM rating_events WHERE at < $1 LIMIT $2)`, now.Add(-eventRetention), pruneBatch)
		if err != nil {
			return total, err
		}
		n, _ := res.RowsAffected()
		total += n
		if n < pruneBatch {
			return total, nil
		}
	}
}

// replaceAt swaps the compute-time parameter of weightedRatingsSQL for now(), for queries run without arguments.
func replaceAt(sqlText string) string { return strings.ReplaceAll(sqlText, "$2::timestamptz", "now()") }
