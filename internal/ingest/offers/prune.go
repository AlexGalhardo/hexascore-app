package offers

import (
	"context"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// Retention of the offers tables (SDD 11 section 5): TMDB data may not be kept past 180 days (ADR 0006), so a
// provider offer is deactivated when it expires and deleted ExpiredKeepDays later; click counters keep 400 days.
// Offers from other sources are never deleted: a deactivated one is the tombstone that stops offers.derive from
// creating it again.
const (
	ExpiredKeepDays = 30
	ClickKeepDays   = 400
)

// Pruned counts what one Prune did.
type Pruned struct {
	Deactivated, Deleted, Clicks int
}

// Prune applies the retention above as of now. It is idempotent.
func Prune(ctx context.Context, db *database.DB, now time.Time) (Pruned, error) {
	var p Pruned

	for _, step := range []struct {
		n    *int
		sql  string
		args []any
	}{
		{
			&p.Deactivated, `UPDATE offers SET active = false WHERE source = 'tmdb_providers' AND active AND expires_at < $1`,
			[]any{now},
		},
		{
			&p.Deleted, `DELETE FROM offers WHERE source = 'tmdb_providers' AND NOT active AND expires_at < $1`,
			[]any{now.AddDate(0, 0, -ExpiredKeepDays)},
		},
		{&p.Clicks, `DELETE FROM offer_clicks WHERE day < $1::date`, []any{now.AddDate(0, 0, -ClickKeepDays)}},
	} {
		res, err := db.ExecContext(ctx, step.sql, step.args...)
		if err != nil {
			return p, fmt.Errorf("offers prune: %w", err)
		}

		n, _ := res.RowsAffected()
		*step.n = int(n)
	}

	return p, nil
}
