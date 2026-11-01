package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
)

// HighlightRows is the row count of each landing block (legacy SectionHighlights).
const HighlightRows = 6

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
}

// SectionHighlights is one item type's landing section.
type SectionHighlights struct {
	Type      string      `json:"type"`
	Count     int         `json:"count"`
	TopRated  []Highlight `json:"topRated"`
	Newest    []Highlight `json:"newest"`
	MostRated []Highlight `json:"mostRated"`
}

// Highlights loads the three landing lists and the public item count for typ.
// Titles fall back from locale to en, then to the original title.
func Highlights(ctx context.Context, db Querier, typ, locale string) (SectionHighlights, error) {
	out := SectionHighlights{Type: typ, TopRated: []Highlight{}, Newest: []Highlight{}, MostRated: []Highlight{}}

	err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM items WHERE type = $1 AND visibility = 'public'`, typ).Scan(&out.Count)
	if err != nil {
		return out, fmt.Errorf("count highlights: %w", err)
	}

	blocks := []struct {
		dst  *[]Highlight
		sort string
		gate string
	}{
		{&out.TopRated, "-score", " AND " + visibleScore},
		{&out.Newest, "-release_date", ""},
		{&out.MostRated, "-rating_count", ""},
	}
	for _, b := range blocks {
		field, _, _ := ParseSort(b.sort)
		q := `SELECT i.id, i.type, i.slug,
  coalesce(t.title, te.title, i.original_title, i.slug),
  extract(year FROM i.release_date)::int,
  ` + starsScore + `,
  i.rating_count
FROM items i
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $2
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
WHERE i.type = $1 AND i.visibility = 'public'` + b.gate + `
ORDER BY ` + sortFields[field].desc + ` DESC, i.id DESC LIMIT ` + strconv.Itoa(HighlightRows)

		rows, err := db.QueryContext(ctx, q, typ, locale)
		if err != nil {
			return out, fmt.Errorf("highlights %s: %w", b.sort, err)
		}
		got, err := scanHighlights(rows)
		if err != nil {
			return out, err
		}
		*b.dst = got
	}

	return out, nil
}

func scanHighlights(rows *sql.Rows) ([]Highlight, error) {
	defer rows.Close()

	got := []Highlight{}
	for rows.Next() {
		var h Highlight
		if err := rows.Scan(&h.ID, &h.Type, &h.Slug, &h.Title, &h.Year, &h.Score, &h.RatingCount); err != nil {
			return nil, fmt.Errorf("scan highlight: %w", err)
		}
		got = append(got, h)
	}

	return got, rows.Err()
}
