// Package library holds the user library: lists and their items (SDD 14).
package library

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/events"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Limits (SDD 14 FR-14-2, legacy LIST_LIMITS).
const (
	MaxListsPerType = 10
	MaxItemsPerList = 500
	MaxTitle        = 80
	MaxDescription  = 500
	MaxNote         = 280
)

// Problem codes (SDD 14 §6.3).
const (
	CodeListLimit    = "list-limit-reached"
	CodeTypeMismatch = "list-type-mismatch"
)

// Events dispatched after commit (SDD 14 §6.5).
type (
	// ListItemAdded follows an item landing in a list.
	ListItemAdded struct {
		ListID, ItemID string
		UserID         int64
		AddedAt        time.Time
		IsPublic       bool
	}
	// ListVisibilityChanged follows a change of is_public; Visible is is_public && !hidden.
	ListVisibilityChanged struct {
		ListID  string
		Visible bool
	}
	// ListDeleted follows a delete.
	ListDeleted struct{ ListID string }
)

// Lists is the list write path.
type Lists struct {
	DB     *database.DB
	Events *events.Bus
	Now    func() time.Time
}

// CreateInput is a validated new list; FirstItemID is optional.
type CreateInput struct {
	Type        string
	Title       string
	Description string
	IsPublic    bool
	FirstItemID string
}

// UpdateInput patches a list; nil fields stay, an empty Description clears it.
type UpdateInput struct {
	Title       *string
	Description *string
	IsPublic    *bool
}

// ErrNotFound answers an unknown list or item as 404.
func ErrNotFound() error {
	return errs.NewKey(http.StatusNotFound, "lists.not_found", "Not found.", map[string]any{"code": "not-found"})
}

const listCols = `l.id, l.user_id, l.type, l.title, l.description, l.is_public, l.hidden, l.item_count, l.version,
l.created_at, l.updated_at, u.profile_private`

type scanner interface{ Scan(dest ...any) error }

func scanList(row scanner) (models.List, error) {
	var l models.List
	err := row.Scan(&l.ID, &l.UserID, &l.Type, &l.Title, &l.Description, &l.IsPublic, &l.Hidden, &l.ItemCount,
		&l.Version, &l.CreatedAt, &l.UpdatedAt, &l.OwnerPrivate)
	if errors.Is(err, sql.ErrNoRows) {
		return l, ErrNotFound()
	}

	return l, err
}

// Get loads a list with its owner's privacy flag; callers authorize the result before using it.
func (s *Lists) Get(ctx context.Context, id string) (models.List, error) {
	if _, err := uuid.Parse(id); err != nil {
		return models.List{}, ErrNotFound()
	}

	return scanList(s.DB.SQL().QueryRowContext(ctx,
		`SELECT `+listCols+` FROM lists l JOIN users u ON u.id = l.user_id WHERE l.id = $1`, id))
}

