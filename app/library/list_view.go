package library

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ListSummary is a row of the lists-management page; CoverThumbs holds at most four thumb keys.
type ListSummary struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Title       string   `json:"title"`
	IsPublic    bool     `json:"isPublic"`
	Hidden      bool     `json:"hidden"`
	ItemCount   int      `json:"itemCount"`
	CoverThumbs []string `json:"coverThumbs"`
}

// ListItemRef is the catalogue data shown next to a list position.
type ListItemRef struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"`
	Slug          string  `json:"slug"`
	Title         string  `json:"title"`
	Year          *int    `json:"year"`
	CoverThumbKey *string `json:"coverThumbKey"`
}

// ListEntry is one position of a list.
type ListEntry struct {
	Position int         `json:"position"`
	Note     *string     `json:"note"`
	AddedAt  time.Time   `json:"addedAt"`
	Item     ListItemRef `json:"item"`
}

// Summaries lists the user's lists, newest activity first.
func (s *Lists) Summaries(ctx context.Context, userID int64) ([]ListSummary, error) {
	rows, err := s.DB.SQL().QueryContext(ctx, `SELECT l.id, l.type, l.title, l.is_public, l.hidden, l.item_count,
  coalesce((SELECT json_agg(c.k) FROM (
    SELECT i.cover_thumb_key AS k FROM list_items li JOIN items i ON i.id = li.item_id
    WHERE li.list_id = l.id AND i.cover_thumb_key IS NOT NULL ORDER BY li.position LIMIT 4) c), '[]')::text
FROM lists l WHERE l.user_id = $1 ORDER BY l.updated_at DESC, l.id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list summaries: %w", err)
	}
	defer rows.Close()

	out := []ListSummary{}
	for rows.Next() {
		var l ListSummary
		var thumbs string
		if err := rows.Scan(&l.ID, &l.Type, &l.Title, &l.IsPublic, &l.Hidden, &l.ItemCount, &thumbs); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(thumbs), &l.CoverThumbs); err != nil {
			return nil, fmt.Errorf("list summaries thumbs: %w", err)
		}

		out = append(out, l)
	}

	return out, rows.Err()
}

// Entries returns the items of a list in order (at most MaxItemsPerList, so unpaged); hidden catalogue rows are skipped.
func (s *Lists) Entries(ctx context.Context, listID, locale string) ([]ListEntry, error) {
	rows, err := s.DB.SQL().
		QueryContext(ctx, `SELECT li.position, li.note, li.added_at, i.id, i.type, i.slug, `+libraryTitle+`,
  extract(year FROM i.release_date)::int, i.cover_thumb_key
FROM list_items li
JOIN items i ON i.id = li.item_id AND i.visibility = 'public'
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $2
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
WHERE li.list_id = $1 ORDER BY li.position`, listID, locale)
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
	}
	defer rows.Close()

	out := []ListEntry{}
	for rows.Next() {
		var e ListEntry
		if err := rows.Scan(&e.Position, &e.Note, &e.AddedAt, &e.Item.ID, &e.Item.Type, &e.Item.Slug, &e.Item.Title,
			&e.Item.Year, &e.Item.CoverThumbKey); err != nil {
			return nil, err
		}

		out = append(out, e)
	}

	return out, rows.Err()
}

// OwnerNickname is the display nickname of the user, or "" for an unknown or anonymised one; banned reports a
// current ban, which keeps the user's lists visible but noindex (SDD 14 §8).
func (s *Lists) OwnerNickname(ctx context.Context, userID int64) (nick string, banned bool, err error) {
	var n sql.NullString

	err = s.DB.SQL().QueryRowContext(ctx, `SELECT coalesce(display_username, username),
  banned_at IS NOT NULL AND (ban_expires_at IS NULL OR ban_expires_at > now())
FROM users WHERE id = $1 AND anonymized_at IS NULL`, userID).Scan(&n, &banned)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}

	return n.String, banned, err
}
