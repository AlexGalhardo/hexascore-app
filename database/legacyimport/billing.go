package legacyimport

import (
	"context"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// billingSQL copies one batch of legacy users' billing rows (SDD 05 §5, SDD 11
// §7.6 step 5). Each statement takes $1 < legacy user id <= $2 (and the import
// time as $3 when it has one), maps the owner through users.legacy_id and is
// idempotent.
var billingSQL = []userSQL{
	{"stripe_customers", `UPDATE users n SET stripe_customer_id = c.customer_id, updated_at = $3
FROM legacy_import.stripe_customers c WHERE n.legacy_id = c.user_id AND c.user_id > $1 AND c.user_id <= $2`, true},
	// Keys stay sha256(full key), so existing keys keep authenticating.
	{"api_keys", `INSERT INTO api_keys (id, user_id, name, prefix, hash, last_used_at, revoked_at, created_at)
SELECT k.id, n.id, k.name, k.prefix, k.hash, k.last_used_at, k.revoked_at, k.created_at
FROM legacy_import.api_keys k JOIN users n ON n.legacy_id = k.user_id
WHERE k.user_id > $1 AND k.user_id <= $2
ON CONFLICT (id) DO NOTHING`, false},
	// Legacy keeps one row per user; free rows carry nothing to import.
	{
		"subscriptions",
		`INSERT INTO subscriptions (user_id, stripe_subscription_id, status, price_id, plan, provider, state,
  current_period_end, cancel_at_period_end, created_at, updated_at)
SELECT n.id, coalesce(s.provider_ref, 'legacy:' || s.user_id), s.status::text, '', s.plan::text, s.provider::text,
  s.status::text, s.current_period_end, s.cancel_at_period_end, s.updated_at, $3
FROM legacy_import.subscriptions s JOIN users n ON n.legacy_id = s.user_id
WHERE s.user_id > $1 AND s.user_id <= $2 AND s.plan::text <> 'free'
ON CONFLICT (stripe_subscription_id) DO UPDATE SET status = EXCLUDED.status, plan = EXCLUDED.plan,
  provider = EXCLUDED.provider, state = EXCLUDED.state, current_period_end = EXCLUDED.current_period_end,
  cancel_at_period_end = EXCLUDED.cancel_at_period_end, updated_at = EXCLUDED.updated_at`,
		true,
	},
	{"api_usage_daily", `INSERT INTO api_usage_daily (user_id, day, count)
SELECT n.id, u.day, u.count FROM legacy_import.api_usage_daily u JOIN users n ON n.legacy_id = u.user_id
WHERE u.user_id > $1 AND u.user_id <= $2
ON CONFLICT (user_id, day) DO UPDATE SET count = EXCLUDED.count`, false},
	// D-23: the remaining balance is the ledger sum, as legacy creditBalance computes it; an overdrawn
	// ledger (debits flushed after the cache allowed them) starts at zero.
	{"api_credit_balances", `INSERT INTO api_credit_balances (user_id, remaining, migrated_at)
SELECT n.id, greatest(sum(l.delta), 0), $3 FROM legacy_import.credit_ledger l JOIN users n ON n.legacy_id = l.user_id
WHERE l.user_id > $1 AND l.user_id <= $2 GROUP BY n.id
ON CONFLICT (user_id) DO UPDATE SET remaining = EXCLUDED.remaining, migrated_at = EXCLUDED.migrated_at`, true},
	{"credit_ledger", `INSERT INTO credit_ledger (id, user_id, delta, reason, ref, created_at)
SELECT l.id, n.id, l.delta, l.reason, l.ref, l.created_at
FROM legacy_import.credit_ledger l JOIN users n ON n.legacy_id = l.user_id
WHERE l.user_id > $1 AND l.user_id <= $2
ON CONFLICT (id) DO NOTHING`, false},
	{"pix_charges", `INSERT INTO pix_charges (id, user_id, provider_ref, purpose, amount_cents, status, br_code,
  br_code_base64, expires_at, paid_at, created_at)
SELECT p.id, n.id, p.provider_ref, p.purpose, p.amount_cents, p.status::text, p.br_code, p.br_code_base64,
  p.expires_at, p.paid_at, p.created_at
FROM legacy_import.pix_charges p JOIN users n ON n.legacy_id = p.user_id
WHERE p.user_id > $1 AND p.user_id <= $2
ON CONFLICT (id) DO NOTHING`, false},
}

// billingBatch imports the billing rows of the next n legacy users.
func billingBatch(now func() time.Time) BatchFunc {
	return userBatches("billing", billingSQL, nil, nil, now)
}

// userSQL is one statement of a stage keyed by legacy user id: it takes $1 < legacy user id <= $2, and
// the import time as $3 when stamped. Statements run once (first batch or end) take only that time, as $1.
type userSQL struct {
	name, query string
	stamped     bool
}

// userBatches imports the rows of the next n legacy users with stmts; the rows counted are users, so
// the checkpoint key is the last legacy user id. first runs with the first batch, last after the final
// one (in the empty batch that ends the stage); both must be idempotent like stmts.
func userBatches(stage string, stmts, first, last []userSQL, now func() time.Time) BatchFunc {
	once := func(ctx context.Context, tx *database.Tx, list []userSQL, at time.Time) error {
		for _, s := range list {
			var args []any
			if s.stamped {
				args = append(args, at)
			}

			if _, err := tx.ExecContext(ctx, s.query, args...); err != nil {
				return fmt.Errorf("%s %s: %w", stage, s.name, err)
			}
		}

		return nil
	}

	return func(ctx context.Context, tx *database.Tx, after string, n int) (string, int, error) {
		var (
			lastKey string
			users   int
		)

		err := tx.QueryRowContext(ctx, `SELECT coalesce(max(id), ''), count(*) FROM (SELECT id FROM legacy_import."user"
WHERE id > $1 ORDER BY id LIMIT $2) b`, after, n).Scan(&lastKey, &users)
		if err != nil {
			return "", 0, err
		}

		at := now().UTC()

		if after == "" {
			if err := once(ctx, tx, first, at); err != nil {
				return "", 0, err
			}
		}

		if users == 0 {
			return "", 0, once(ctx, tx, last, at)
		}

		for _, s := range stmts {
			args := []any{after, lastKey}
			if s.stamped {
				args = append(args, at)
			}

			if _, err := tx.ExecContext(ctx, s.query, args...); err != nil {
				return "", 0, fmt.Errorf("%s %s: %w", stage, s.name, err)
			}
		}

		return lastKey, users, nil
	}
}
