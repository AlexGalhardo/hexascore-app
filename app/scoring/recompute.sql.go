package scoring

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// recomputeSQL is formula v1 (SDD 02 §7.1) in one statement: $1 item id,
// $2 compute time, $3 m_t, $4 priors C_{t,1..6} as numeric[]. It must stay
// equal to Compute and Weight (TestSQLMatchesGo).
const recomputeSQL = `
WITH w AS (
  SELECT r.c1, r.c2, r.c3, r.c4, r.c5, r.c6, r.overall,
    (CASE WHEN $2::timestamptz - u.created_at < interval '7 days' THEN 0.25
          WHEN $2::timestamptz - u.created_at < interval '30 days' THEN 0.5
          WHEN $2::timestamptz - u.created_at < interval '90 days' THEN 0.75 ELSE 1 END)
    * (CASE st.status WHEN 'in_progress' THEN 0.85 WHEN 'dropped' THEN 0.85
          WHEN 'completed' THEN CASE WHEN i.type = 'game' AND t.user_id IS NULL THEN 0.9 ELSE 1 END
          ELSE 0.7 END)
    * COALESCE(f.shadow_weight, 1)
    * (CASE WHEN u.banned_at IS NOT NULL THEN 0 ELSE 1 END)
    * (CASE
        WHEN i.release_date IS NULL THEN CASE WHEN i.release_status IN ('announced','cancelled') THEN 0 ELSE 1 END
        WHEN (r.created_at AT TIME ZONE 'America/Sao_Paulo')::date < (CASE i.release_precision
            WHEN 'month' THEN (date_trunc('month', i.release_date) + interval '1 month - 1 day')::date
            WHEN 'year' THEN make_date(extract(year FROM i.release_date)::int, 12, 31)
            ELSE i.release_date END) THEN 0
        ELSE 1 END) AS weight
  FROM ratings r
  JOIN items i ON i.id = r.item_id
  JOIN users u ON u.id = r.user_id
  LEFT JOIN item_status st ON st.user_id = r.user_id AND st.item_id = r.item_id
  LEFT JOIN time_to_beat t ON t.user_id = r.user_id AND t.item_id = r.item_id
  LEFT JOIN rating_flags f ON f.user_id = r.user_id AND f.item_id = r.item_id
  WHERE r.item_id = $1
), agg AS (
  SELECT count(*) FILTER (WHERE weight > 0) AS n,
    COALESCE(sum(weight) FILTER (WHERE weight > 0), 0) AS w,
    COALESCE(sum(weight * c1), 0) AS s1, COALESCE(sum(weight * c2), 0) AS s2,
    COALESCE(sum(weight * c3), 0) AS s3, COALESCE(sum(weight * c4), 0) AS s4,
    COALESCE(sum(weight * c5), 0) AS s5, COALESCE(sum(weight * c6), 0) AS s6,
    COALESCE(sum(weight * overall), 0) AS swo
  FROM w
), k AS (
  SELECT n, w, swo,
    (s1 + $3 * ($4::numeric[])[1]) / (w + $3) AS b1, (s2 + $3 * ($4::numeric[])[2]) / (w + $3) AS b2,
    (s3 + $3 * ($4::numeric[])[3]) / (w + $3) AS b3, (s4 + $3 * ($4::numeric[])[4]) / (w + $3) AS b4,
    (s5 + $3 * ($4::numeric[])[5]) / (w + $3) AS b5, (s6 + $3 * ($4::numeric[])[6]) / (w + $3) AS b6
  FROM agg
), se AS (
  SELECT CASE WHEN k.w > 0 THEN sqrt(COALESCE(sum(x.weight * power(x.overall - k.swo / NULLIF(k.w, 0), 2)), 0) / NULLIF(k.w, 0)) / sqrt(NULLIF(k.w, 0)) END AS se
  FROM k LEFT JOIN w x ON x.weight > 0
  GROUP BY k.w, k.swo
)
UPDATE items SET
  rating_count = k.n,
  rating_weight = k.w,
  score = CASE WHEN k.n > 0 THEN (k.b1 + k.b2 + k.b3 + k.b4 + k.b5 + k.b6) / 6 END,
  score_c1 = CASE WHEN k.n > 0 THEN k.b1 END, score_c2 = CASE WHEN k.n > 0 THEN k.b2 END,
  score_c3 = CASE WHEN k.n > 0 THEN k.b3 END, score_c4 = CASE WHEN k.n > 0 THEN k.b4 END,
  score_c5 = CASE WHEN k.n > 0 THEN k.b5 END, score_c6 = CASE WHEN k.n > 0 THEN k.b6 END,
  score_se = se.se,
  score_confidence = CASE WHEN k.n < 5 THEN NULL
    WHEN k.n >= 50 AND se.se <= 2.0 THEN 'high'
    WHEN k.n >= 15 AND se.se <= 4.0 THEN 'medium' ELSE 'low' END,
  score_computed_at = $2
FROM k, se
WHERE items.id = $1`

// Executor runs a statement; *sql.DB and *database.Tx satisfy it.
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Recompute stores formula v1 for one item, computed in SQL.
func Recompute(ctx context.Context, db Executor, itemID string, at time.Time, p Params) error {
	prior := fmt.Sprintf("{%g,%g,%g,%g,%g,%g}", p.Prior[0], p.Prior[1], p.Prior[2], p.Prior[3], p.Prior[4], p.Prior[5])
	if _, err := db.ExecContext(ctx, recomputeSQL, itemID, at, p.M, prior); err != nil {
		return fmt.Errorf("recompute item %s: %w", itemID, err)
	}
	return nil
}
