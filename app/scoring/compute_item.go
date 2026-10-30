package scoring

import (
	"context"
	"database/sql"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// ComputeItem is the Go path of formula v1 for one stored item: it loads the ratings with their weight
// factors and the active version's inputs at time at, and applies Weight and Compute. It writes nothing;
// `hexascore:cutover-verify` compares it with the SQL recompute (SDD 11 §7.6 step 6).
func ComputeItem(ctx context.Context, tx *database.Tx, itemID string, at time.Time) (Result, error) {
	rows, err := tx.QueryContext(ctx, `SELECT r.c1, r.c2, r.c3, r.c4, r.c5, r.c6, r.created_at, u.created_at,
  u.banned_at IS NOT NULL, coalesce(st.status::text, ''), i.type, t.user_id IS NOT NULL, coalesce(f.shadow_weight, 1),
  i.release_date, i.release_precision, i.release_status
FROM ratings r
JOIN items i ON i.id = r.item_id
JOIN users u ON u.id = r.user_id
LEFT JOIN item_status st ON st.user_id = r.user_id AND st.item_id = r.item_id
LEFT JOIN time_to_beat t ON t.user_id = r.user_id AND t.item_id = r.item_id
LEFT JOIN rating_flags f ON f.user_id = r.user_id AND f.item_id = r.item_id
WHERE r.item_id = $1`, itemID)
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = rows.Close() }()

	var (
		inputs []Rating
		typ    string
	)

	for rows.Next() {
		var (
			r                   Rating
			ratedAt, userSince  time.Time
			banned, ttb         bool
			status, prec, relSt string
			shadow              float64
			release             sql.NullTime
		)

		if err := rows.Scan(&r.C[0], &r.C[1], &r.C[2], &r.C[3], &r.C[4], &r.C[5], &ratedAt, &userSince,
			&banned, &status, &typ, &ttb, &shadow, &release, &prec, &relSt); err != nil {
			return Result{}, err
		}

		var rel *time.Time
		if release.Valid {
			rel = &release.Time
		}

		r.Weight = Weight(Factors{
			AccountAge:    at.Sub(userSince),
			Status:        status,
			IsGame:        typ == "game",
			HasTimeToBeat: ttb,
			Shadow:        shadow,
			Banned:        banned,
			BeforeRelease: RatedBeforeRelease(ratedAt, rel, prec, relSt),
		})
		inputs = append(inputs, r)
	}

	if err := rows.Err(); err != nil {
		return Result{}, err
	}

	if typ == "" {
		if err := tx.QueryRowContext(ctx, `SELECT type FROM items WHERE id = $1`, itemID).Scan(&typ); err != nil {
			return Result{}, err
		}
	}

	p, _, err := loadParams(ctx, tx, typ, at)
	if err != nil {
		return Result{}, err
	}

	return Compute(inputs, p), nil
}
