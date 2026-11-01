package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// ProfilePageSize is the recent ratings per profile page (SDD 14 §7.2).
const ProfilePageSize = 20

// ProfileHead is the part of a profile every visitor sees. A private profile (and nothing else) shows only
// this: nickname and initial, 200 + noindex, no counts (SDD 14 FR-14-8, open issue 14-5).
type ProfileHead struct {
	UserID    int64  `json:"-"`
	Nickname  string `json:"nickname"`
	Initial   string `json:"initial"`
	IsPrivate bool   `json:"isPrivate"`
	IsOwner   bool   `json:"isOwner"`
	// Banned is a currently banned owner: the profile still shows but is noindex (SDD 14 §8).
	Banned bool `json:"-"`
}

// ProfileRating is one recent rating on a profile; Stars are 0.5–10.
type ProfileRating struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Slug          string         `json:"slug"`
	Title         string         `json:"title"`
	Year          *int           `json:"year"`
	CoverThumbKey *string        `json:"coverThumbKey"`
	Cover         *catalog.Cover `json:"cover"`
	Stars         float64        `json:"stars"`
	Score         *float64       `json:"score"`
	RatedAt       time.Time      `json:"ratedAt"`
}

// ProfileList is a public list shown on a profile.
type ProfileList struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	ItemCount int    `json:"itemCount"`
}

// ProfileStats are the rating and status counts per type.
type ProfileStats struct {
	RatingsByType map[string]int            `json:"ratingsByType"`
	StatusByType  map[string]map[string]int `json:"statusByType"`
}

// Profile is a visitor's view of a profile. Stats, Ratings and Lists are nil for a private profile the
// viewer does not own. NextRatings is the offset of the next ratings page (-1 on the last).
type Profile struct {
	ProfileHead
	Stats       *ProfileStats   `json:"stats,omitempty"`
	Ratings     []ProfileRating `json:"ratings,omitempty"`
	NextRatings int             `json:"-"`
	Lists       []ProfileList   `json:"lists,omitempty"`
}

// LoadProfile finds the profile of nickname (case-insensitive) as seen by viewerID (0 when anonymous). Unknown,
// and anonymised accounts are ErrNotFound. ratingsOffset pages the recent ratings.
func LoadProfile(
	ctx context.Context, db *database.DB, nickname string, viewerID int64, locale string, ratingsOffset int,
) (Profile, error) {
	p := Profile{NextRatings: -1}
	var display sql.NullString

	err := db.SQL().QueryRowContext(ctx, `SELECT id, coalesce(display_username, username), profile_private,
  banned_at IS NOT NULL AND (ban_expires_at IS NULL OR ban_expires_at > now()) FROM users
WHERE username = $1 AND anonymized_at IS NULL`, strings.ToLower(nickname)).
		Scan(&p.UserID, &display, &p.IsPrivate, &p.Banned)
	if errors.Is(err, sql.ErrNoRows) {
		return p, ErrNotFound()
	}

	if err != nil {
		return p, fmt.Errorf("load profile: %w", err)
	}

	p.Nickname = display.String
	p.Initial = strings.ToUpper(string([]rune(p.Nickname)[:1]))
	p.IsOwner = viewerID != 0 && viewerID == p.UserID

	if p.IsPrivate && !p.IsOwner {
		return p, nil
	}

	if p.Stats, err = profileStats(ctx, db, p.UserID); err != nil {
		return p, err
	}

	if p.Ratings, p.NextRatings, err = profileRatings(ctx, db, p.UserID, locale, max(ratingsOffset, 0)); err != nil {
		return p, err
	}

	p.Lists, err = profileLists(ctx, db, p.UserID)

	return p, err
}

func profileStats(ctx context.Context, db *database.DB, userID int64) (*ProfileStats, error) {
	st := &ProfileStats{RatingsByType: map[string]int{}, StatusByType: map[string]map[string]int{}}

	rows, err := db.SQL().QueryContext(ctx, `SELECT i.type, count(*) FROM ratings r
JOIN items i ON i.id = r.item_id AND i.visibility = 'public' WHERE r.user_id = $1 GROUP BY 1`, userID)
	if err != nil {
		return nil, fmt.Errorf("profile rating counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var typ string
		var n int
		if err := rows.Scan(&typ, &n); err != nil {
			return nil, err
		}

		st.RatingsByType[typ] = n
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	srows, err := db.SQL().QueryContext(ctx, `SELECT i.type, s.status::text, count(*) FROM item_status s
JOIN items i ON i.id = s.item_id AND i.visibility = 'public' WHERE s.user_id = $1 GROUP BY 1, 2`, userID)
	if err != nil {
		return nil, fmt.Errorf("profile status counts: %w", err)
	}
	defer srows.Close()

	for srows.Next() {
		var typ, status string
		var n int
		if err := srows.Scan(&typ, &status, &n); err != nil {
			return nil, err
		}

		if st.StatusByType[typ] == nil {
			st.StatusByType[typ] = map[string]int{}
		}

		st.StatusByType[typ][status] = n
	}

	return st, srows.Err()
}

func profileRatings(
	ctx context.Context, db *database.DB, userID int64, locale string, offset int,
) ([]ProfileRating, int, error) {
	rows, err := db.SQL().QueryContext(ctx, `SELECT i.id, i.type, i.slug, `+libraryTitle+`,
  extract(year FROM i.release_date)::int, i.cover_thumb_key, round(r.overall / 10, 1)::float8, `+catalog.StarsScore+`, r.updated_at
FROM ratings r
JOIN items i ON i.id = r.item_id AND i.visibility = 'public'
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $2
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
WHERE r.user_id = $1 ORDER BY r.updated_at DESC, i.id DESC LIMIT $3 OFFSET $4`,
		userID, locale, ProfilePageSize+1, offset)
	if err != nil {
		return nil, -1, fmt.Errorf("profile ratings: %w", err)
	}
	defer rows.Close()

	out := []ProfileRating{}
	for rows.Next() {
		var r ProfileRating
		if err := rows.Scan(
			&r.ID,
			&r.Type,
			&r.Slug,
			&r.Title,
			&r.Year,
			&r.CoverThumbKey,
			&r.Stars,
			&r.Score,
			&r.RatedAt,
		); err != nil {
			return nil, -1, err
		}

		r.Cover = catalog.CoverFromThumbKey(r.CoverThumbKey)
		out = append(out, r)
	}

	if err := rows.Err(); err != nil {
		return nil, -1, err
	}

	next := -1
	if len(out) > ProfilePageSize {
		out, next = out[:ProfilePageSize], offset+ProfilePageSize
	}

	return out, next, nil
}

func profileLists(ctx context.Context, db *database.DB, userID int64) ([]ProfileList, error) {
	rows, err := db.SQL().QueryContext(ctx, `SELECT id, type, title, item_count FROM lists
WHERE user_id = $1 AND is_public AND NOT hidden ORDER BY updated_at DESC, id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("profile lists: %w", err)
	}
	defer rows.Close()

	out := []ProfileList{}
	for rows.Next() {
		var l ProfileList
		if err := rows.Scan(&l.ID, &l.Type, &l.Title, &l.ItemCount); err != nil {
			return nil, err
		}

		out = append(out, l)
	}

	return out, rows.Err()
}
