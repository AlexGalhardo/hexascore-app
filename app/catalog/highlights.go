package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// PoolSize is the number of released trending items shown in each landing block.
const PoolSize = 12

// Highlight is one ranked landing row. Score is nil until the item has 5
// counted ratings (SDD 02 read gate).
type Highlight struct {
	ID               string   `json:"id"`
	Type             string   `json:"type"`
	Slug             string   `json:"slug"`
	Title            string   `json:"title"`
	Year             *int     `json:"year"`
	ReleaseDate      *string  `json:"releaseDate,omitempty"`
	ReleasePrecision string   `json:"releasePrecision"`
	Score            *float64 `json:"score"`
	RatingCount      int      `json:"ratingCount"`
	Cover            *Cover   `json:"cover"`
}

// SectionHighlights is one item type's landing section: the public item count and the recommendation pool.
type SectionHighlights struct {
	Type  string      `json:"type"`
	Count int         `json:"count"`
	Pool  []Highlight `json:"pool"`
}

// Highlights loads the public item count and the most active released items over 30 days.
// popularity_prior carries external source popularity when on-site activity is sparse.
func Highlights(ctx context.Context, db Querier, typ, locale string) (SectionHighlights, error) {
	out := SectionHighlights{Type: typ, Pool: []Highlight{}}

	err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM items WHERE type = $1 AND visibility = 'public'`, typ).Scan(&out.Count)
	if err != nil {
		return out, fmt.Errorf("count highlights: %w", err)
	}

	q := `WITH activity AS (
  SELECT item_id, sum(views + 3*search_clicks + 5*want_w) AS hype
  FROM item_activity_daily WHERE day > CURRENT_DATE - 30 GROUP BY item_id
)
SELECT i.id, i.type, i.slug,
  coalesce(t.title, te.title, i.original_title, i.slug),
  extract(year FROM i.release_date)::int, to_char(i.release_date, 'YYYY-MM-DD'), i.release_precision,
  ` + starsScore + `,
  i.rating_count, i.cover_key, i.cover_thumb_key
FROM items i
LEFT JOIN activity a ON a.item_id = i.id
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $2
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
WHERE i.type = $1 AND i.visibility = 'public' AND i.release_status = 'released'
  AND (i.release_date IS NULL OR i.release_date <= CURRENT_DATE)
ORDER BY coalesce(a.hype, 0) DESC,
  i.popularity_prior DESC, i.rating_count DESC, i.id
LIMIT ` + strconv.Itoa(PoolSize)
	rows, err := db.QueryContext(ctx, q, typ, locale)
	if err != nil {
		return out, fmt.Errorf("highlights trending: %w", err)
	}
	out.Pool, err = scanHighlights(rows)
	if err != nil {
		return out, err
	}

	return out, nil
}

// Upcoming ranks public announcements in the next three calendar months by external popularity and on-site intent.
func Upcoming(ctx context.Context, db Querier, typ, locale string) ([]Highlight, error) {
	rows, err := db.QueryContext(ctx, `SELECT i.id, i.type, i.slug,
  coalesce(t.title, te.title, i.original_title, i.slug),
  extract(year FROM i.release_date)::int, to_char(i.release_date, 'YYYY-MM-DD'), i.release_precision, `+starsScore+`,
  i.rating_count, i.cover_key, i.cover_thumb_key
FROM items i
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $2
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
WHERE i.type = $1 AND i.visibility = 'public' AND i.release_status = 'announced'
  AND i.release_date > CURRENT_DATE AND i.release_date <= (CURRENT_DATE + INTERVAL '3 months')::date
ORDER BY (0.7 * i.popularity_prior + 10 * ln(1 +
  (SELECT count(*) FROM item_status s WHERE s.item_id = i.id AND s.status = 'want'))) DESC,
  i.release_date, i.id
LIMIT 12`, typ, locale)
	if err != nil {
		return nil, fmt.Errorf("upcoming highlights: %w", err)
	}
	return scanHighlights(rows)
}

func scanHighlights(rows *sql.Rows) ([]Highlight, error) {
	defer rows.Close()

	got := []Highlight{}
	for rows.Next() {
		var h Highlight
		var full, thumb, releaseDate sql.NullString
		err := rows.Scan(&h.ID, &h.Type, &h.Slug, &h.Title, &h.Year, &releaseDate, &h.ReleasePrecision, &h.Score, &h.RatingCount, &full, &thumb)
		if err != nil {
			return nil, fmt.Errorf("scan highlight: %w", err)
		}
		h.Cover = NewCover(full, thumb)
		if releaseDate.Valid {
			h.ReleaseDate = &releaseDate.String
		}
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
  extract(year FROM i.release_date)::int, to_char(i.release_date, 'YYYY-MM-DD'), i.release_precision,
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