// Create inserts a list for userID, enforcing the per-type limit under a per-user advisory lock (SDD 14 §7.1).
func (s *Lists) Create(ctx context.Context, userID int64, in CreateInput) (models.List, error) {
	var out models.List
	id := uuid.NewV7().String()

	err := s.DB.Tx(ctx, func(tx *database.Tx) error {
		now := s.Now().UTC()

		if _, err := tx.ExecContext(
			ctx,
			`SELECT pg_advisory_xact_lock(hashtext('lists:' || $1::bigint::text))`,
			userID,
		); err != nil {
			return err
		}

		var n int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM lists WHERE user_id = $1 AND type = $2`, userID, in.Type).
			Scan(&n); err != nil {
			return err
		}
		if n >= MaxListsPerType {
			return errs.NewKey(http.StatusConflict, "lists.limit_reached", "List limit reached for this type.",
				map[string]any{"code": CodeListLimit})
		}

		if in.FirstItemID != "" {
			if err := requireItemOfType(ctx, tx, in.FirstItemID, in.Type); err != nil {
				return err
			}
		}

		var desc any
		if in.Description != "" {
			desc = in.Description
		}
		count := 0
		if in.FirstItemID != "" {
			count = 1
		}
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO lists (id, user_id, type, title, description, is_public, item_count, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $7, $8, $8)`,
			id,
			userID,
			in.Type,
			in.Title,
			desc,
			in.IsPublic,
			count,
			now,
		); err != nil {
			return err
		}
		if in.FirstItemID != "" {
			if _, err := tx.ExecContext(
				ctx,
				`INSERT INTO list_items (list_id, item_id, position, added_at) VALUES ($1, $2, 1, $3)`,
				id,
				in.FirstItemID,
				now,
			); err != nil {
				return err
			}
		}

		var err error
		if out, err = scanList(tx.QueryRowContext(ctx,
			`SELECT `+listCols+` FROM lists l JOIN users u ON u.id = l.user_id WHERE l.id = $1`, id)); err != nil {
			return err
		}

		if in.FirstItemID != "" {
			s.dispatch(
				tx,
				ListItemAdded{ListID: id, ItemID: in.FirstItemID, UserID: userID, AddedAt: now, IsPublic: in.IsPublic},
			)
		}

		return nil
	})

	return out, err
}

// requireItemOfType answers 404 for an unknown or non-public item and 422 list-type-mismatch for another type.
func requireItemOfType(ctx context.Context, tx *database.Tx, itemID, listType string) error {
	if _, err := uuid.Parse(itemID); err != nil {
		return ErrNotFound()
	}

	var typ string
	err := tx.QueryRowContext(ctx, `SELECT type FROM items WHERE id = $1 AND visibility = 'public'`, itemID).Scan(&typ)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound()
	}
	if err != nil {
		return err
	}
	if typ != listType {
		return errs.NewKey(
			http.StatusUnprocessableEntity,
			"lists.type_mismatch",
			"This list holds another type of item.",
			map[string]any{"code": CodeTypeMismatch},
		)
	}

	return nil
}

// Update applies in to an already authorized list; the user_id filter is defence in depth.
func (s *Lists) Update(ctx context.Context, l models.List, in UpdateInput) (models.List, error) {
	var out models.List

	err := s.DB.Tx(ctx, func(tx *database.Tx) error {
		var desc any
		setDesc := in.Description != nil
		if setDesc && *in.Description != "" {
			desc = *in.Description
		}

		if _, err := tx.ExecContext(ctx, `UPDATE lists SET title = COALESCE($3, title),
  description = CASE WHEN $4::boolean THEN $5 ELSE description END,
  is_public = COALESCE($6, is_public), updated_at = $7
WHERE id = $1 AND user_id = $2`, l.ID, l.UserID, in.Title, setDesc, desc, in.IsPublic, s.Now().UTC()); err != nil {
			return err
		}

		var err error
		if out, err = scanList(tx.QueryRowContext(ctx,
			`SELECT `+listCols+` FROM lists l JOIN users u ON u.id = l.user_id WHERE l.id = $1`, l.ID)); err != nil {
			return err
		}
		if out.IsPublic != l.IsPublic {
			s.dispatch(tx, ListVisibilityChanged{ListID: l.ID, Visible: out.IsPublic && !out.Hidden})
		}

		return nil
	})

	return out, err
}

// Delete removes an already authorized list and, by cascade, its items.
func (s *Lists) Delete(ctx context.Context, l models.List) error {
	return s.DB.Tx(ctx, func(tx *database.Tx) error {
		if _, err := tx.ExecContext(
			ctx,
			`DELETE FROM lists WHERE id = $1 AND user_id = $2`,
			l.ID,
			l.UserID,
		); err != nil {
			return err
		}
		s.dispatch(tx, ListDeleted{ListID: l.ID})

		return nil
	})
}

func (s *Lists) dispatch(tx *database.Tx, ev any) {
	if s.Events == nil {
		return
	}

	tx.AfterCommit(func(ctx context.Context) {
		// A failing listener must not undo a committed write.
		_ = s.Events.Dispatch(ctx, ev)
	})
}
