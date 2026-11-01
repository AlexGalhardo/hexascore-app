package social

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/events"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/scoring"
)

// Event kinds (SDD 14 FR-14-16); KindReview lives in reviews.go.
const (
	KindRating   = "rating_saved"
	KindStatus   = "status_changed"
	KindList     = "list_created"
	KindListItem = "list_item_added"
)

// FeedPageSize is the /feed page (SDD 14 FR-14-16).
const FeedPageSize = 20

// ratingWindow folds repeated star changes into one event (ADR 0021).
const ratingWindow = 10 * time.Minute

// Listeners record activity after each write commits, only while the flag is on and the actor opted in (SDD 14
// FR-14-18). They run after commit (the write paths dispatch then), so a crash in between loses one feed entry,
// never a rating.
func (s *Follows) Listeners() []events.Registration {
	return []events.Registration{
		events.On(func(ctx context.Context, e scoring.RatingCreated) error {
			return s.record(ctx, e.UserID, KindRating, e.ItemID, "", nil)
		}),
		events.On(func(ctx context.Context, e scoring.RatingUpdated) error {
			return s.record(ctx, e.UserID, KindRating, e.ItemID, "", nil)
		}),
		events.On(func(ctx context.Context, e scoring.StatusChanged) error {
			if e.Status == "" {
				return nil
			}

			return s.record(ctx, e.UserID, KindStatus, e.ItemID, "", map[string]any{"status": e.Status})
		}),
		events.On(func(ctx context.Context, e library.ListCreated) error {
			return s.record(ctx, e.UserID, KindList, "", e.ListID, nil)
		}),
		events.On(func(ctx context.Context, e library.ListItemAdded) error {
			return s.record(ctx, e.UserID, KindListItem, e.ItemID, e.ListID, nil)
		}),
	}
}

func nullable(s string) any {
	if s == "" {
		return nil
	}

	return s
}

