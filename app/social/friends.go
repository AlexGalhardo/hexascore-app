package social

import (
	"context"
	"fmt"
	"strconv"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// FriendRating is one followed member's rating of an item (the 0 to 10 score, one decimal).
type FriendRating struct {
	Nickname string  `json:"nickname"`
	Score    float64 `json:"score"`
}

// FriendsScore is the average of the ratings of the people the viewer follows (SDD 08 §6, SDD 14 §12).
type FriendsScore struct {
	Count   int            `json:"count"`
	Average float64        `json:"average"`
	Friends []FriendRating `json:"friends"`
}

// friendRatings joins ratings to the viewer's accepted follows, keeping only members a count may include
// (§12.5: public, not banned, not anonymised).
const friendRatings = `FROM ratings r
JOIN follows f ON f.followee_id = r.user_id AND f.follower_id = $1 AND f.state = 'accepted'
JOIN users u ON u.id = r.user_id AND ` + visible

// FriendsScore averages the ratings of itemID by the people viewerID follows; nil when there are none.
func (s *Follows) FriendsScore(ctx context.Context, viewerID int64, itemID string) (*FriendsScore, error) {
	if !s.Enabled(ctx) || viewerID == 0 {
		return nil, nil
	}

	rows, err := s.DB.QueryContext(ctx,
		`SELECT coalesce(u.display_username, u.username), round(r.overall / 10, 1)::float8 `+friendRatings+
			` WHERE r.item_id = $2 ORDER BY r.updated_at DESC`, viewerID, itemID)
	if err != nil {
		return nil, fmt.Errorf("social: friends score: %w", err)
	}
	defer rows.Close()

	out := &FriendsScore{Friends: []FriendRating{}}

	var sum float64

	for rows.Next() {
		var f FriendRating
		if err := rows.Scan(&f.Nickname, &f.Score); err != nil {
			return nil, fmt.Errorf("social: friends score: %w", err)
		}

		out.Friends, sum = append(out.Friends, f), sum+f.Score
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("social: friends score: %w", err)
	}

	if out.Count = len(out.Friends); out.Count == 0 {
		return nil, nil
	}

	out.Average = float64(int(sum/float64(out.Count)*10+0.5)) / 10

	return out, nil
}

// FriendItemRating is a recent rating by a followed member, with the item to link to.
type FriendItemRating struct {
	FriendRating
	Item FeedItem `json:"item"`
}

// friendsStrip is how many ratings the home strip shows.
const friendsStrip = 12

// FriendsRated lists the latest ratings of public items by the people viewerID follows.
func (s *Follows) FriendsRated(ctx context.Context, viewerID int64, locale string) ([]FriendItemRating, error) {
	out := []FriendItemRating{}
	if !s.Enabled(ctx) || viewerID == 0 {
		return out, nil
	}

	rows, err := s.DB.QueryContext(ctx, `SELECT coalesce(u.display_username, u.username),
  round(r.overall / 10, 1)::float8, i.id, i.type, i.slug, coalesce(t.title, te.title, i.original_title, i.slug),
  i.cover_thumb_key `+friendRatings+`
JOIN items i ON i.id = r.item_id AND i.visibility = 'public'
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $2
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
ORDER BY r.updated_at DESC LIMIT `+strconv.Itoa(friendsStrip), viewerID, locale)
	if err != nil {
		return nil, fmt.Errorf("social: friends rated: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			f     FriendItemRating
			cover *string
		)

		if err := rows.Scan(&f.Nickname, &f.Score, &f.Item.ID, &f.Item.Type, &f.Item.Slug, &f.Item.Title,
			&cover); err != nil {
			return nil, fmt.Errorf("social: friends rated: %w", err)
		}

		f.Item.Cover = catalog.CoverFromThumbKey(cover)

		out = append(out, f)
	}

	return out, rows.Err()
}
