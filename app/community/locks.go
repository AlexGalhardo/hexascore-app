package community

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

var translationTitle = regexp.MustCompile(`^translations\.([A-Za-z-]+)\.title$`)

func fieldLocked(ctx context.Context, tx *database.Tx, targetType, targetID string, fields []string, now time.Time) error {
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
	for _, item := range fields {
		if item == field {
			return true
		}
	}

	return false
}

func staleBefore(ctx context.Context, tx *database.Tx, targetType, targetID string, changes map[string]json.RawMessage) error {
	if targetType != "item" {
		return nil
	}
	for field, raw := range changes {
		before, ok := objectField(raw, "before")
		if !ok {
			continue
		}
		current, err := scalarCurrent(ctx, tx, targetID, field)
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

func scalarCurrent(ctx context.Context, tx *database.Tx, itemID, field string) (json.RawMessage, error) {
	if m := translationTitle.FindStringSubmatch(field); m != nil {
		var title string
		err := tx.QueryRowContext(ctx, `SELECT title FROM item_translations WHERE item_id = $1 AND locale = $2`, itemID, m[1]).Scan(&title)
		if errors.Is(err, sql.ErrNoRows) {
			return json.RawMessage("null"), nil
		}
		if err != nil {
			return nil, fmt.Errorf("community: read current title: %w", err)
		}
		encoded, err := json.Marshal(title)
		if err != nil {
			return nil, fmt.Errorf("community: encode current title: %w", err)
		}

		return encoded, nil
	}
	if field == "release_date" {
		var day sql.NullTime
		err := tx.QueryRowContext(ctx, `SELECT release_date FROM items WHERE id = $1`, itemID).Scan(&day)
		if errors.Is(err, sql.ErrNoRows) || !day.Valid {
			return json.RawMessage("null"), nil
		}
		if err != nil {
			return nil, fmt.Errorf("community: read current release date: %w", err)
		}
		encoded, err := json.Marshal(day.Time.UTC().Format("2006-01-02"))
		if err != nil {
			return nil, fmt.Errorf("community: encode current release date: %w", err)
		}

		return encoded, nil
	}

	return json.RawMessage("null"), nil
}
