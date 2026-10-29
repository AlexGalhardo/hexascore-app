package scoring

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/uuid"
)

// Problem codes of the time-to-beat writes (SDD 02 §6).
const (
	CodeStatusNotCompleted = "status-not-completed"
	CodeNotAGame           = "not-a-game"
)

// TimeToBeatHours is one user's report in hours (0.5 steps, 0.5–10,000); nil styles are not reported.
type TimeToBeatHours struct {
	Main          *float64 `json:"mainHours"`
	Extras        *float64 `json:"extrasHours"`
	Completionist *float64 `json:"completionistHours"`
}

func minutes(h *float64) any {
	if h == nil {
		return nil
	}

	return int(math.Round(*h * 60))
}

// SaveTimeToBeat upserts userID's report. Only games qualify (422 not-a-game) and only while the user's status
// is "completed" (409 status-not-completed). Lock order: item row, then the status row (SDD 02 §7.2).
func (w *Writer) SaveTimeToBeat(ctx context.Context, userID int64, itemID string, h TimeToBeatHours) (Summary, error) {
	return w.mutateItem(
		ctx,
		itemID,
		func(ctx context.Context, tx *database.Tx, typ string, now time.Time) ([]any, error) {
			if typ != "game" {
				return nil, errs.NewKey(http.StatusUnprocessableEntity, "time_to_beat.not_a_game",
					"Time to beat is only available for games.", nil).WithExtension("code", CodeNotAGame)
			}

			var status string
			err := tx.QueryRowContext(ctx, `SELECT status::text FROM item_status WHERE user_id = $1 AND item_id = $2 FOR UPDATE`,
				userID, itemID).
				Scan(&status)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
			if status != "completed" {
				return nil, errs.NewKey(
					http.StatusConflict,
					"time_to_beat.status_not_completed",
					"Mark the game as completed before reporting a time.",
					nil,
				).WithExtension("code", CodeStatusNotCompleted)
			}

			if _, err := tx.ExecContext(
				ctx,
				`INSERT INTO time_to_beat (user_id, item_id, main_minutes, extras_minutes, completionist_minutes, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $6)
ON CONFLICT (user_id, item_id) DO UPDATE SET main_minutes = EXCLUDED.main_minutes,
  extras_minutes = EXCLUDED.extras_minutes, completionist_minutes = EXCLUDED.completionist_minutes, updated_at = EXCLUDED.updated_at`,
				userID,
				itemID,
				minutes(h.Main),
				minutes(h.Extras),
				minutes(h.Completionist),
				now,
			); err != nil {
				return nil, err
			}
			if err := insertEvent(ctx, tx, "time_to_beat_reported", userID, itemID, typ, now); err != nil {
				return nil, err
			}

			return []any{TimeToBeatReported{UserID: userID, ItemID: itemID, At: now}}, nil
		},
	)
}

// DeleteTimeToBeat removes userID's report; a missing report is a no-op.
func (w *Writer) DeleteTimeToBeat(ctx context.Context, userID int64, itemID string) (Summary, error) {
	return w.mutateItem(
		ctx,
		itemID,
		func(ctx context.Context, tx *database.Tx, typ string, now time.Time) ([]any, error) {
			_, err := deleteTimeToBeat(ctx, tx, userID, itemID, typ, now)

			return nil, err
		},
	)
}

func deleteTimeToBeat(
	ctx context.Context,
	tx *database.Tx,
	userID int64,
	itemID, typ string,
	now time.Time,
) (bool, error) {
	res, err := tx.ExecContext(ctx, `DELETE FROM time_to_beat WHERE user_id = $1 AND item_id = $2`, userID, itemID)
	if err != nil {
		return false, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return false, nil
	}

	return true, insertEvent(ctx, tx, "time_to_beat_cleared", userID, itemID, typ, now)
}

// Me is the signed-in user's own state on an item (GET /app/v1/items/{id}/me).
type Me struct {
	Rating     map[string]float64 `json:"rating"`
	Status     *string            `json:"status"`
	TimeToBeat *TimeToBeatHours   `json:"timeToBeat"`
}

// Me reads the user's rating (stars per category key), status and time-to-beat of itemID.
func (w *Writer) Me(ctx context.Context, userID int64, itemID string) (Me, error) {
	var out Me

	typ, err := w.itemType(ctx, itemID)
	if err != nil {
		return out, err
	}

	var c [Categories]int16
	err = w.DB.SQL().
		QueryRowContext(ctx, `SELECT c1, c2, c3, c4, c5, c6 FROM ratings WHERE user_id = $1 AND item_id = $2`, userID, itemID).
		Scan(&c[0], &c[1], &c[2], &c[3], &c[4], &c[5])
	switch {
	case err == nil:
		out.Rating = map[string]float64{}
		for i, k := range CategoryKeys[typ] {
			out.Rating[k] = float64(c[i]) / 10
		}
	case !errors.Is(err, sql.ErrNoRows):
		return out, err
	}

	var st string
	err = w.DB.SQL().
		QueryRowContext(ctx, `SELECT status::text FROM item_status WHERE user_id = $1 AND item_id = $2`, userID, itemID).
		Scan(&st)
	switch {
	case err == nil:
		out.Status = &st
	case !errors.Is(err, sql.ErrNoRows):
		return out, err
	}

	var m, e, cm sql.NullInt64
	err = w.DB.SQL().QueryRowContext(ctx, `SELECT main_minutes, extras_minutes, completionist_minutes FROM time_to_beat
WHERE user_id = $1 AND item_id = $2`, userID, itemID).Scan(&m, &e, &cm)
	switch {
	case err == nil:
		hours := func(n sql.NullInt64) *float64 {
			if !n.Valid {
				return nil
			}
			v := float64(n.Int64) / 60

			return &v
		}
		out.TimeToBeat = &TimeToBeatHours{Main: hours(m), Extras: hours(e), Completionist: hours(cm)}
	case !errors.Is(err, sql.ErrNoRows):
		return out, err
	}

	return out, nil
}

func (w *Writer) itemType(ctx context.Context, itemID string) (string, error) {
	if _, err := uuid.Parse(itemID); err != nil {
		return "", errNotFound()
	}

	var typ string
	err := w.DB.SQL().QueryRowContext(ctx, `SELECT type FROM items WHERE id = $1`, itemID).Scan(&typ)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errNotFound()
	}

	return typ, err
}
