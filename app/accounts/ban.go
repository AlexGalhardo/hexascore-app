package accounts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/orm"

	"github.com/AlexGalhardo/hexascore-app/app/mail"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Ban and role errors; handlers map them to 409 or 422 problems.
var (
	ErrAlreadyBanned = errors.New("accounts: user is already banned")
	ErrNotBanned     = errors.New("accounts: user is not banned")
	ErrBanSelf       = errors.New("accounts: admins cannot ban themselves")
	ErrBanAdmin      = errors.New("accounts: revoke the admin role before banning")
	ErrReasonTooLong = errors.New("accounts: ban reason is longer than 500 characters")
	ErrExpiryPast    = errors.New("accounts: ban expiry is not in the future")
)

// maxBanReason mirrors the users_ban_reason_length CHECK (500 characters).
const maxBanReason = 500

// UserBanned is dispatched after a ban commits (SDD 03 FR-03-14); SDD 02
// recomputes the scores of the user's rated items.
type UserBanned struct {
	UserID int64
	At     time.Time
}

// UserUnbanned is dispatched after an unban commits.
type UserUnbanned struct {
	UserID int64
	At     time.Time
}

// Clock returns the current time of the accounts service.
func (a *Accounts) Clock() time.Time { return a.now() }

// Ban bans userID on behalf of actorID: it stamps the ban, ends sessions and
// tokens, writes the account event and, after commit, dispatches UserBanned.
// reason may be empty; expires nil means permanent. The row is re-read under
// lock inside the transaction.
func (a *Accounts) Ban(ctx context.Context, actorID, userID int64, reason string, expires *time.Time) error {
	return a.DB.Tx(ctx, func(tx *database.Tx) error {
		return a.BanTx(ctx, tx, actorID, userID, reason, expires)
	})
}

// BanTx is Ban inside the caller's transaction (the admin panel's AfterSave
// hook); the post-commit work runs when that transaction commits.
func (a *Accounts) BanTx(
	ctx context.Context,
	tx *database.Tx,
	actorID, userID int64,
	reason string,
	expires *time.Time,
) error {
	reason = strings.TrimSpace(reason)
	now := a.now().UTC()

	switch {
	case actorID == userID:
		return ErrBanSelf
	case len([]rune(reason)) > maxBanReason:
		return ErrReasonTooLong
	case expires != nil && !expires.After(now):
		return ErrExpiryPast
	}

	u, err := orm.QueryOf[models.User](tx).Where("id", userID).LockForUpdate().First(ctx)
	if err != nil {
		return err
	}

	switch {
	case u.IsBanned(now):
		return ErrAlreadyBanned
	case u.Role == models.RoleAdmin:
		return ErrBanAdmin
	}

	set := orm.Set{"banned_at": now, "ban_reason": nil, "ban_expires_at": nil}
	if reason != "" {
		set["ban_reason"] = reason
	}

	if expires != nil {
		set["ban_expires_at"] = expires.UTC()
	}

	if _, err := orm.QueryOf[models.User](tx).Where("id", userID).Update(ctx, set); err != nil {
		return err
	}

	if err := a.revokeAccess(ctx, tx, userID); err != nil {
		return err
	}

	data := map[string]any{"reason": reason}
	if expires != nil {
		data["expiresAt"] = expires.UTC().Format(time.RFC3339)
	}

	if err := recordEvent(ctx, tx, userID, &actorID, "banned", data, now); err != nil {
		return err
	}

	tx.AfterCommit(func(ctx context.Context) {
		a.afterBanChange(ctx, userID, true, now)
	})

	return nil
}

// Unban lifts an active ban and dispatches UserUnbanned after commit.
func (a *Accounts) Unban(ctx context.Context, actorID, userID int64) error {
	return a.DB.Tx(ctx, func(tx *database.Tx) error {
		return a.UnbanTx(ctx, tx, actorID, userID)
	})
}

// UnbanTx is Unban inside the caller's transaction.
func (a *Accounts) UnbanTx(ctx context.Context, tx *database.Tx, actorID, userID int64) error {
	now := a.now().UTC()

	u, err := orm.QueryOf[models.User](tx).Where("id", userID).LockForUpdate().First(ctx)
	if err != nil {
		return err
	}

	if !u.BannedAt.Valid {
		return ErrNotBanned
	}

	set := orm.Set{"banned_at": nil, "ban_reason": nil, "ban_expires_at": nil}
	if _, err := orm.QueryOf[models.User](tx).Where("id", userID).Update(ctx, set); err != nil {
		return err
	}

	if err := recordEvent(ctx, tx, userID, &actorID, "unbanned", map[string]any{}, now); err != nil {
		return err
	}

	tx.AfterCommit(func(ctx context.Context) {
		a.afterBanChange(ctx, userID, false, now)
	})

	return nil
}

