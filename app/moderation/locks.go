package moderation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/uuid"
)

// ErrLockExists is a second lock on the same target and field.
var ErrLockExists = errors.New("moderation: lock already exists")

// ErrTargetNotFound is a lock on a catalogue row that does not exist.
var ErrTargetNotFound = errors.New("moderation: lock target not found")

var lockTables = map[string]string{"item": "items", "entity": "entities", "edition": "editions"}

// Lock is one catalog_locks row; Field "*" locks the whole record (SDD 04 FR-04-7).
type Lock struct {
	ID         string    `json:"id"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	Field      string    `json:"field"`
	Reason     string    `json:"reason"`
	LockedBy   int64     `json:"locked_by"`
	ExpiresAt  string    `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// NewLock is the input of CreateLock; a zero TTL never expires.
type NewLock struct {
	TargetType, TargetID, Field, Reason string
	TTL                                 time.Duration
	By                                  int64
}

// Locks lists the active locks (expired ones are ignored by the write path and hidden here), newest first.
func (q Queue) Locks(ctx context.Context) ([]Lock, error) {
	var raw []byte

	err := q.DB.QueryRowContext(ctx, `SELECT COALESCE(json_agg(row_to_json(t) ORDER BY t.created_at DESC), '[]') FROM (
SELECT id::text AS id, target_type::text AS target_type, target_id::text AS target_id, field, reason,
  COALESCE(locked_by, 0) AS locked_by, COALESCE(to_char(expires_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI'), '')
  AS expires_at, created_at
FROM catalog_locks WHERE expires_at IS NULL OR expires_at > $1) t`, q.now().UTC()).Scan(&raw)
	if err != nil {
		return nil, fmt.Errorf("moderation: list locks: %w", err)
	}

	var locks []Lock
	if err := json.Unmarshal(raw, &locks); err != nil {
		return nil, fmt.Errorf("moderation: decode locks: %w", err)
	}

	return locks, nil
}

// CreateLock stores a lock after checking the target exists. An expired lock on the same target and field is
// replaced; an active one is ErrLockExists.
func (q Queue) CreateLock(ctx context.Context, in NewLock) (string, error) {
	table := lockTables[in.TargetType]
	if table == "" {
		return "", &InvalidError{Field: "target", Msg: "Choose item, entity or edition."}
	}

	if _, err := uuid.Parse(in.TargetID); err != nil {
		return "", &InvalidError{Field: "target_id", Msg: "The target id must be a uuid."}
	}

	in.Field = strings.TrimSpace(in.Field)
	if in.Field == "" {
		in.Field = "*"
	}

	in.Reason = strings.TrimSpace(in.Reason)
	if in.Reason == "" || len(in.Reason) > 500 {
		return "", &InvalidError{Field: "reason", Msg: "Give a reason (up to 500 characters)."}
	}

	var exists bool
	if err := q.DB.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM `+table+` WHERE id = $1)`, in.TargetID).
		Scan(&exists); err != nil {
		return "", fmt.Errorf("moderation: lock target: %w", err)
	}

	if !exists {
		return "", ErrTargetNotFound
	}

	now := q.now().UTC()

	var expires sql.NullTime
	if in.TTL > 0 {
		expires = sql.NullTime{Time: now.Add(in.TTL), Valid: true}
	}

	var got string

	err := q.DB.QueryRowContext(ctx, `INSERT INTO catalog_locks
(id, target_type, target_id, field, reason, locked_by, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (target_type, target_id, field) DO UPDATE SET id = EXCLUDED.id, reason = EXCLUDED.reason,
  locked_by = EXCLUDED.locked_by, expires_at = EXCLUDED.expires_at, created_at = EXCLUDED.created_at
  WHERE catalog_locks.expires_at IS NOT NULL AND catalog_locks.expires_at <= EXCLUDED.created_at
RETURNING id::text`, uuid.NewV7().String(), in.TargetType, in.TargetID, in.Field, in.Reason, in.By, expires, now).Scan(&got)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrLockExists
	}

	if err != nil {
		return "", fmt.Errorf("moderation: create lock: %w", err)
	}

	return got, nil
}

// DeleteLock removes a lock; an unknown id is ErrNotFound.
func (q Queue) DeleteLock(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrNotFound
	}

	res, err := q.DB.ExecContext(ctx, `DELETE FROM catalog_locks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("moderation: delete lock: %w", err)
	}

	if n, err := res.RowsAffected(); err != nil || n == 0 {
		return ErrNotFound
	}

	return nil
}
