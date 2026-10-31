package trending

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Querier is the read side of the database the readers need.
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Row is one rank of a stored snapshot.
type Row struct {
	Rank   int
	Basis  string
	Score  float64
	ItemID string
}

// Latest returns the newest snapshot of the type, and the zero time when there is none. Only public items are
// returned: a snapshot row of an item hidden since then is dropped.
func Latest(ctx context.Context, q Querier, typ string) (time.Time, []Row, error) {
	rows, err := q.QueryContext(ctx, `SELECT s.computed_at, s.rank, s.basis, s.trend_score, s.item_id::text
FROM trending_snapshots s JOIN items i ON i.id = s.item_id AND i.visibility = 'public'
WHERE s.item_type = $1 AND s.computed_at = (SELECT max(computed_at) FROM trending_snapshots WHERE item_type = $1)
ORDER BY s.rank`, typ)
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("trending latest: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var (
		at  time.Time
		out []Row
	)

	for rows.Next() {
		var r Row
		if err := rows.Scan(&at, &r.Rank, &r.Basis, &r.Score, &r.ItemID); err != nil {
			return time.Time{}, nil, err
		}

		out = append(out, r)
	}

	return at.UTC(), out, rows.Err()
}

// DayCounts is the raw, unweighted activity of one UTC day. The weighted columns are never exposed.
type DayCounts struct {
	Day              string
	Views            int
	SearchClicks     int
	RatingsNew       int
	RatingsUpdated   int
	StatusWant       int
	StatusInProgress int
	StatusCompleted  int
	StatusDropped    int
	ListAdds         int
	Suggestions      int
}

// Activity returns the item's daily counts for from..to inclusive, oldest first. Days without a row are absent.
func Activity(ctx context.Context, q Querier, itemID string, from, to time.Time) ([]DayCounts, error) {
	rows, err := q.QueryContext(ctx, `SELECT day::text, views, search_clicks, ratings_new, ratings_updated,
status_want, status_in_progress, status_completed, status_dropped, list_adds, suggestions
FROM item_activity_daily WHERE item_id = $1 AND day BETWEEN $2::date AND $3::date ORDER BY day`, itemID, from, to)
	if err != nil {
		return nil, fmt.Errorf("trending activity: %w", err)
	}

	defer func() { _ = rows.Close() }()

	out := []DayCounts{}

	for rows.Next() {
		var d DayCounts
		if err := rows.Scan(&d.Day, &d.Views, &d.SearchClicks, &d.RatingsNew, &d.RatingsUpdated, &d.StatusWant,
			&d.StatusInProgress, &d.StatusCompleted, &d.StatusDropped, &d.ListAdds, &d.Suggestions); err != nil {
			return nil, err
		}

		out = append(out, d)
	}

	return out, rows.Err()
}