func (s *Follows) record(
	ctx context.Context,
	actorID int64,
	kind, itemID, listID string,
	payload map[string]any,
) error {
	if !s.Enabled(ctx) {
		return nil
	}

	if payload == nil {
		payload = map[string]any{}
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("social: activity payload: %w", err)
	}

	return s.DB.Tx(ctx, func(tx *database.Tx) error {
		var optedIn bool
		if err := tx.QueryRowContext(ctx, `SELECT feed_opt_in FROM users WHERE id = $1`, actorID).
			Scan(&optedIn); err != nil || !optedIn {
			return err
		}

		if kind == KindRating {
			if _, err := tx.ExecContext(ctx, `DELETE FROM activity_events WHERE actor_id = $1 AND kind = $2
  AND item_id = $3 AND created_at > now() - make_interval(secs => $4)`,
				actorID, kind, itemID, ratingWindow.Seconds()); err != nil {
				return fmt.Errorf("social: fold rating event: %w", err)
			}
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO activity_events (id, actor_id, kind, item_id, list_id, payload)
VALUES ($1, $2, $3, $4, $5, $6)`, uuid.NewV7().String(), actorID, kind, nullable(itemID), nullable(listID),
			string(raw)); err != nil {
			return fmt.Errorf("social: record activity: %w", err)
		}

		return nil
	})
}

// Actor is the user behind a feed entry.
type Actor struct {
	Nickname string `json:"nickname"`
	Initial  string `json:"initial"`
}

// FeedItem is the item a feed entry is about.
type FeedItem struct {
	ID    string         `json:"id"`
	Type  string         `json:"type"`
	Slug  string         `json:"slug"`
	Title string         `json:"title"`
	Cover *catalog.Cover `json:"cover"`
	Stars *float64       `json:"stars,omitempty"`
}

// FeedList is the list a feed entry is about.
type FeedList struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Entry is one feed card.
type Entry struct {
	ID      string         `json:"id"`
	Kind    string         `json:"kind"`
	Actor   Actor          `json:"actor"`
	Item    *FeedItem      `json:"item,omitempty"`
	List    *FeedList      `json:"list,omitempty"`
	Payload map[string]any `json:"payload"`
	At      time.Time      `json:"at"`
}

// entryQuery selects feed cards: $1 is bound by the caller's join, $2 the locale, $3 the cursor ("" for none),
// $4 the limit. The visibility filters of SDD 14 §12.5 for items and lists apply to every caller.
func entryQuery(join string) string {
	return `SELECT e.id, e.kind, coalesce(u.display_username, u.username), e.payload,
  e.created_at, i.id, i.type, i.slug, coalesce(t.title, te.title, i.original_title, i.slug), i.cover_thumb_key,
  round(r.overall / 10, 1)::float8, l.id, l.title
FROM activity_events e
` + join + `
LEFT JOIN items i ON i.id = e.item_id
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $2
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
LEFT JOIN ratings r ON e.kind = 'rating_saved' AND r.user_id = e.actor_id AND r.item_id = e.item_id
LEFT JOIN lists l ON l.id = e.list_id
WHERE (e.item_id IS NULL OR i.visibility = 'public')
  AND (e.list_id IS NULL OR (l.is_public AND NOT l.hidden))
  AND (e.kind <> 'review_published' OR EXISTS (SELECT 1 FROM reviews v
    WHERE v.user_id = e.actor_id AND v.item_id = e.item_id AND NOT v.hidden))
  AND ($3 = '' OR e.id < $3::uuid)
ORDER BY e.id DESC LIMIT $4`
}

// Feed returns up to FeedPageSize entries of the users viewerID follows (accepted), newest first, below cursor
// (an event id; "" starts at the top), and the next cursor ("" on the last page). The privacy matrix of SDD 14
// §12.5 is applied here, at read time: opted-in, not banned, not anonymised actors; public items; public, visible
// lists.
func (s *Follows) Feed(ctx context.Context, viewerID int64, locale, cursor string) ([]Entry, string, error) {
	if _, err := uuid.Parse(cursor); err != nil {
		cursor = ""
	}

	out, err := s.entries(ctx, entryQuery(`JOIN follows f ON f.followee_id = e.actor_id
  AND f.follower_id = $1 AND f.state = 'accepted'
JOIN users u ON u.id = e.actor_id AND u.feed_opt_in AND u.anonymized_at IS NULL
  AND NOT (u.banned_at IS NOT NULL AND (u.ban_expires_at IS NULL OR u.ban_expires_at > now()))`),
		viewerID, locale, cursor, FeedPageSize+1)
	if err != nil {
		return nil, "", err
	}

	next := ""
	if len(out) > FeedPageSize {
		out = out[:FeedPageSize]
		next = out[FeedPageSize-1].ID
	}

	return out, next, nil
}

// ActivityAllowed reports whether viewerID may see the "recent activity" tab of userID (SDD 14 FR-14-17, §12.5):
// never while the layer is off; the owner always; others only while the owner is opted in and, for a private
// profile, when they are an accepted follower. Ban and anonymisation are handled by the profile page, which renders
// nothing social for those users.
func (s *Follows) ActivityAllowed(ctx context.Context, viewerID, userID int64) (bool, error) {
	if !s.Enabled(ctx) {
		return false, nil
	}

	if viewerID == userID {
		return true, nil
	}

	var allowed bool

	err := s.DB.QueryRowContext(ctx, `SELECT u.feed_opt_in AND (NOT u.profile_private OR EXISTS (
  SELECT 1 FROM follows f WHERE f.follower_id = $2 AND f.followee_id = u.id AND f.state = 'accepted'))
FROM users u WHERE u.id = $1`, userID, viewerID).Scan(&allowed)
	if err != nil {
		return false, fmt.Errorf("social: activity access: %w", err)
	}

	return allowed, nil
}

// Activity is userID's latest FeedPageSize entries; call ActivityAllowed first.
func (s *Follows) Activity(ctx context.Context, userID int64, locale string) ([]Entry, error) {
	return s.entries(ctx, entryQuery(`JOIN users u ON u.id = e.actor_id AND e.actor_id = $1`),
		userID, locale, "", FeedPageSize)
}

// SetFeedOptIn records the user's choice (SDD 14 FR-14-18); turning it off hides past events at read time.
func (s *Follows) SetFeedOptIn(ctx context.Context, userID int64, on bool) error {
	if _, err := s.DB.ExecContext(ctx, `UPDATE users SET feed_opt_in = $2 WHERE id = $1`, userID, on); err != nil {
		return fmt.Errorf("social: feed opt-in: %w", err)
	}

	return nil
}

// FeedOptIn reports the user's current choice.
func (s *Follows) FeedOptIn(ctx context.Context, userID int64) (bool, error) {
	var on bool
	if err := s.DB.QueryRowContext(ctx, `SELECT feed_opt_in FROM users WHERE id = $1`, userID).Scan(&on); err != nil {
		return false, fmt.Errorf("social: feed opt-in: %w", err)
	}

	return on, nil
}

// ActivityRetention is how long activity events are kept (SDD 14 NFR-14-9).
const ActivityRetention = 180 * 24 * time.Hour

// PruneActivity deletes events older than ActivityRetention at now; it runs whatever the flag, so turning the
// layer off never keeps data longer.
func PruneActivity(ctx context.Context, db *database.DB, now time.Time) (int64, error) {
	res, err := db.ExecContext(ctx, `DELETE FROM activity_events WHERE created_at < $1`, now.Add(-ActivityRetention))
	if err != nil {
		return 0, fmt.Errorf("social: prune activity: %w", err)
	}

	return res.RowsAffected()
}

func (s *Follows) entries(ctx context.Context, query string, args ...any) ([]Entry, error) {
	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("social: feed: %w", err)
	}
	defer rows.Close()

	out := []Entry{}

	for rows.Next() {
		var (
			e                               Entry
			payload                         []byte
			itemID, typ, slug, title, cover *string
			stars                           *float64
			listID, listTitle               *string
		)

		if err := rows.Scan(&e.ID, &e.Kind, &e.Actor.Nickname, &payload, &e.At, &itemID, &typ, &slug, &title, &cover,
			&stars, &listID, &listTitle); err != nil {
			return nil, fmt.Errorf("social: feed row: %w", err)
		}

		if err := json.Unmarshal(payload, &e.Payload); err != nil {
			return nil, fmt.Errorf("social: feed payload: %w", err)
		}

		if e.Actor.Nickname != "" {
			e.Actor.Initial = strings.ToUpper(string([]rune(e.Actor.Nickname)[:1]))
		}

		if itemID != nil {
			e.Item = &FeedItem{ID: *itemID, Type: *typ, Slug: *slug, Title: *title, Stars: stars}
			e.Item.Cover = catalog.CoverFromThumbKey(cover)
		}

		if listID != nil {
			e.List = &FeedList{ID: *listID, Title: *listTitle}
		}

		out = append(out, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("social: feed rows: %w", err)
	}

	return out, nil
}
