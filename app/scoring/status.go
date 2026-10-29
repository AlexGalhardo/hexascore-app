package scoring

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/uuid"
)

// Statuses are the item_status values a user can set (SDD 02 §6).
var Statuses = []string{"want", "in_progress", "completed", "dropped"}

// Events dispatched after commit by the status and time-to-beat writes (SDD 02 §6).
type (
	// StatusChanged follows a status set, change or clear; Status is "" once cleared.
	StatusChanged struct {
		UserID int64
		ItemID string
		Status string
		At     time.Time
	}
	// TimeToBeatReported follows a time-to-beat upsert.
	TimeToBeatReported struct {
		UserID int64
		ItemID string
		At     time.Time
	}
)

// itemWrite is one change applied inside the item-locked transaction; it returns the events to dispatch.
type itemWrite func(ctx context.Context, tx *database.Tx, typ string, now time.Time) ([]any, error)

// mutateItem is the shared tail of the status and time-to-beat writes (SDD 02 §7.2): lock the item row, apply
// the change, recompute in the same transaction, dispatch the events after commit.
func (w *Writer) mutateItem(ctx context.Context, itemID string, apply itemWrite) (Summary, error) {
	var out Summary

	if _, err := uuid.Parse(itemID); err != nil {
		return out, errNotFound()
	}

	err := w.DB.Tx(ctx, func(tx *database.Tx) error {
		now := w.Now().UTC()

		var typ string
		var before sql.NullFloat64
		err := tx.QueryRowContext(ctx, `SELECT type, score FROM items WHERE id = $1 FOR UPDATE`, itemID).
			Scan(&typ, &before)
		if errors.Is(err, sql.ErrNoRows) {
			return errNotFound()
		}
		if err != nil {
			return err
		}

		evs, err := apply(ctx, tx, typ, now)
		if err != nil {
			return err
		}

		version, err := recomputeItem(ctx, tx, itemID, typ, now)
		if err != nil {
			return err
		}
		if out, err = summarize(ctx, tx, itemID); err != nil {
			return err
		}
		out.FormulaVersion = version

		var after sql.NullFloat64
		if err := tx.QueryRowContext(ctx, `SELECT score FROM items WHERE id = $1`, itemID).Scan(&after); err != nil {
			return err
		}

		if w.Events != nil && len(evs) > 0 {
			evs = append(
				evs,
				ItemScoreChanged{
					ItemID:         itemID,
					Before:         nullable(before),
					After:          nullable(after),
					FormulaVersion: version,
				},
			)
			tx.AfterCommit(func(ctx context.Context) {
				for _, ev := range evs {
					// A failing listener must not undo a committed write.
					_ = w.Events.Dispatch(ctx, ev)
				}
			})
		}

		return nil
	})

	return out, err
}

func insertEvent(
	ctx context.Context,
	tx *database.Tx,
	kind string,
	userID int64,
	itemID, typ string,
	at time.Time,
) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO rating_events (at, kind, user_id, item_id, item_type)
VALUES ($1, $2::rating_event_kind, $3, $4, $5)`, at, kind, userID, itemID, typ)

	return err
}

// SetStatus sets userID's status of itemID; "" clears it. Leaving "completed" deletes the time-to-beat report
// (legacy parity), and the recompute follows because the status changes the rating weight.
func (w *Writer) SetStatus(ctx context.Context, userID int64, itemID, status string) (Summary, error) {
	return w.mutateItem(
		ctx,
		itemID,
		func(ctx context.Context, tx *database.Tx, typ string, now time.Time) ([]any, error) {
			var prev string
			err := tx.QueryRowContext(ctx, `SELECT status::text FROM item_status WHERE user_id = $1 AND item_id = $2 FOR UPDATE`,
				userID, itemID).
				Scan(&prev)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}

			if status == "" {
				_, err = tx.ExecContext(
					ctx,
					`DELETE FROM item_status WHERE user_id = $1 AND item_id = $2`,
					userID,
					itemID,
				)
			} else {
				_, err = tx.ExecContext(ctx, `INSERT INTO item_status (user_id, item_id, status, updated_at)
VALUES ($1, $2, $3::item_status_value, $4)
ON CONFLICT (user_id, item_id) DO UPDATE SET status = EXCLUDED.status, updated_at = EXCLUDED.updated_at`,
					userID, itemID, status, now)
			}
			if err != nil {
				return nil, err
			}

			var evs []any
			if prev != status {
				if err := insertEvent(ctx, tx, "status_changed", userID, itemID, typ, now); err != nil {
					return nil, err
				}
				evs = append(evs, StatusChanged{UserID: userID, ItemID: itemID, Status: status, At: now})
			}

			if status != "completed" {
				if _, err := deleteTimeToBeat(ctx, tx, userID, itemID, typ, now); err != nil {
					return nil, err
				}
			}

			return evs, nil
		},
	)
}
