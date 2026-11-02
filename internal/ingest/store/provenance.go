package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// MinConfidence is the floor of a field that may reach items and the public API (RFC-001 section 5 D): below it the
// value goes to field_reviews and the serialiser drops the field.
const MinConfidence = 60

// FieldProv is one entry of items.provenance_v2.
type FieldProv struct {
	Source     string    `json:"source"`
	FetchedAt  time.Time `json:"fetched_at"`
	Licence    string    `json:"licence"`
	Confidence int       `json:"confidence"`
}

// NeedsReview reports whether the value must wait for a person.
func (p FieldProv) NeedsReview() bool { return p.Confidence < MinConfidence }

// sourceBase is the base confidence per data_sources code. "llm" is the Haiku normaliser (CW-10), always 50.
var sourceBase = map[string]int{
	"wikidata": 90, "wikipedia": 85, "openlibrary": 80, "tmdb": 85, "igdb": 85, "steam": 75, "gog": 70,
	"llm": 50, "hexascore": 100,
}

// unknownSourceBase keeps a source nobody scored under the floor.
const unknownSourceBase = 50

// Signals are the evidence that moves a confidence from its source base.
type Signals struct {
	// Disagree: another source holds a different value (-10).
	Disagree bool
	// FailedDQ: the value failed a data-quality assert (-20).
	FailedDQ bool
	// Agree: two independent sources hold the same value (+10).
	Agree bool
}

// Confidence is the RFC-001 formula, clamped to 0-100.
func Confidence(source string, s Signals) int {
	c, ok := sourceBase[source]
	if !ok {
		c = unknownSourceBase
	}

	if s.Disagree {
		c -= 10
	}

	if s.FailedDQ {
		c -= 20
	}

	if s.Agree {
		c += 10
	}

	return max(0, min(100, c))
}

// SaveProvenance merges fields into items.provenance_v2 of an item.
func SaveProvenance(ctx context.Context, ex database.Executor, itemID string, fields map[string]FieldProv) error {
	b, err := json.Marshal(fields)
	if err != nil {
		return err
	}

	if _, err := ex.ExecContext(ctx, `UPDATE items SET provenance_v2 = provenance_v2 || $2::jsonb WHERE id = $1`,
		itemID, string(b)); err != nil {
		return fmt.Errorf("store: provenance v2 of %s: %w", itemID, err)
	}

	return nil
}

// Review is a value held back from items.
type Review struct {
	ItemID     string
	Field      string
	Value      json.RawMessage
	Source     string
	Confidence int
	Reason     string
}

// QueueReview opens a review, or refreshes the open one of the same item field (one open review per field).
func QueueReview(ctx context.Context, ex database.Executor, r Review) error {
	_, err := ex.ExecContext(ctx, `INSERT INTO field_reviews (item_id, field, value, source, confidence, reason)
VALUES ($1, $2, $3::jsonb, $4, $5, $6)
ON CONFLICT (item_id, field) WHERE resolved_at IS NULL
DO UPDATE SET value = EXCLUDED.value, source = EXCLUDED.source, confidence = EXCLUDED.confidence, reason = EXCLUDED.reason`,
		r.ItemID, r.Field, string(r.Value), r.Source, r.Confidence, r.Reason)
	if err != nil {
		return fmt.Errorf("store: queue review %s.%s: %w", r.ItemID, r.Field, err)
	}

	return nil
}

// OpenReview is a review row with its item.
type OpenReview struct {
	ID         int64
	ItemID     string
	ItemType   string
	ItemTitle  string
	Field      string
	Value      string
	Source     string
	Confidence int
	Reason     string
	CreatedAt  time.Time
}

