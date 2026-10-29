package accounts

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	"github.com/AlexGalhardo/vukem/database"
)

// ReplayStats counts what ReplayDeletions changed in the target.
type ReplayStats struct {
	Anonymised, Scheduled, Cancelled int
}

type deletionState struct {
	scheduled, anonymised sql.NullTime
}

// ReplayDeletions makes the deletion state of every user in target match
// source, the database that kept running after the dump (SDD 13 §8, open
// issue 11): a restored user never reappears after being anonymised, a
// scheduled deletion keeps its clock and a cancelled one stays cancelled. It
// is idempotent and sends no mail: the owners were told when it happened.
// Users missing from target were created after the dump and are skipped.
func (a *Accounts) ReplayDeletions(ctx context.Context, source, target *database.DB) (ReplayStats, error) {
	var stats ReplayStats

	want, err := deletionStates(ctx, source)
	if err != nil {
		return stats, fmt.Errorf("replay: read source: %w", err)
	}

	have, err := deletionStates(ctx, target)
	if err != nil {
		return stats, fmt.Errorf("replay: read target: %w", err)
	}

	for id, w := range want {
		h := have[id]

		switch {
		case h.anonymised.Valid:
		case w.anonymised.Valid:
			done, err := a.replayAnonymise(ctx, target, id, w)
			if err != nil {
				return stats, fmt.Errorf("replay: anonymise user %d: %w", id, err)
			}

			if done {
				stats.Anonymised++
			}
		case w.scheduled.Valid != h.scheduled.Valid || !w.scheduled.Time.Equal(h.scheduled.Time):
			n, err := a.setScheduled(ctx, target, id, w.scheduled)
			if err != nil {
				return stats, fmt.Errorf("replay: schedule user %d: %w", id, err)
			}

			stats.Scheduled += n
		}
	}

	for id, h := range have {
		if _, ok := want[id]; ok || h.anonymised.Valid || !h.scheduled.Valid {
			continue
		}

		n, err := a.setScheduled(ctx, target, id, sql.NullTime{})
		if err != nil {
			return stats, fmt.Errorf("replay: cancel user %d: %w", id, err)
		}

		stats.Cancelled += n
	}

	return stats, nil
}

// deletionStates returns the users with a scheduled or done deletion; the
// set is small (a grace period's worth plus the anonymised).
func deletionStates(ctx context.Context, db *database.DB) (map[int64]deletionState, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, deletion_scheduled_at, anonymized_at FROM users
WHERE deletion_scheduled_at IS NOT NULL OR anonymized_at IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := map[int64]deletionState{}

	for rows.Next() {
		var (
			id int64
			s  deletionState
		)

		if err := rows.Scan(&id, &s.scheduled, &s.anonymised); err != nil {
			return nil, err
		}

		out[id] = s
	}

	return out, rows.Err()
}

func (a *Accounts) replayAnonymise(ctx context.Context, target *database.DB, id int64, w deletionState) (bool, error) {
	pw, err := a.Hasher.Make(rand.Text() + rand.Text())
	if err != nil {
		return false, err
	}

	done := false

	err = target.Tx(ctx, func(tx *database.Tx) error {
		// anonymiseTx only acts on a scheduled deletion; the dump may predate the schedule.
		at := w.scheduled
		if !at.Valid {
			at = w.anonymised
		}

		res, err := tx.ExecContext(ctx, `UPDATE users SET deletion_scheduled_at = coalesce(deletion_scheduled_at, $2)
WHERE id = $1`, id, at.Time)
		if err != nil {
			return err
		}

		if n, _ := res.RowsAffected(); n == 0 {
			return nil // created after the dump
		}

		_, done, err = a.anonymiseTx(ctx, tx, id, pw, w.anonymised.Time.UTC())

		return err
	})

	return done, err
}

// setScheduled copies the source schedule (or its absence) to the target
// user and returns 1 when a row changed. A new schedule ends every session
// and token, as ScheduleDeletion does.
func (a *Accounts) setScheduled(ctx context.Context, target *database.DB, id int64, at sql.NullTime) (int, error) {
	changed := 0

	err := target.Tx(ctx, func(tx *database.Tx) error {
		res, err := tx.ExecContext(ctx, `UPDATE users SET deletion_scheduled_at = $2
WHERE id = $1 AND anonymized_at IS NULL AND deletion_scheduled_at IS DISTINCT FROM $2`, id, at)
		if err != nil {
			return err
		}

		if n, _ := res.RowsAffected(); n == 0 {
			return nil
		}

		changed = 1
		kind, data := "deletion_cancelled", map[string]any{"replayed": true}

		if at.Valid {
			if err := a.revokeAccess(ctx, tx, id); err != nil {
				return err
			}

			kind = "deletion_scheduled"
			data["purgeAt"] = DeletionDate(at.Time).Format(time.RFC3339)
		}

		return recordEvent(ctx, tx, id, nil, kind, data, a.now().UTC())
	})

	return changed, err
}
