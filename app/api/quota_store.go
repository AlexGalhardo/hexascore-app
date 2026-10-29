package api

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AlexGalhardo/vukem/database"
)

// PGQuotaStore keeps the meter's counters in api_quota_counters, api_credit_balances, api_usage_daily and
// api_keys.last_used_at.
type PGQuotaStore struct{ DB *database.DB }

// Load implements QuotaStore.
func (s PGQuotaStore) Load(ctx context.Context, userID int64, period string) (used, credits int, err error) {
	err = s.DB.QueryRowContext(ctx, `SELECT
  COALESCE((SELECT used FROM api_quota_counters WHERE user_id = $1 AND period = $2), 0),
  COALESCE((SELECT remaining FROM api_credit_balances WHERE user_id = $1), 0)`, userID, period).Scan(&used, &credits)

	return used, credits, err
}

// Save implements QuotaStore. Refunds can make a delta negative; GREATEST keeps the CHECK (used >= 0).
func (s PGQuotaStore) Save(ctx context.Context, batch []UserDelta) error {
	return s.DB.Tx(ctx, func(tx *database.Tx) error {
		for _, d := range batch {
			for period, n := range d.Used {
				if n == 0 {
					continue
				}

				if _, err := tx.ExecContext(ctx, `INSERT INTO api_quota_counters (user_id, period, used)
VALUES ($1, $2, GREATEST($3, 0))
ON CONFLICT (user_id, period) DO UPDATE SET used = GREATEST(api_quota_counters.used + $3, 0)`,
					d.UserID, period, n); err != nil {
					return err
				}
			}

			if d.Credits != 0 {
				if _, err := tx.ExecContext(ctx,
					`UPDATE api_credit_balances SET remaining = GREATEST(remaining - $2, 0) WHERE user_id = $1`,
					d.UserID, d.Credits); err != nil {
					return err
				}
			}

			days := map[string]bool{}
			for day := range d.Count {
				days[day] = true
			}

			for day := range d.Refunded {
				days[day] = true
			}

			for day := range days {
				if _, err := tx.ExecContext(ctx, `INSERT INTO api_usage_daily (user_id, day, count, refunded)
VALUES ($1, $2, GREATEST($3, 0), $4)
ON CONFLICT (user_id, day) DO UPDATE SET count = GREATEST(api_usage_daily.count + $3, 0),
  refunded = api_usage_daily.refunded + $4`,
					d.UserID, day, d.Count[day], d.Refunded[day]); err != nil {
					return err
				}
			}

			for id, at := range d.Keys {
				if _, err := tx.ExecContext(ctx,
					`UPDATE api_keys SET last_used_at = GREATEST(COALESCE(last_used_at, $2), $2) WHERE id = $1`,
					id, at.UTC()); err != nil && !errors.Is(err, sql.ErrNoRows) {
					return err
				}
			}
		}

		return nil
	})
}
