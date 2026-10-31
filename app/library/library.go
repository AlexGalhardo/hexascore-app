package library

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// LibraryPageSize is the rows per type in a library page (SDD 14 §6.2 `Scroll`, 30 per type).
const LibraryPageSize = 30

// CodeEditionNotOfWork is the 422 problem for an edition of another work (SDD 14 §6.3).
const CodeEditionNotOfWork = "edition-not-of-work"

// ItemTypes is the display order of library groups (legacy ITEM_TYPES).
var ItemTypes = []string{"game", "movie", "series", "book"}

// Statuses are the item_status values.
var Statuses = []string{"want", "in_progress", "completed", "dropped"}

// Sorts are the library sort keys; the first is the default.
var Sorts = []string{"updated", "title", "score"}

// Filter narrows the library; empty fields mean "all" (Sort empty means updated).
type Filter struct {
	Status, Type, Sort string
	// Offset pages the single type named by Type; it is ignored otherwise.
	Offset int
}

// EditionRef is a user's chosen book edition.
type EditionRef struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	ISBN13   string `json:"isbn13,omitempty"`
	Language string `json:"language"`
}

// Item is one row of a library group. Stars are 0.5–10 and TimeToBeat is the main story in minutes.
type Item struct {
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	Slug          string         `json:"slug"`
	Title         string         `json:"title"`
	Year          *int           `json:"year"`
	CoverThumbKey *string        `json:"coverThumbKey"`
	Cover         *catalog.Cover `json:"cover"`
	Status        string         `json:"status"`
	Stars         *float64       `json:"stars"`
	TimeToBeat    *int           `json:"timeToBeat"`
	Edition       *EditionRef    `json:"edition"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

// Group is the items of one type; Next is the offset of the following page (-1 on the last).
type Group struct {
	Type  string `json:"type"`
	Total int    `json:"total"`
	Items []Item `json:"items"`
	Next  int    `json:"next"`
}

// View is what GET /library and GET /app/v1/me/library answer. Counts maps status -> type -> n over the
// whole library, not the filtered one, so filters can show their sizes.
type View struct {
	Counts map[string]map[string]int `json:"counts"`
	Groups []Group                   `json:"groups"`
}

// Validate rejects unknown filter values with a 422 naming the field.
func (f Filter) Validate() error {
	v := errs.NewValidation()
	if f.Status != "" && !slices.Contains(Statuses, f.Status) {
		v.Append("status", "invalid")
	}

	if f.Type != "" && !slices.Contains(ItemTypes, f.Type) {
		v.Append("type", "invalid")
	}

	if f.Sort != "" && !slices.Contains(Sorts, f.Sort) {
		v.Append("sort", "invalid")
	}

	if v.Empty() {
		return nil
	}

	return v
}

// Library loads the library of userID: statuses joined to the user's rating, time to beat and chosen edition,
// grouped in ItemTypes order. Hidden and taken-down items are skipped (SDD 14 §8).
func Library(ctx context.Context, db *database.DB, userID int64, locale string, f Filter) (View, error) {
	view := View{Counts: map[string]map[string]int{}, Groups: []Group{}}
	if err := f.Validate(); err != nil {
		return view, err
	}

	rows, err := db.SQL().QueryContext(ctx, `SELECT s.status::text, i.type, count(*) FROM item_status s
JOIN items i ON i.id = s.item_id AND i.visibility = 'public' WHERE s.user_id = $1 GROUP BY 1, 2`, userID)
	if err != nil {
		return view, fmt.Errorf("library counts: %w", err)
	}
	defer rows.Close()

	totals := map[string]int{}
	for rows.Next() {
		var st, typ string
		var n int
		if err := rows.Scan(&st, &typ, &n); err != nil {
			return view, err
		}

		if view.Counts[st] == nil {
			view.Counts[st] = map[string]int{}
		}

		view.Counts[st][typ] = n
		if f.Status == "" || f.Status == st {
			totals[typ] += n
		}
	}

	if err := rows.Err(); err != nil {
		return view, err
	}

	for _, typ := range ItemTypes {
		if (f.Type != "" && f.Type != typ) || totals[typ] == 0 {
			continue
		}

		offset := 0
		if f.Type != "" {
			offset = max(f.Offset, 0)
		}

		g := Group{Type: typ, Total: totals[typ], Next: -1}
		if g.Items, err = libraryItems(ctx, db, userID, locale, typ, f, offset); err != nil {
			return view, err
		}

		if offset+len(g.Items) < g.Total {
			g.Next = offset + len(g.Items)
		}

		view.Groups = append(view.Groups, g)
	}

	return view, nil
}

const libraryTitle = `coalesce(t.title, te.title, i.original_title, i.slug)`

var libraryOrderBy = map[string]string{
	"updated": "s.updated_at DESC, i.id DESC",
	"title":   "lower(" + libraryTitle + ") ASC, i.id ASC",
	"score":   "r.overall DESC NULLS LAST, i.id ASC",
}

func libraryItems(
	ctx context.Context, db *database.DB, userID int64, locale, typ string, f Filter, offset int,
) ([]Item, error) {
	sortKey := f.Sort
	if sortKey == "" {
		sortKey = Sorts[0]
	}

	rows, err := db.SQL().QueryContext(ctx, `SELECT i.id, i.type, i.slug, `+libraryTitle+`,
  extract(year FROM i.release_date)::int, i.cover_thumb_key, s.status::text, round(r.overall / 10, 1)::float8,
  b.main_minutes, s.updated_at, e.id, e.title, e.isbn13, e.language
FROM item_status s
JOIN items i ON i.id = s.item_id AND i.visibility = 'public'
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $2
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
LEFT JOIN ratings r ON r.user_id = s.user_id AND r.item_id = s.item_id
LEFT JOIN time_to_beat b ON b.user_id = s.user_id AND b.item_id = s.item_id
LEFT JOIN user_item_editions ue ON ue.user_id = s.user_id AND ue.item_id = s.item_id
LEFT JOIN editions e ON e.id = ue.edition_id
WHERE s.user_id = $1 AND i.type = $3 AND ($4 = '' OR s.status::text = $4)
ORDER BY `+libraryOrderBy[sortKey]+` LIMIT $5 OFFSET $6`,
		userID, locale, typ, f.Status, LibraryPageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("library items: %w", err)
	}
	defer rows.Close()

	out := []Item{}
	for rows.Next() {
		var it Item
		var edID, edTitle, edISBN, edLang sql.NullString
		if err := rows.Scan(&it.ID, &it.Type, &it.Slug, &it.Title, &it.Year, &it.CoverThumbKey, &it.Status,
			&it.Stars, &it.TimeToBeat, &it.UpdatedAt, &edID, &edTitle, &edISBN, &edLang); err != nil {
			return nil, err
		}

		it.Cover = catalog.CoverFromThumbKey(it.CoverThumbKey)

		if edID.Valid {
			it.Edition = &EditionRef{
				ID:       edID.String,
				Title:    edTitle.String,
				ISBN13:   edISBN.String,
				Language: edLang.String,
			}
		}

		out = append(out, it)
	}

	return out, rows.Err()
}

func editionNotOfWork() error {
	return errs.NewKey(
		http.StatusUnprocessableEntity,
		"library.edition_not_of_work",
		"This edition belongs to another work.", nil).WithExtension("code", CodeEditionNotOfWork)
}

// SetEdition records editionID as userID's edition of the book itemID. An unknown or non-public item is 404;
// an edition that is unknown or belongs to another work is 422 edition-not-of-work.
func SetEdition(ctx context.Context, db *database.DB, userID int64, itemID, editionID string, now time.Time) error {
	if _, err := uuid.Parse(itemID); err != nil {
		return ErrNotFound()
	}

	var ok bool
	if err := db.SQL().QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM items WHERE id = $1 AND type = 'book' AND visibility = 'public')`,
		itemID).Scan(&ok); err != nil {
		return err
	}

	if !ok {
		return ErrNotFound()
	}

	if _, err := uuid.Parse(editionID); err != nil {
		return editionNotOfWork()
	}

	err := db.SQL().QueryRowContext(ctx, `SELECT true FROM editions WHERE id = $1 AND work_id = $2`, editionID, itemID).
		Scan(&ok)
	if errors.Is(err, sql.ErrNoRows) {
		return editionNotOfWork()
	}

	if err != nil {
		return err
	}

	_, err = db.SQL().ExecContext(ctx, `INSERT INTO user_item_editions (user_id, item_id, edition_id, updated_at)
VALUES ($1, $2, $3, $4) ON CONFLICT (user_id, item_id) DO UPDATE SET edition_id = EXCLUDED.edition_id,
updated_at = EXCLUDED.updated_at`, userID, itemID, editionID, now)

	return err
}

// ClearEdition forgets the user's chosen edition of itemID; it is a no-op when none is set.
func ClearEdition(ctx context.Context, db *database.DB, userID int64, itemID string) error {
	if _, err := uuid.Parse(itemID); err != nil {
		return ErrNotFound()
	}

	_, err := db.SQL().ExecContext(ctx, `DELETE FROM user_item_editions WHERE user_id = $1 AND item_id = $2`,
		userID, itemID)

	return err
}

// ParseOffset reads a cursor query value; anything that is not a non-negative integer is page one.
func ParseOffset(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0
	}

	return n
}
