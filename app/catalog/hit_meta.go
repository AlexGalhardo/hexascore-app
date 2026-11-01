package catalog

import (
	"context"
	"fmt"
	"strings"
)

// StarsScore is the gated 0-10 score expression over items aliased i, for other packages' queries.
const StarsScore = starsScore

// HitMeta is the release year and the community score (0-10, nil under 5 counted ratings) shown on a result row.
type HitMeta struct {
	Year  *int
	Score *float64
}

// HitMetaFor loads HitMeta for the given item ids; unknown ids are absent from the map.
func HitMetaFor(ctx context.Context, db Querier, ids []string) (map[string]HitMeta, error) {
	out := map[string]HitMeta{}
	if len(ids) == 0 {
		return out, nil
	}

	rows, err := db.QueryContext(
		ctx,
		`SELECT i.id::text, extract(year FROM i.release_date)::int, `+StarsScore+` FROM items i WHERE i.id = ANY($1::uuid[])`,
		"{"+strings.Join(ids, ",")+"}",
	)
	if err != nil {
		return nil, fmt.Errorf("load hit meta: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var m HitMeta
		if err := rows.Scan(&id, &m.Year, &m.Score); err != nil {
			return nil, fmt.Errorf("scan hit meta: %w", err)
		}

		out[id] = m
	}

	return out, rows.Err()
}
