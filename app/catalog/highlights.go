package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// PoolSize is how many items each landing block sends for the "More …" grid; the client shows a random 12 of them
// after hydration, so the cached HTML stays identical for every visitor (owner round 2026-11-01 item 26).
const PoolSize = 36

// Highlight is one ranked landing row. Score is nil until the item has 5
// counted ratings (SDD 02 read gate).
type Highlight struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Year        *int     `json:"year"`
	Score       *float64 `json:"score"`
	RatingCount int      `json:"ratingCount"`
	Cover       *Cover   `json:"cover"`
}

// SectionHighlights is one item type's landing section: the public item count and the recommendation pool.
type SectionHighlights struct {
	Type  string      `json:"type"`
	Count int         `json:"count"`
	Pool  []Highlight `json:"pool"`
}

// Highlights loads the public item count and the pool of typ: the best scored items, then the most rated ones
// until PoolSize. Titles fall back from locale to en, then to the original title.
func Highlights(ctx context.Context, db Querier, typ, locale string) (SectionHighlights, error) {
	out := SectionHighlights{Type: typ, Pool: []Highlight{}}

	err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM items WHERE type = $1 AND visibility = 'public'`, typ).Scan(&out.Count)
	if err != nil {
		return out, fmt.Errorf("count highlights: %w", err)
	}

	seen := map[string]bool{}
	for _, b := range []struct{ sort, gate string }{
		{"-score", " AND " + visibleScore},
		{"-rating_count", ""},
	} {
		field, _, _ := ParseSort(b.sort)
		q := `SELECT i.id, i.type, i.slug,
  coalesce(t.title, te.title, i.original_title, i.slug),
  extract(year FROM i.release_date)::int,
  ` + starsScore + `,
  i.rating_count, i.cover_key, i.cover_thumb_key
FROM items i
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $2
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
WHERE i.type = $1 AND i.visibility = 'public'` + b.gate + `
ORDER BY ` + sortFields[field].desc + ` DESC, i.id DESC LIMIT ` + strconv.Itoa(PoolSize)

		rows, err := db.QueryContext(ctx, q, typ, locale)
		if err != nil {
			return out, fmt.Errorf("highlights %s: %w", b.sort, err)
		}
		got, err := scanHighlights(rows)
		if err != nil {
			return out, err
		}
		for _, h := range got {
			if len(out.Pool) < PoolSize && !seen[h.ID] {
				seen[h.ID] = true
				out.Pool = append(out.Pool, h)
			}
		}
	}

	return out, nil
}

func scanHighlights(rows *sql.Rows) ([]Highlight, error) {
	defer rows.Close()

	got := []Highlight{}
	for rows.Next() {
		var h Highlight
		var full, thumb sql.NullString
		err := rows.Scan(&h.ID, &h.Type, &h.Slug, &h.Title, &h.Year, &h.Score, &h.RatingCount, &full, &thumb)
		if err != nil {
			return nil, fmt.Errorf("scan highlight: %w", err)
		}
		h.Cover = NewCover(full, thumb)
		got = append(got, h)
	}

	return got, rows.Err()
}

// HighlightsByID loads the public items among ids, in the order of ids. Titles fall back like Highlights.
func HighlightsByID(ctx context.Context, db Querier, ids []string, locale string) ([]Highlight, error) {
	if len(ids) == 0 {
		return []Highlight{}, nil
	}

	rows, err := db.QueryContext(ctx, `SELECT i.id, i.type, i.slug,
  coalesce(t.title, te.title, i.original_title, i.slug),
  extract(year FROM i.release_date)::int,
  `+starsScore+`,
  i.rating_count, i.cover_key, i.cover_thumb_key
FROM unnest($1::uuid[]) WITH ORDINALITY AS w(id, n)
JOIN items i ON i.id = w.id AND i.visibility = 'public'
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $2
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
ORDER BY w.n`, "{"+strings.Join(ids, ",")+"}", locale)
	if err != nil {
		return nil, fmt.Errorf("highlights by id: %w", err)
	}

	return scanHighlights(rows)
}