// afterBanChange runs once the transaction committed: server-side sessions
// go (the version bump already ended the encrypted-cookie ones) and the
// event is dispatched. Failures are logged; the ban itself is durable.
func (a *Accounts) afterBanChange(ctx context.Context, userID int64, banned bool, at time.Time) {
	if banned {
		if err := a.revokeServerSessions(ctx, userID, ""); err != nil {
			a.log.ErrorContext(ctx, "ban: revoke server sessions", "user", userID, "err", err)
		}
	}

	if a.Events == nil {
		return
	}

	var event any = UserUnbanned{UserID: userID, At: at}
	if banned {
		event = UserBanned{UserID: userID, At: at}
	}

	if err := a.Events.Dispatch(ctx, event); err != nil {
		a.log.ErrorContext(ctx, "ban: dispatch event", "user", userID, "err", err)
	}
}

// SetRole grants or revokes the admin role from the CLI (actor NULL,
// data.cli true). It reports false when the role was already set.
func (a *Accounts) SetRole(ctx context.Context, email, role string) (bool, error) {
	if role != "user" && role != models.RoleAdmin {
		return false, fmt.Errorf("accounts: unknown role %q", role)
	}

	changed := false
	now := a.now().UTC()

	err := a.DB.Tx(ctx, func(tx *database.Tx) error {
		u, err := orm.QueryOf[models.User](
			tx,
		).Where("email", strings.ToLower(strings.TrimSpace(email))).
			LockForUpdate().
			First(ctx)
		if err != nil {
			return err
		}

		if u.Role == role {
			return nil
		}

		if _, err := orm.QueryOf[models.User](tx).Where("id", u.ID).Update(ctx, orm.Set{"role": role}); err != nil {
			return err
		}

		kind := "role_granted"
		if role != models.RoleAdmin {
			kind = "role_revoked"
		}

		changed = true

		// A role change is a privilege change: the CLI cannot Regenerate the
		// user's sessions, so it ends them (SDD 13 §7.5) and the next login
		// starts a fresh one.
		if err := a.bumpSessionVersion(ctx, tx, u.ID); err != nil {
			return err
		}

		del := "DELETE FROM remember_tokens WHERE user_id = " + tx.Dialect().Placeholder(1)
		if _, err := tx.ExecContext(ctx, del, u.AuthID()); err != nil {
			return err
		}

		return recordEvent(ctx, tx, u.ID, nil, kind, map[string]any{"cli": true, "role": role}, now)
	})

	return changed, err
}

// recordEvent appends to account_events (the user-facing security history).
func recordEvent(
	ctx context.Context,
	ex database.Executor,
	userID int64,
	actorID *int64,
	kind string,
	data map[string]any,
	at time.Time,
) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	var actor any
	if actorID != nil {
		actor = *actorID
	}

	p := ex.Dialect().Placeholder
	q := "INSERT INTO account_events (user_id, actor_id, kind, data, created_at) VALUES (" +
		p(1) + ", " + p(2) + ", " + p(3) + ", " + p(4) + ", " + p(5) + ")"

	_, err = ex.ExecContext(ctx, q, userID, actor, kind, string(raw), at)

	return err
}

// ResetTwoFactor clears a user's two-factor setup after the owner verified
// their identity by hand (CLI only): everything is revoked, the event is
// recorded with actor NULL and data.cli true, and the user is mailed.
func (a *Accounts) ResetTwoFactor(ctx context.Context, email, reason string) error {
	now := a.now().UTC()

	var u models.User

	err := a.DB.Tx(ctx, func(tx *database.Tx) error {
		var err error

		u, err = orm.QueryOf[models.User](
			tx,
		).Where("email", strings.ToLower(strings.TrimSpace(email))).
			LockForUpdate().
			First(ctx)
		if err != nil {
			return err
		}

		reset := orm.Set{
			"two_factor_secret":         nil,
			"two_factor_recovery_codes": nil,
			"two_factor_confirmed_at":   nil,
			"two_factor_last_step":      0,
		}
		if _, err := orm.QueryOf[models.TwoFactor](tx).Where("id", u.ID).Update(ctx, reset); err != nil {
			return err
		}

		if err := a.revokeAccess(ctx, tx, u.ID); err != nil {
			return err
		}

		return recordEvent(ctx, tx, u.ID, nil, "two_factor_reset", map[string]any{"cli": true, "reason": reason}, now)
	})
	if err != nil {
		return err
	}

	if err := a.revokeServerSessions(ctx, u.ID, ""); err != nil {
		a.log.ErrorContext(ctx, "two-factor reset: revoke server sessions", "user", u.ID, "err", err)
	}

	return a.Mailer.SendLater(ctx, mail.TwoFactorReset{
		User: u, ResetLink: a.URL + "/recover-password",
		Strings: a.mailStrings(ctx, u, "two_factor_reset", nil),
	})
}
