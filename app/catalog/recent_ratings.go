package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
)

// RecentRatingRows is the size of the home "recently rated" strip.
const RecentRatingRows = 12

// RecentRating is one public rating shown on the home page. Score is the
// member's own overall (0-10), not the item's aggregate.
type RecentRating struct {
	ID    string  `json:"id"`
	Type  string  `json:"type"`
	Slug  string  `json:"slug"`
	Title string  `json:"title"`
	Score float64 `json:"score"`
	By    string  `json:"by"`
	Cover *Cover  `json:"cover"`
}

// RecentRatings returns the latest ratings of public items by members whose
// profile is public, not banned and not anonymized (SDD 03 §7.5).
func RecentRatings(ctx context.Context, db Querier, locale string) ([]RecentRating, error) {
	rows, err := db.QueryContext(ctx, `SELECT i.id, i.type, i.slug,
  coalesce(t.title, te.title, i.original_title, i.slug),
  round(r.overall) / 10, coalesce(u.display_username, u.username), i.cover_key, i.cover_thumb_key
FROM ratings r
JOIN users u ON u.id = r.user_id AND u.username IS NOT NULL AND NOT u.profile_private
  AND u.anonymized_at IS NULL AND u.deletion_scheduled_at IS NULL AND NOT (u.banned_at IS NOT NULL AND (u.ban_expires_at IS NULL OR u.ban_expires_at > now()))
JOIN items i ON i.id = r.item_id AND i.visibility = 'public'
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $1
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
ORDER BY r.updated_at DESC, i.id DESC LIMIT `+strconv.Itoa(RecentRatingRows), locale)
	if err != nil {
		return nil, fmt.Errorf("recent ratings: %w", err)
	}
	defer rows.Close()

	out := []RecentRating{}
	for rows.Next() {
		var r RecentRating
		var full, thumb sql.NullString
		if err := rows.Scan(&r.ID, &r.Type, &r.Slug, &r.Title, &r.Score, &r.By, &full, &thumb); err != nil {
			return nil, fmt.Errorf("scan recent rating: %w", err)
		}
		r.Cover = NewCover(full, thumb)
		out = append(out, r)
	}

	return out, rows.Err()
}