// OpenReviews lists the oldest open reviews, at most limit.
func OpenReviews(ctx context.Context, ex database.Executor, limit int) ([]OpenReview, error) {
	rows, err := ex.QueryContext(
		ctx,
		`SELECT r.id, r.item_id::text, i.type, coalesce(i.original_title, i.slug), r.field,
  r.value::text, r.source, r.confidence, r.reason, r.created_at
FROM field_reviews r JOIN items i ON i.id = r.item_id
WHERE r.resolved_at IS NULL ORDER BY r.created_at, r.id LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list field reviews: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []OpenReview

	for rows.Next() {
		var r OpenReview
		if err := rows.Scan(&r.ID, &r.ItemID, &r.ItemType, &r.ItemTitle, &r.Field, &r.Value, &r.Source,
			&r.Confidence, &r.Reason, &r.CreatedAt); err != nil {
			return nil, err
		}

		out = append(out, r)
	}

	return out, rows.Err()
}

// Review resolutions.
const (
	ResolutionAccepted = "accepted"
	ResolutionRejected = "rejected"
)

var (
	// ErrReviewNotFound means no field review has that id.
	ErrReviewNotFound = errors.New("store: field review not found")
	// ErrReviewResolved means the review was already closed; resolution is final.
	ErrReviewResolved = errors.New("store: field review already resolved")
	// ErrInvalidReview means the resolution or the field cannot be applied.
	ErrInvalidReview = errors.New("store: invalid field review resolution")
)

// reviewColumns maps a reviewable field to its items column and SQL type. Only these may be written from a review;
// the column name is spliced into SQL, so it never comes from the row.
var reviewColumns = map[string]struct{ col, typ string }{
	"release_date":    {"release_date", "date"},
	"runtime_minutes": {"runtime_minutes", "integer"},
	"page_count":      {"page_count", "integer"},
	"original_title":  {"original_title", "text"},
}

// ResolveReview closes review id. Accepting writes the held value to the item, records a legacy provenance source
// and a v2 entry with confidence 100 (a person verified it); rejecting only closes the review.
func ResolveReview(ctx context.Context, db *database.DB, id int64, resolution string, by int64) error {
	if resolution != ResolutionAccepted && resolution != ResolutionRejected {
		return fmt.Errorf("%w: unknown resolution %q", ErrInvalidReview, resolution)
	}

	return db.Tx(ctx, func(tx *database.Tx) error {
		var (
			item, field, value, source string
			resolved                   bool
		)

		err := tx.QueryRowContext(ctx, `SELECT item_id::text, field, value::text, source, resolved_at IS NOT NULL
FROM field_reviews WHERE id = $1 FOR UPDATE`, id).Scan(&item, &field, &value, &source, &resolved)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrReviewNotFound
		}

		if err != nil {
			return fmt.Errorf("store: load field review %d: %w", id, err)
		}

		if resolved {
			return ErrReviewResolved
		}

		if resolution == ResolutionAccepted {
			if err := applyReview(ctx, tx, item, field, value, source); err != nil {
				return err
			}
		}

		if _, err := tx.ExecContext(
			ctx,
			`UPDATE field_reviews SET resolution = $2, resolved_by = $3, resolved_at = now()
WHERE id = $1`,
			id,
			resolution,
			by,
		); err != nil {
			return fmt.Errorf("store: close field review %d: %w", id, err)
		}

		return nil
	})
}

func applyReview(ctx context.Context, tx *database.Tx, item, field, value, source string) error {
	c, ok := reviewColumns[field]
	if !ok {
		return fmt.Errorf("%w: field %q is not writable from a review", ErrInvalidReview, field)
	}

	res, err := tx.ExecContext(ctx, `UPDATE items SET `+c.col+` = ($2::jsonb #>> '{}')::`+c.typ+`,
  provenance = provenance || jsonb_build_object($3::text, $4::text),
  provenance_v2 = provenance_v2 || jsonb_build_object($3::text, jsonb_build_object('source', $4::text,
    'fetched_at', now(), 'licence', coalesce((SELECT licence FROM data_sources WHERE code = $4), ''), 'confidence', 100)),
  updated_at = now()
WHERE id = $1`, item, value, field, source)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidReview, err)
	}

	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("%w: item %s is gone", ErrInvalidReview, item)
	}

	return nil
}
