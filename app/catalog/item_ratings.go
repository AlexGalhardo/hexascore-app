package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

const (
	// RatingsPageSize is the legacy public page (10 rows).
	RatingsPageSize = 10
	// RatingsMaxPages caps the numbered window at the newest 1000 ratings.
	RatingsMaxPages = 100
)

// PublicRating is one row of GET /api/items/{id}/ratings.
// Overall and categories are 0–10 stars with one decimal, matching the item page.
type PublicRating struct {
	Nickname   string     `json:"nickname"`
	Overall    float64    `json:"overall"`
	Categories [6]float64 `json:"categories"`
	RatedAt    string     `json:"ratedAt"`
}

// PublicRatingsPage is the legacy payload: data, page, page_count.
type PublicRatingsPage struct {
	Data      []PublicRating `json:"data"`
	Page      int            `json:"page"`
	PageCount int            `json:"page_count"`
}

// PublicItemRatings lists public profiles for a public item.
// page is 1-based and already validated. A page past the end returns an empty data array.
func PublicItemRatings(ctx context.Context, db Querier, itemID string, page int) (PublicRatingsPage, error) {
	var one int
	err := db.QueryRowContext(ctx,
		`SELECT 1 FROM items WHERE id = $1 AND visibility = 'public'`, itemID).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return PublicRatingsPage{}, ErrItemNotFound
	}
	if err != nil {
		return PublicRatingsPage{}, fmt.Errorf("ratings item: %w", err)
	}

	const visible = `
FROM ratings r
JOIN users u ON u.id = r.user_id
WHERE r.item_id = $1
  AND u.profile_private = false
  AND u.username IS NOT NULL
  AND (
    u.banned_at IS NULL
    OR (u.ban_expires_at IS NOT NULL AND u.ban_expires_at <= now())
  )
  AND u.deletion_scheduled_at IS NULL`

	var total int
	err = db.QueryRowContext(ctx,
		`SELECT count(*) FROM (SELECT 1 `+visible+` LIMIT $2) capped`,
		itemID, RatingsPageSize*RatingsMaxPages).Scan(&total)
	if err != nil {
		return PublicRatingsPage{}, fmt.Errorf("ratings count: %w", err)
	}

	out := PublicRatingsPage{
		Data:      []PublicRating{},
		Page:      page,
		PageCount: (total + RatingsPageSize - 1) / RatingsPageSize,
	}
	rows, err := db.QueryContext(ctx,
		`SELECT u.username, r.overall, r.c1, r.c2, r.c3, r.c4, r.c5, r.c6, r.created_at `+visible+`
ORDER BY r.created_at DESC, r.user_id DESC
LIMIT $2 OFFSET $3`,
		itemID, RatingsPageSize, (page-1)*RatingsPageSize)
	if err != nil {
		return PublicRatingsPage{}, fmt.Errorf("ratings page: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			row PublicRating
			c   [6]float64
			at  time.Time
		)
		if err := rows.Scan(&row.Nickname, &row.Overall, &c[0], &c[1], &c[2], &c[3], &c[4], &c[5], &at); err != nil {
			return PublicRatingsPage{}, fmt.Errorf("scan rating: %w", err)
		}
		row.Overall = starsOut(row.Overall)
		for i := range c {
			row.Categories[i] = starsOut(c[i])
		}
		row.RatedAt = at.UTC().Format(time.RFC3339)
		out.Data = append(out.Data, row)
	}
	if err := rows.Err(); err != nil {
		return PublicRatingsPage{}, fmt.Errorf("ratings rows: %w", err)
	}
	return out, nil
}

func starsOut(stored float64) float64 {
	return math.Round(stored) / 10
}

func validYoutubeID(id string) bool {
	if len(id) != 11 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
		default:
			return false
		}
	}
	return true
}

func assignTrailers(raw []byte) ([]models.Trailer, error) {
	out := []models.Trailer{}
	if len(raw) == 0 || string(raw) == "null" {
		return out, nil
	}
	var all []models.Trailer
	if err := json.Unmarshal(raw, &all); err != nil {
		return nil, fmt.Errorf("trailers: %w", err)
	}
	for _, tr := range all {
		if validYoutubeID(tr.YoutubeID) {
			out = append(out, tr)
		}
	}
	return out, nil
}
