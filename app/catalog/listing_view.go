package catalog

import (
	"context"
	"fmt"
	"strings"
)

// Summaries returns the listing cards for ids, in the same order. Titles fall
// back from locale to en, then to the original title; the score follows the
// 5-rating read gate.
func Summaries(ctx context.Context, db Querier, ids []string, locale string) ([]Highlight, error) {
	if len(ids) == 0 {
		return []Highlight{}, nil
	}

	rows, err := db.QueryContext(ctx, `SELECT i.id, i.type, i.slug,
  coalesce(t.title, te.title, i.original_title, i.slug),
  extract(year FROM i.release_date)::int,
  CASE WHEN `+visibleScore+` THEN round(i.score, 1)::float8 END,
  i.rating_count
FROM unnest($1::uuid[]) WITH ORDINALITY AS o(id, n)
JOIN items i ON i.id = o.id
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $2
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
ORDER BY o.n`, "{"+strings.Join(ids, ",")+"}", locale)
	if err != nil {
		return nil, fmt.Errorf("listing summaries: %w", err)
	}

	return scanHighlights(rows)
}

// Facet is one selectable genre or platform with its public item count.
type Facet struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Facets lists the genres and platforms used by public items of typ, most used first; non-nil sources keeps
// only relations from those sources (API).
func Facets(ctx context.Context, db Querier, typ string, sources []string) (genres, platforms []Facet, err error) {
	var srcArray any // NULL = every source (site)
	if sources != nil {
		srcArray = "{" + strings.Join(sources, ",") + "}"
	}
	load := func(kind string) ([]Facet, error) {
		rows, err := db.QueryContext(ctx, `SELECT e.slug, e.default_name, count(*)
FROM entities e
JOIN item_entities ie ON ie.entity_id = e.id AND ie.role = e.kind AND ($3::text[] IS NULL OR ie.source = ANY($3))
JOIN items i ON i.id = ie.item_id AND i.type = $1 AND i.visibility = 'public'
WHERE e.kind = $2
GROUP BY e.slug, e.default_name
ORDER BY count(*) DESC, e.default_name
LIMIT 200`, typ, kind, srcArray)
		if err != nil {
			return nil, fmt.Errorf("facets %s: %w", kind, err)
		}
		defer rows.Close()

		out := []Facet{}
		for rows.Next() {
			var f Facet
			if err := rows.Scan(&f.Slug, &f.Name, &f.Count); err != nil {
				return nil, fmt.Errorf("scan facet: %w", err)
			}
			out = append(out, f)
		}

		return out, rows.Err()
	}

	if genres, err = load("genre"); err != nil {
		return nil, nil, err
	}
	platforms, err = load("platform")

	return genres, platforms, err
}
