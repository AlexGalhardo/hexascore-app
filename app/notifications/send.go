// Package notifications is the in-app inbox (SDD 14 §5, §6.5, §7.5). Send is
// the only writer of user_notifications.
package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"
)

// Kinds mirror the table's CHECK.
const (
	SuggestionDecided  = "suggestion_decided"
	SuggestionReverted = "suggestion_reverted"
	SuggestionExpired  = "suggestion_expired"
	ReportResolved     = "report_resolved"
	RatingFlagged      = "rating_flagged"
	ListHidden         = "list_hidden"
	System             = "system"
	NewFollower        = "new_follower"
	FollowRequest      = "follow_request"
)

// PageSize is the web page size; MaxReadIDs caps one mark-read request.
const (
	PageSize   = 20
	MaxReadIDs = 100
)

// ErrNotFound answers an unknown id or one that belongs to another user.
var ErrNotFound = errors.New("notifications: not found")

// Notification is one inbox row. Payload holds ids and i18n parameters,
// never HTML.
type Notification struct {
	ID        string          `json:"id"`
	Kind      string          `json:"kind"`
	Payload   json.RawMessage `json:"payload"`
	ReadAt    *time.Time      `json:"readAt"`
	CreatedAt time.Time       `json:"createdAt"`
}

// Send inserts a notification for userID inside ex (pass the transaction of
// the write that caused it). A repeated dedupeKey for the same user is a
// no-op, so a retried job writes one row; created reports whether this call
// inserted it.
func Send(
	ctx context.Context, ex database.Executor, userID int64, kind, dedupeKey string, payload map[string]any,
) (created bool, err error) {
	if payload == nil {
		payload = map[string]any{}
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("notification payload: %w", err)
	}

	p := ex.Dialect().Placeholder
	res, err := ex.ExecContext(ctx,
		"INSERT INTO user_notifications (id, user_id, kind, payload, dedupe_key) VALUES ("+
			p(1)+", "+p(2)+", "+p(3)+", "+p(4)+", "+p(5)+") ON CONFLICT (user_id, dedupe_key) DO NOTHING",
		uuid.NewV7().String(), userID, kind, string(raw), dedupeKey)
	if err != nil {
		return false, fmt.Errorf("send notification: %w", err)
	}

	n, err := res.RowsAffected()

	return n == 1, err
}

// Page is one page of the inbox, newest first; Next is the cursor of the
// following page ("" on the last one).
type Page struct {
	Data []Notification `json:"data"`
	Next string         `json:"next"`
}

// List returns up to limit notifications of userID with id below cursor
// (UUID v7 ids sort by time). An unparsable cursor starts from the top.
func List(
	ctx context.Context, db database.Executor, userID int64, unreadOnly bool, cursor string, limit int,
) (Page, error) {
	q := `SELECT id, kind, payload, read_at, created_at FROM user_notifications WHERE user_id = $1`
	args := []any{userID}

	if unreadOnly {
		q += " AND read_at IS NULL"
	}

	if _, err := uuid.Parse(cursor); err == nil {
		args = append(args, cursor)
		q += fmt.Sprintf(" AND id < $%d", len(args))
	}

	args = append(args, limit+1)
	q += fmt.Sprintf(" ORDER BY id DESC LIMIT $%d", len(args))

	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return Page{}, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	page := Page{Data: []Notification{}}

	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.Kind, &n.Payload, &n.ReadAt, &n.CreatedAt); err != nil {
			return Page{}, fmt.Errorf("scan notification: %w", err)
		}

		page.Data = append(page.Data, n)
	}

	if err := rows.Err(); err != nil {
		return Page{}, err
	}

	if len(page.Data) > limit {
		page.Data = page.Data[:limit]
		page.Next = page.Data[limit-1].ID
	}

	return page, nil
}

// Unread counts the unread notifications of userID, stopping at 100 on the
// partial index (the badge shows "99+").
func Unread(ctx context.Context, db database.Executor, userID int64) (int, error) {
	var n int

	err := db.QueryRowContext(ctx,
		`SELECT count(*) FROM (SELECT 1 FROM user_notifications WHERE user_id = $1 AND read_at IS NULL LIMIT 100) u`,
		userID).Scan(&n)

	return n, err
}

// MarkRead marks the listed notifications of userID read. Any id that is
// not the user's (or not a UUID) fails the whole call with ErrNotFound and
// changes nothing, so ids of other users cannot be probed.
func MarkRead(ctx context.Context, db *database.DB, userID int64, ids []string) error {
	set := make(map[string]struct{}, len(ids))

	for _, id := range ids {
		u, err := uuid.Parse(id)
		if err != nil {
			return ErrNotFound
		}

		set[u.String()] = struct{}{}
	}

	// Canonical UUIDs hold only hex digits and dashes, so the literal is safe.
	literal := "{" + strings.Join(slices.Collect(maps.Keys(set)), ",") + "}"

	return db.Tx(ctx, func(tx *database.Tx) error {
		var owned int

		err := tx.QueryRowContext(ctx,
			`SELECT count(*) FROM user_notifications WHERE user_id = $1 AND id = ANY($2::uuid[])`,
			userID, literal).Scan(&owned)
		if err != nil {
			return err
		}

		if owned != len(set) {
			return ErrNotFound
		}

		_, err = tx.ExecContext(
			ctx,
			`UPDATE user_notifications SET read_at = now() WHERE user_id = $1 AND id = ANY($2::uuid[]) AND read_at IS NULL`,
			userID,
			literal,
		)

		return err
	})
}

// MarkAllRead marks every unread notification of userID read.
func MarkAllRead(ctx context.Context, db database.Executor, userID int64) error {
	_, err := db.ExecContext(ctx,
		`UPDATE user_notifications SET read_at = now() WHERE user_id = $1 AND read_at IS NULL`, userID)

	return err
}
