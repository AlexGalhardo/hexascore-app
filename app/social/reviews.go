package social

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// KindReview is the feed event of a first published review (SDD 14 FR-14-16).
const KindReview = "review_published"

// ReviewMaxLen is the review length limit in characters (SDD 14 §12.10).
const ReviewMaxLen = 5000

// reviewPage caps a review list; shortcut: no pagination yet, add a cursor when an item passes 50 reviews.
const reviewPage = 50

// ErrRatingRequired answers a review of an item the user has not rated: a review is attached to a rating.
func ErrRatingRequired() error {
	return errs.NewKey(http.StatusUnprocessableEntity, "social.review.rating_required", "Rate the item first.", nil).
		WithExtension("code", "rating-required")
}

// Review is one review card. Source, the raw Markdown, is set only on the viewer's own review (the edit form).
type Review struct {
	ID      string    `json:"id"`
	Author  Actor     `json:"author"`
	Item    *FeedItem `json:"item,omitempty"`
	Stars   *float64  `json:"stars"`
	Body    []Node    `json:"body"`
	Source  string    `json:"source,omitempty"`
	Spoiler bool      `json:"spoiler"`
	Helpful int       `json:"helpful"`
	Likes   int       `json:"likes"`
	// MyReaction and Comments are filled on the item page only (P7.B.1.2).
	MyReaction string    `json:"myReaction,omitempty"`
	Comments   []Comment `json:"comments,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// SaveReview creates or edits userID's review of itemID. The user must have rated the item; the first publication
// records a feed event after commit, edits do not.
func (s *Follows) SaveReview(ctx context.Context, userID int64, itemID, body string, spoiler bool) error {
	if _, err := uuid.Parse(itemID); err != nil {
		return ErrNotFound()
	}

	var created bool

	err := s.DB.Tx(ctx, func(tx *database.Tx) error {
		var rated bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM ratings WHERE user_id = $1 AND item_id = $2)`,
			userID, itemID).Scan(&rated); err != nil {
			return fmt.Errorf("social: review rating: %w", err)
		}

		if !rated {
			return ErrRatingRequired()
		}

		// xmax = 0 only on a freshly inserted row, so it tells a first publication from an edit.
		return tx.QueryRowContext(ctx, `INSERT INTO reviews (id, user_id, item_id, body, spoiler)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, item_id) DO UPDATE SET body = EXCLUDED.body, spoiler = EXCLUDED.spoiler, updated_at = now()
RETURNING xmax = 0`, uuid.NewV7().String(), userID, itemID, strings.TrimSpace(body), spoiler).Scan(&created)
	})
	if err != nil || !created {
		return err
	}

	return s.record(ctx, userID, KindReview, itemID, "", nil)
}

// DeleteReview removes userID's review of itemID; no review is not an error.
func (s *Follows) DeleteReview(ctx context.Context, userID int64, itemID string) error {
	if _, err := uuid.Parse(itemID); err != nil {
		return ErrNotFound()
	}

	if _, err := s.DB.ExecContext(ctx, `DELETE FROM reviews WHERE user_id = $1 AND item_id = $2`,
		userID, itemID); err != nil {
		return fmt.Errorf("social: delete review: %w", err)
	}

	return nil
}

// reviewQuery lists reviews whose rating still exists, by helpfulness then date; $1 is the locale, filter uses $2+.
func reviewQuery(filter string) string {
	return `SELECT v.id, v.user_id, coalesce(u.display_username, u.username), v.body, v.spoiler, v.helpful_count, v.like_count,
  v.created_at, v.updated_at, round(r.overall / 10, 1)::float8,
  i.id, i.type, i.slug, coalesce(t.title, te.title, i.original_title, i.slug), i.cover_thumb_key
FROM reviews v
JOIN ratings r ON r.user_id = v.user_id AND r.item_id = v.item_id
JOIN users u ON u.id = v.user_id
JOIN items i ON i.id = v.item_id
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $1
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
WHERE NOT v.hidden AND ` + filter + `
ORDER BY v.helpful_count DESC, v.created_at DESC LIMIT ` + strconv.Itoa(
		reviewPage,
	)
}

// ItemReviews lists the reviews of itemID by public, active authors (SDD 14 §12.5), plus the viewer's own review
// (also when their profile is private) as mine, with its Markdown source; mine is nil without one.
func (s *Follows) ItemReviews(
	ctx context.Context, viewerID int64, itemID, locale string,
) (data []Review, mine *Review, err error) {
	all, err := s.reviews(ctx, viewerID, reviewQuery(`v.item_id = $2 AND (v.user_id = $3 OR (`+visible+`))`),
		locale, itemID, viewerID)
	if err != nil {
		return nil, nil, err
	}

	ids := make([]string, len(all))
	for i := range all {
		ids[i] = all[i].ID
	}

	mineByID, err := s.myReactions(ctx, viewerID, "review", ids)
	if err != nil {
		return nil, nil, err
	}

	threads, err := s.Comments(ctx, viewerID, "review", ids)
	if err != nil {
		return nil, nil, err
	}

	data = []Review{}

	for i := range all {
		all[i].MyReaction, all[i].Comments = mineByID[all[i].ID], threads[all[i].ID]

		if all[i].Source != "" {
			mine = &all[i]
		}

		data = append(data, all[i])
	}

	return data, mine, nil
}

// UserReviews lists userID's reviews of public items for the profile tab; the caller checks the profile is visible.
func (s *Follows) UserReviews(ctx context.Context, viewerID, userID int64, locale string) ([]Review, error) {
	return s.reviews(ctx, viewerID, reviewQuery(`v.user_id = $2 AND i.visibility = 'public'`), locale, userID)
}

func (s *Follows) reviews(ctx context.Context, viewerID int64, query string, args ...any) ([]Review, error) {
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("social: reviews: %w", err)
	}
	defer rows.Close()

	out := []Review{}

	for rows.Next() {
		var (
			v      Review
			author int64
			body   string
			it     FeedItem
			cover  *string
		)

		if err := rows.Scan(&v.ID, &author, &v.Author.Nickname, &body, &v.Spoiler, &v.Helpful, &v.Likes, &v.CreatedAt,
			&v.UpdatedAt, &v.Stars, &it.ID, &it.Type, &it.Slug, &it.Title, &cover); err != nil {
			return nil, fmt.Errorf("social: review row: %w", err)
		}

		if v.Author.Nickname != "" {
			v.Author.Initial = strings.ToUpper(string([]rune(v.Author.Nickname)[:1]))
		}

		it.Cover = catalog.CoverFromThumbKey(cover)
		v.Item, v.Body = &it, Markdown(body)

		if viewerID != 0 && author == viewerID {
			v.Source = body
		}

		out = append(out, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("social: review rows: %w", err)
	}

	return out, nil
}
