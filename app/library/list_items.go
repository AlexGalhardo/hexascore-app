package library

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/uuid"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// ListItemRemoved follows an item leaving a list.
type ListItemRemoved struct{ ListID, ItemID string }

// lockList re-reads the list under FOR UPDATE, filtered by the owner (defence in depth after the policy).
func lockList(ctx context.Context, tx *database.Tx, l models.List) (models.List, error) {
	return scanList(tx.QueryRowContext(ctx, `SELECT `+listCols+` FROM lists l JOIN users u ON u.id = l.user_id
WHERE l.id = $1 AND l.user_id = $2 FOR UPDATE OF l`, l.ID, l.UserID))
}

// AddItem puts itemID at the end of an already authorized list. Adding an item the list holds is a no-op
// (legacy ON CONFLICT DO NOTHING); the note is only stored on the first add.
func (s *Lists) AddItem(ctx context.Context, l models.List, itemID, note string) (models.List, error) {
	var out models.List

	err := s.DB.Tx(ctx, func(tx *database.Tx) error {
		var err error
		if l, err = lockList(ctx, tx, l); err != nil {
			return err
		}
		if err := requireItemOfType(ctx, tx, itemID, l.Type); err != nil {
			return err
		}

		var held bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM list_items WHERE list_id = $1 AND item_id = $2)`,
			l.ID, itemID).
			Scan(&held); err != nil {
			return err
		}
		if held {
			out = l

			return nil
		}
		if l.ItemCount >= MaxItemsPerList {
			return errs.NewKey(http.StatusConflict, "lists.item_limit_reached", "This list is full.", nil).
				WithExtension("code", CodeItemLimit)
		}

		now := s.Now().UTC()
		var n any
		if note != "" {
			n = note
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO list_items (list_id, item_id, position, note, added_at)
VALUES ($1, $2, (SELECT coalesce(max(position), 0) + 1 FROM list_items WHERE list_id = $1), $3, $4)`,
			l.ID, itemID, n, now); err != nil {
			return err
		}
		if out, err = touch(ctx, tx, l, 1, now); err != nil {
			return err
		}

		s.dispatch(
			tx,
			ListItemAdded{ListID: l.ID, ItemID: itemID, UserID: l.UserID, AddedAt: now, IsPublic: l.IsPublic},
		)

		return nil
	})

	return out, err
}

// RemoveItem takes itemID out of an already authorized list; a missing item is a no-op. Removal leaves a gap in
// the positions (SDD 14 §5).
func (s *Lists) RemoveItem(ctx context.Context, l models.List, itemID string) error {
	if _, err := uuid.Parse(itemID); err != nil {
		return ErrNotFound()
	}

	return s.DB.Tx(ctx, func(tx *database.Tx) error {
		l, err := lockList(ctx, tx, l)
		if err != nil {
			return err
		}

		res, err := tx.ExecContext(ctx, `DELETE FROM list_items WHERE list_id = $1 AND item_id = $2`, l.ID, itemID)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return nil
		}
		if _, err := touch(ctx, tx, l, -1, s.Now().UTC()); err != nil {
			return err
		}

		s.dispatch(tx, ListItemRemoved{ListID: l.ID, ItemID: itemID})

		return nil
	})
}

// Reorder replaces the whole order of an already authorized list atomically; it returns the new version.
func (s *Lists) Reorder(ctx context.Context, l models.List, version int, itemIDs []string) (int, error) {
	var out int

	err := s.DB.Tx(ctx, func(tx *database.Tx) error {
		var err error
		if l, err = lockList(ctx, tx, l); err != nil {
			return err
		}

		rows, err := tx.QueryContext(
			ctx,
			`SELECT item_id::text FROM list_items WHERE list_id = $1 ORDER BY position`,
			l.ID,
		)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		var current []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			current = append(current, id)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		_, err = PlanOrder(current, l.Version, version, itemIDs)
		if err != nil {
			return err
		}

		// The (list_id, position) uniqueness is deferred, so one statement can renumber every row.
		if _, err := tx.ExecContext(ctx, `UPDATE list_items li SET position = o.ord FROM
(SELECT id::uuid AS item_id, ord::int FROM unnest($2::text[]) WITH ORDINALITY AS t(id, ord)) o
WHERE li.list_id = $1 AND li.item_id = o.item_id`, l.ID, itemIDs); err != nil {
			return err
		}

		moved, err := touch(ctx, tx, l, 0, s.Now().UTC())
		out = moved.Version

		return err
	})

	return out, err
}

// touch bumps version and updated_at and moves item_count by delta, returning the stored list.
func touch(ctx context.Context, tx *database.Tx, l models.List, delta int, now time.Time) (models.List, error) {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE lists SET item_count = item_count + $2, version = version + 1, updated_at = $3
WHERE id = $1`,
		l.ID,
		delta,
		now,
	); err != nil {
		return l, err
	}

	out, err := scanList(
		tx.QueryRowContext(
			ctx,
			`SELECT `+listCols+` FROM lists l JOIN users u ON u.id = l.user_id WHERE l.id = $1`,
			l.ID,
		),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return out, ErrNotFound()
	}

	return out, err
}
