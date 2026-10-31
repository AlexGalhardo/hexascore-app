package trending

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// primitive pairs a count column of item_activity_daily with its weight column. Views and search clicks are not
// here: the in-memory counter owns them and the rollup never touches them.
type primitive struct{ name, count, weight string }

var primitives = []primitive{
	{"ratings_new", "ratings_new", "ratings_new_w"},
	{"ratings_updated", "ratings_updated", "ratings_updated_w"},
	{"want", "status_want", "want_w"},
	{"in_progress", "status_in_progress", "in_progress_w"},
	{"completed", "status_completed", "completed_w"},
	{"dropped", "status_dropped", "dropped_w"},
	{"list_adds", "list_adds", "list_adds_w"},
	{"suggestions", "suggestions", "suggestions_w"},
}

// tier is the SDD 02 §7.1 age weight of an account at the moment of the event.
func tier(created, at string) string {
	age := at + " - " + created

	return "(CASE WHEN " + age + " < interval '7 days' THEN 0.25 WHEN " + age + " < interval '30 days' THEN 0.5 " +
		"WHEN " + age + " < interval '90 days' THEN 0.75 ELSE 1.0 END)"
}

// weigh scales a base weight by the shadow weight of (user, item) and zeroes banned accounts. $1 and $2 bound the day.
func weigh(base string) string {
	return base + " * coalesce(f.shadow_weight, 1) * (CASE WHEN u.banned_at IS NULL THEN 1 ELSE 0 END)"
}

// contributions yields (item_id, user_id, prim, w) for the day [$1, $2): one row per qualifying event.
var contributions = `
SELECT e.item_id, e.user_id,
  CASE e.kind WHEN 'rating_created' THEN 'ratings_new' ELSE 'ratings_updated' END AS prim,
  ` + weigh(`coalesce(e.weight, `+tier("u.created_at", "e.at")+`)`) + `::numeric AS w
FROM rating_events e
JOIN users u ON u.id = e.user_id
LEFT JOIN rating_flags f ON f.user_id = e.user_id AND f.item_id = e.item_id
WHERE e.kind IN ('rating_created', 'rating_updated') AND e.at >= $1 AND e.at < $2
  AND NOT EXISTS (SELECT 1 FROM rating_events d WHERE d.kind = 'rating_deleted' AND d.user_id = e.user_id
    AND d.item_id = e.item_id AND d.at > e.at AND d.at < $2)
UNION ALL
-- status_changed events carry no status_after, so the current status row stands for the day it changed.
SELECT s.item_id, s.user_id, s.status::text, ` + weigh(tier("u.created_at", "s.updated_at")) + `::numeric
FROM item_status s
JOIN users u ON u.id = s.user_id
LEFT JOIN rating_flags f ON f.user_id = s.user_id AND f.item_id = s.item_id
WHERE s.updated_at >= $1 AND s.updated_at < $2
UNION ALL
SELECT li.item_id, l.user_id, 'list_adds', ` + weigh(tier("u.created_at", "li.added_at")) + `::numeric
FROM list_items li
JOIN lists l ON l.id = li.list_id AND l.is_public AND NOT l.hidden
JOIN users u ON u.id = l.user_id
LEFT JOIN rating_flags f ON f.user_id = l.user_id AND f.item_id = li.item_id
WHERE li.added_at >= $1 AND li.added_at < $2 AND li.added_at - u.created_at >= interval '7 days'
UNION ALL
SELECT s.target_id, s.author_id, 'suggestions', ` + weigh(tier("u.created_at", "ev.created_at")) + `::numeric
FROM suggestion_events ev
JOIN suggestions s ON s.id = ev.suggestion_id AND s.target_type = 'item' AND s.target_id IS NOT NULL
JOIN users u ON u.id = s.author_id
LEFT JOIN rating_flags f ON f.user_id = s.author_id AND f.item_id = s.target_id
WHERE ev.to_status = 'submitted' AND ev.created_at >= $1 AND ev.created_at < $2`

// rollupSQL builds the replace statements. A user counts once per (item, day, primitive) at their maximum weight.
func rollupSQL() (reset, replace string) {
	var zero, cols, vals, sets []string

	for _, p := range primitives {
		zero = append(zero, p.count+" = 0", p.weight+" = 0")
		cols = append(cols, p.count, p.weight)
		vals = append(vals,
			fmt.Sprintf("count(*) FILTER (WHERE prim = '%s')::int", p.name),
			fmt.Sprintf("coalesce(sum(w) FILTER (WHERE prim = '%s'), 0)::real", p.name))
		sets = append(sets, p.count+" = EXCLUDED."+p.count, p.weight+" = EXCLUDED."+p.weight)
	}

	reset = `UPDATE item_activity_daily SET ` + strings.Join(zero, ", ") + ` WHERE day = ($1 AT TIME ZONE 'UTC')::date`
	replace = `WITH c AS (` + contributions + `),
best AS (SELECT item_id, user_id, prim, max(w) AS w FROM c GROUP BY 1, 2, 3 HAVING max(w) > 0)
INSERT INTO item_activity_daily (item_id, day, item_type, ` + strings.Join(cols, ", ") + `)
SELECT b.item_id, ($1 AT TIME ZONE 'UTC')::date, i.type, ` + strings.Join(vals, ", ") + `
FROM best b JOIN items i ON i.id = b.item_id GROUP BY b.item_id, i.type
ON CONFLICT (item_id, day) DO UPDATE SET ` + strings.Join(sets, ", ")

	return reset, replace
}

// Rollup replaces the durable columns of item_activity_daily for the UTC day of now and the day before, in one
// transaction (SDD 08 §7.1). Views and search clicks are left alone. It is idempotent: running it twice yields
// identical rows, and a day that lost a contribution (rating deleted, user banned) is zeroed first.
func Rollup(ctx context.Context, db *database.DB, now time.Time) error {
	reset, replace := rollupSQL()
	today := now.UTC().Truncate(24 * time.Hour)

	return db.Tx(ctx, func(tx *database.Tx) error {
		for _, lo := range []time.Time{today.AddDate(0, 0, -1), today} {
			if _, err := tx.ExecContext(ctx, reset, lo); err != nil {
				return fmt.Errorf("trending rollup reset: %w", err)
			}

			if _, err := tx.ExecContext(ctx, replace, lo, lo.AddDate(0, 0, 1)); err != nil {
				return fmt.Errorf("trending rollup %s: %w", lo.Format(time.DateOnly), err)
			}
		}

		return nil
	})
}
