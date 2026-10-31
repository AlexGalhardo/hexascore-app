package community

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

var (
	translationTitle = regexp.MustCompile(`^translations\.([A-Za-z-]+)\.title$`)
	synopsisField    = regexp.MustCompile(`^synopsis\.([A-Za-z-]+)$`)
	externalField    = regexp.MustCompile(`^external_ids\.([A-Za-z0-9]+)$`)
)

func fieldLocked(
	ctx context.Context,
	tx *database.Tx,
	targetType, targetID string,
	fields []string,
	now time.Time,
) error {
	rows, err := tx.QueryContext(ctx, `
SELECT field, reason FROM catalog_locks
WHERE target_type = $1 AND target_id = $2
  AND (expires_at IS NULL OR expires_at > $3)`, targetType, targetID, now.UTC())
	if err != nil {
		return fmt.Errorf("community: read catalog locks: %w", err)
	}
	defer rows.Close()
	type lock struct{ field, reason string }
	var locks []lock
	for rows.Next() {
		var item lock
		if err := rows.Scan(&item.field, &item.reason); err != nil {
			return fmt.Errorf("community: read catalog locks: %w", err)
		}
		locks = append(locks, item)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("community: read catalog locks: %w", err)
	}
	for _, item := range locks {
		if item.field != "*" && !containsField(fields, item.field) {
			continue
		}
		return &Problem{
			Status:     http.StatusLocked,
			Code:       "field-locked",
			Type:       problemBase + "field-locked",
			Field:      item.field,
			LockReason: item.reason,
		}
	}

	return nil
}

func containsField(fields []string, field string) bool {
	return slices.Contains(fields, field)
}

func staleBefore(
	ctx context.Context,
	tx *database.Tx,
	targetType, targetID string,
	changes map[string]json.RawMessage,
) error {
	for field, raw := range changes {
		before, ok := objectField(raw, "before")
		if !ok {
			continue
		}
		current, err := scalarCurrent(ctx, tx, targetType, targetID, field)
		if err != nil {
			return err
		}
		if jsonEqual(before, current) {
			continue
		}
		return &Problem{
			Status:  http.StatusConflict,
			Code:    "stale-before",
			Type:    problemBase + "stale-before",
			Field:   field,
			Current: string(current),
		}
	}

	return nil
}

func scalarCurrent(ctx context.Context, tx *database.Tx, targetType, targetID, field string) (json.RawMessage, error) {
	switch targetType {
	case "item":
		return itemCurrent(ctx, tx, targetID, field)
	case "edition":
		return editionCurrent(ctx, tx, targetID, field)
	case "entity":
		return entityCurrent(ctx, tx, targetID, field)
	default:
		return json.RawMessage("null"), nil
	}
}

func itemCurrent(ctx context.Context, tx *database.Tx, itemID, field string) (json.RawMessage, error) {
	if m := translationTitle.FindStringSubmatch(field); m != nil {
		return nullableText(
			ctx,
			tx,
			`SELECT title FROM item_translations WHERE item_id = $1 AND locale = $2`,
			itemID,
			m[1],
		)
	}
	if m := synopsisField.FindStringSubmatch(field); m != nil {
		return nullableText(
			ctx,
			tx,
			`SELECT synopsis FROM item_translations WHERE item_id = $1 AND locale = $2`,
			itemID,
			m[1],
		)
	}
	if field == "title" {
		return nullableText(ctx, tx, `SELECT original_title FROM items WHERE id = $1`, itemID)
	}
	if field == "cover" {
		return nullableText(ctx, tx, `SELECT cover_key FROM items WHERE id = $1`, itemID)
	}
	if m := externalField.FindStringSubmatch(field); m != nil {
		return nullableText(ctx, tx, `SELECT external_ids ->> $2 FROM items WHERE id = $1`, itemID, m[1])
	}
	if field != "release_date" {
		return json.RawMessage("null"), nil
	}
	var day sql.NullTime
	var precision string
	err := tx.QueryRowContext(ctx, `SELECT release_date, release_precision FROM items WHERE id = $1`, itemID).
		Scan(&day, &precision)
	if errors.Is(err, sql.ErrNoRows) || !day.Valid {
		return json.RawMessage("null"), nil
	}
	if err != nil {
		return nil, fmt.Errorf("community: read current release date: %w", err)
	}
	encoded, err := json.Marshal(struct {
		Value     string `json:"value"`
		Precision string `json:"precision"`
	}{day.Time.UTC().Format("2006-01-02"), precision})
	if err != nil {
		return nil, fmt.Errorf("community: encode current release date: %w", err)
	}

	return encoded, nil
}

func editionCurrent(ctx context.Context, tx *database.Tx, editionID, field string) (json.RawMessage, error) {
	switch field {
	case "title":
		return nullableText(ctx, tx, `SELECT title FROM editions WHERE id = $1`, editionID)
	case "cover":
		return nullableText(ctx, tx, `SELECT cover_key FROM editions WHERE id = $1`, editionID)
	}
	provider := externalField.FindStringSubmatch(field)
	if provider == nil {
		return json.RawMessage("null"), nil
	}
	column := map[string]string{"isbn13": "btrim(isbn13)", "wikidata": "wikidata_qid", "openlibrary": "openlibrary_id"}[provider[1]]
	if column == "" {
		return json.RawMessage("null"), nil
	}

	return nullableText(ctx, tx, `SELECT `+column+` FROM editions WHERE id = $1`, editionID)
}

func entityCurrent(ctx context.Context, tx *database.Tx, entityID, field string) (json.RawMessage, error) {
	switch field {
	case "title":
		return nullableText(ctx, tx, `SELECT default_name FROM entities WHERE id = $1`, entityID)
	case "cover":
		return nullableText(ctx, tx, `SELECT image_key FROM entities WHERE id = $1`, entityID)
	}
	if m := externalField.FindStringSubmatch(field); m != nil {
		return nullableText(ctx, tx, `SELECT external_ids ->> $2 FROM entities WHERE id = $1`, entityID, m[1])
	}

	return json.RawMessage("null"), nil
}

func nullableText(ctx context.Context, tx *database.Tx, query string, args ...any) (json.RawMessage, error) {
	var value sql.NullString
	err := tx.QueryRowContext(ctx, query, args...).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) || !value.Valid {
		return json.RawMessage("null"), nil
	}
	if err != nil {
		return nil, fmt.Errorf("community: read current field: %w", err)
	}
	encoded, err := json.Marshal(value.String)
	if err != nil {
		return nil, fmt.Errorf("community: encode current field: %w", err)
	}

	return encoded, nil
}

// CurrentValue reads the live value of one suggestion field for the moderation diff.
func CurrentValue(ctx context.Context, db *database.DB, targetType, targetID, field string) (json.RawMessage, error) {
	var out json.RawMessage

	err := db.Tx(ctx, func(tx *database.Tx) (err error) {
		out, err = scalarCurrent(ctx, tx, targetType, targetID, field)

		return err
	})

	return out, err
}
