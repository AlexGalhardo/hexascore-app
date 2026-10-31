package community

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// guardTarget rejects a duplicate, a live lock, or a stale scalar before.
// The advisory lock is per target, so two submits cannot both pass the scan.
func guardTarget(ctx context.Context, tx *database.Tx, id string, now time.Time) error {
	var targetType string
	var targetID sql.NullString
	var raw []byte
	err := tx.QueryRowContext(ctx, `SELECT target_type, target_id::text, changes FROM suggestions WHERE id = $1`, id).
		Scan(&targetType, &targetID, &raw)
	if err != nil {
		return fmt.Errorf("community: load suggestion changes: %w", err)
	}
	changes := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &changes); err != nil {
		return fmt.Errorf("community: decode changes: %w", err)
	}
	if !targetID.Valid {
		return nil
	}
	if _, err := tx.ExecContext(
		ctx,
		`SELECT pg_advisory_xact_lock(hashtext($1), hashtext($2))`,
		targetType,
		targetID.String,
	); err != nil {
		return fmt.Errorf("community: lock suggestion target: %w", err)
	}
	fields := make([]string, 0, len(changes))
	for field := range changes {
		fields = append(fields, field)
	}
	if err := fieldLocked(ctx, tx, targetType, targetID.String, fields, now); err != nil {
		return err
	}
	if err := rejectDuplicate(ctx, tx, id, targetType, targetID.String, changes); err != nil {
		return err
	}

	return staleBefore(ctx, tx, targetType, targetID.String, changes)
}

func rejectDuplicate(
	ctx context.Context,
	tx *database.Tx,
	id, targetType, targetID string,
	changes map[string]json.RawMessage,
) error {
	if len(changes) == 0 {
		return nil
	}
	var author sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT author_id FROM suggestions WHERE id = $1`, id).Scan(&author); err != nil {
		return fmt.Errorf("community: load suggestion author: %w", err)
	}
	var rivals []string
	for field, raw := range changes {
		rows, err := tx.QueryContext(ctx, `
SELECT id::text, author_id, changes -> $4
FROM suggestions
WHERE target_type = $1 AND target_id = $2
  AND status IN ('draft','submitted','in_review','approved')
  AND id <> $3
  AND changes ? $4`, targetType, targetID, id, field)
		if err != nil {
			return fmt.Errorf("community: scan duplicate suggestions: %w", err)
		}
		var ownID string
		var identical bool
		var fieldRivals []string
		for rows.Next() {
			var otherID string
			var otherAuthor sql.NullInt64
			var other []byte
			if err := rows.Scan(&otherID, &otherAuthor, &other); err != nil {
				if closeErr := finishRows(rows); closeErr != nil {
					err = errors.Join(err, closeErr)
				}

				return fmt.Errorf("community: scan duplicate suggestions: %w", err)
			}
			if author.Valid && otherAuthor.Valid && author.Int64 == otherAuthor.Int64 {
				ownID = otherID

				continue
			}
			if sameAfter(raw, other) {
				identical = true

				continue
			}
			fieldRivals = append(fieldRivals, otherID)
		}
		if scanErr := finishRows(rows); scanErr != nil {
			return fmt.Errorf("community: scan duplicate suggestions: %w", scanErr)
		}
		if ownID != "" {
			return duplicateProblem(ownID)
		}
		if identical {
			return duplicateProblem("")
		}
		rivals = append(rivals, fieldRivals...)
	}
	if len(rivals) > 0 {
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE suggestions SET meta = meta || '{"competing":true}'::jsonb WHERE id = $1`,
			id,
		); err != nil {
			return fmt.Errorf("community: flag competing suggestion: %w", err)
		}
		for _, otherID := range rivals {
			if _, err := tx.ExecContext(
				ctx,
				`UPDATE suggestions SET meta = meta || '{"competing":true}'::jsonb WHERE id = $1`,
				otherID,
			); err != nil {
				return fmt.Errorf("community: flag competing suggestion: %w", err)
			}
		}
	}

	return nil
}

func duplicateProblem(existingID string) *Problem {
	return &Problem{
		Status:     http.StatusConflict,
		Code:       "suggestion-duplicate",
		Type:       problemBase + "suggestion-duplicate",
		ExistingID: existingID,
	}
}

func sameAfter(a json.RawMessage, b []byte) bool {
	left, okLeft := objectField(a, "after")
	right, okRight := objectField(b, "after")
	if !okLeft || !okRight {
		return false
	}

	return jsonEqual(left, right)
}

func objectField(raw json.RawMessage, key string) (json.RawMessage, bool) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, false
	}
	value, ok := obj[key]

	return value, ok
}

func jsonEqual(a, b json.RawMessage) bool {
	var left, right any
	if json.Unmarshal(a, &left) != nil || json.Unmarshal(b, &right) != nil {
		return string(a) == string(b)
	}
	leftEncoded, errA := json.Marshal(left)
	rightEncoded, errB := json.Marshal(right)
	if errA != nil || errB != nil {
		return string(a) == string(b)
	}

	return string(leftEncoded) == string(rightEncoded)
}

// finishRows closes rows before the next statement on this transaction.
// A defer would keep the portal open, and the following UPDATE would fail.
func finishRows(rows *sql.Rows) error {
	err := rows.Err()
	if closeErr := rows.Close(); closeErr != nil {
		err = errors.Join(err, closeErr)
	}

	return err
}
