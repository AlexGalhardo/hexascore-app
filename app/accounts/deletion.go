package accounts

import (
	"context"
	"crypto/rand"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/orm"

	"github.com/AlexGalhardo/hexascore-app/app/mail"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// DeletionGraceDays is how long a scheduled deletion can still be cancelled
// (legacy DELETION_GRACE_DAYS, SDD 03 section 7.5).
const DeletionGraceDays = 30

// PurgeBatch is how many due accounts one purge run anonymises.
const PurgeBatch = 100

// ErrNotScheduled is returned when cancelling an account that is not scheduled for deletion.
var ErrNotScheduled = errors.New("accounts: deletion is not scheduled")

// SubscriptionCanceller ends a user's live paid subscription before the
// account is anonymised. CancelLive returns nil when there is nothing to
// cancel; an error keeps the user scheduled and the purge retries next run, so
// a deleted account is never billed again. billing.Canceller implements it for Stripe.
type SubscriptionCanceller interface {
	CancelLive(ctx context.Context, userID int64) error
}

// DeletionDate is when a deletion scheduled at scheduledAt takes effect.
func DeletionDate(scheduledAt time.Time) time.Time {
	return scheduledAt.AddDate(0, 0, DeletionGraceDays)
}

// ScheduleDeletion starts the grace period: it ends every session and token,
// records the event and mails the owner. It returns when the deletion takes
// effect. Scheduling again keeps the original clock.
func (a *Accounts) ScheduleDeletion(ctx context.Context, userID int64) (time.Time, error) {
	now := a.now().UTC()
	scheduled := now
	fresh := false

	var u models.User

	err := a.DB.Tx(ctx, func(tx *database.Tx) error {
		var err error

		u, err = orm.QueryOf[models.User](tx).Where("id", userID).LockForUpdate().First(ctx)
		if err != nil {
			return err
		}

		if u.AnonymizedAt.Valid {
			return errs.New(http.StatusGone, a.T(ctx, "auth.deletion.gone"))
		}

		if u.DeletionScheduledAt.Valid {
			scheduled = u.DeletionScheduledAt.V

			return nil
		}

		fresh = true

		set := orm.Set{"deletion_scheduled_at": now}
		if _, err := orm.QueryOf[models.User](tx).Where("id", userID).Update(ctx, set); err != nil {
			return err
		}

		if err := a.revokeAccess(ctx, tx, userID); err != nil {
			return err
		}

		data := map[string]any{"purgeAt": DeletionDate(now).Format(time.RFC3339)}

		return recordEvent(ctx, tx, userID, nil, "deletion_scheduled", data, now)
	})
	if err != nil || !fresh {
		return DeletionDate(scheduled), err
	}

	// After the commit: the session store may share SQLite's one writer.
	if err := a.revokeServerSessions(ctx, userID, ""); err != nil {
		a.log.ErrorContext(ctx, "deletion: revoke server sessions", "user", userID, "err", err)
	}

	err = a.Mailer.SendLater(ctx, mail.DeletionScheduled{
		User: u, Date: DeletionDate(now), CancelLink: a.URL + "/login",
		Strings: a.mailStrings(
			ctx, u, "deletion_scheduled", noticeFields, "date", DeletionDate(now).Format("2006-01-02")),
	})

	return DeletionDate(now), err
}

// CancelDeletion ends the grace period with the data intact. Tokens stay
// revoked, so every device signs in again.
func (a *Accounts) CancelDeletion(ctx context.Context, userID int64) error {
	now := a.now().UTC()

	return a.DB.Tx(ctx, func(tx *database.Tx) error {
		u, err := orm.QueryOf[models.User](tx).Where("id", userID).LockForUpdate().First(ctx)
		if err != nil {
			return err
		}

		if !u.DeletionScheduledAt.Valid || u.AnonymizedAt.Valid {
			return ErrNotScheduled
		}

		if _, err := orm.QueryOf[models.User](tx).Where("id", userID).
			Update(ctx, orm.Set{"deletion_scheduled_at": nil}); err != nil {
			return err
		}

		return recordEvent(ctx, tx, userID, nil, "deletion_cancelled", map[string]any{}, now)
	})
}

// PurgeDue anonymises up to PurgeBatch accounts whose grace period ended and
// returns their ids. A failing subscription cancellation keeps that user
// scheduled for the next run; the others still go. Idempotent: anonymised
// users are never selected again.
func (a *Accounts) PurgeDue(ctx context.Context) ([]int64, error) {
	cutoff := a.now().UTC().AddDate(0, 0, -DeletionGraceDays)
	p := a.DB.Dialect().Placeholder

	rows, err := a.DB.QueryContext(
		ctx,
		"SELECT id FROM users WHERE deletion_scheduled_at IS NOT NULL AND anonymized_at IS NULL AND deletion_scheduled_at < "+
			p(
				1,
			)+" ORDER BY deletion_scheduled_at, id LIMIT "+strconv.Itoa(
			PurgeBatch,
		),
		cutoff,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var due []int64

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		due = append(due, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var done []int64

	for _, id := range due {
		if err := ctx.Err(); err != nil {
			return done, err
		}

		if err := a.Subscriptions.CancelLive(ctx, id); err != nil {
			a.log.ErrorContext(ctx, "purge: subscription not cancelled, retrying next run", "user", id, "err", err)

			continue
		}

		ok, err := a.anonymise(ctx, id)
		if err != nil {
			a.log.ErrorContext(ctx, "purge: anonymise failed, retrying next run", "user", id, "err", err)

			continue
		}

		if ok {
			done = append(done, id)
		}
	}

	return done, nil
}

// anonymise keeps the row (ratings keep their user_id) but leaves nothing
// that identifies the person and no credential that still works (SDD 03
// section 5 invariants). It reports false when the user cancelled meanwhile.
func (a *Accounts) anonymise(ctx context.Context, userID int64) (bool, error) {
	pw, err := a.Hasher.Make(rand.Text() + rand.Text())
	if err != nil {
		return false, err
	}

	now := a.now().UTC()
	done := false

	var old models.User

	err = a.DB.Tx(ctx, func(tx *database.Tx) error {
		var err error

		old, done, err = a.anonymiseTx(ctx, tx, userID, pw, now)

		return err
	})
	if err != nil || !done {
		return false, err
	}

	if err := a.revokeServerSessions(ctx, userID, ""); err != nil {
		a.log.ErrorContext(ctx, "purge: revoke server sessions", "user", userID, "err", err)
	}

	err = a.Mailer.SendLater(ctx, mail.AccountAnonymized{
		User: old, Strings: a.mailStrings(ctx, old, "account_anonymized", bodyFields),
	})
	if err != nil {
		a.log.ErrorContext(ctx, "purge: notice not queued", "user", userID, "err", err)
	}

	return true, nil
}

// anonymiseTx is the database part of anonymise: it rewrites the user row of
// a user whose deletion is scheduled and drops what identifies them. Callers
// handle mail and the session store; ReplayDeletions runs it on a restored
// database with password a fresh random hash and now the original time.
func (a *Accounts) anonymiseTx(
	ctx context.Context,
	tx *database.Tx,
	userID int64,
	pw string,
	now time.Time,
) (models.User, bool, error) {
	old, err := orm.QueryOf[models.User](tx).Where("id", userID).LockForUpdate().First(ctx)
	if err != nil {
		return old, false, err
	}

	if !old.DeletionScheduledAt.Valid || old.AnonymizedAt.Valid {
		return old, false, nil
	}

	p := tx.Dialect().Placeholder
	res, err := tx.ExecContext(ctx, `UPDATE users SET email = 'deleted-' || id || '@invalid', name = 'Deleted user',
  username = NULL, display_username = NULL, image_url = NULL, email_verified_at = NULL, password = `+p(1)+`,
  profile_private = `+p(2)+`, two_factor_secret = NULL, two_factor_recovery_codes = NULL,
  two_factor_confirmed_at = NULL, two_factor_last_step = 0, deletion_scheduled_at = NULL,
  anonymized_at = `+p(3)+`, updated_at = `+p(3)+` WHERE id = `+p(4), pw, true, now, userID)
	if err != nil {
		return old, false, err
	}

	if n, _ := res.RowsAffected(); n == 0 {
		return old, false, nil
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM social_accounts WHERE user_id = "+p(1), userID); err != nil {
		return old, false, err
	}

	// The user row is anonymised, not deleted, so ON DELETE CASCADE does not remove passkeys.
	if _, err := tx.ExecContext(ctx, "DELETE FROM passkeys WHERE user_id = "+p(1), userID); err != nil {
		return old, false, err
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM password_resets WHERE email = "+p(1), old.Email); err != nil {
		return old, false, err
	}

	if err := a.revokeAccess(ctx, tx, userID); err != nil {
		return old, false, err
	}

	if err := purgeUserLists(ctx, tx, userID); err != nil {
		return old, false, err
	}

	if tx.Dialect().Name() == "postgres" {
		if _, err := tx.ExecContext(
			ctx,
			"DELETE FROM user_item_editions WHERE user_id = "+p(1),
			userID,
		); err != nil {
			return old, false, err
		}
	}

	return old, true, recordEvent(ctx, tx, userID, nil, "anonymized", map[string]any{}, now)
}

// purgeUserLists applies the architect default of open issue 14-7: public
// lists of a deleted user are anonymised, private ones deleted. The lists
// tables belong to SDD 14 and do not exist yet, so there is nothing to
// rewrite; the owner's answer to 14-7 changes only this function.
func purgeUserLists(context.Context, database.Executor, int64) error { return nil }
