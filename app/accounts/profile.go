package accounts

import (
	"context"
	"errors"
	"strconv"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/session"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Profile holds the users columns the profile page edits (SDD 03 section 5).
// They are read with SQL because models.User must not carry columns whose
// zero value would break an insert (locale has a CHECK).
type Profile struct {
	Locale                string
	PasswordSet           bool
	ProfilePrivate        bool
	PersonalizationOptOut bool
}

// Profile loads the profile columns of userID.
func (a *Accounts) Profile(ctx context.Context, userID int64) (Profile, error) {
	var p Profile

	// Fail closed: only an account positively known to be social-only (a
	// linked social account and no password_set_at) may set a password
	// without the current one. Accounts created before password_set_at
	// existed, or imported without it, keep requiring their password
	// (security review, P1.C.2.2).
	err := a.DB.QueryRowContext(ctx,
		`SELECT locale, password_set_at IS NOT NULL
  OR NOT EXISTS (SELECT 1 FROM social_accounts s WHERE s.user_id = users.id),
  profile_private, personalization_opt_out FROM users WHERE id = `+
			a.DB.Dialect().
				Placeholder(1),
		userID).Scan(&p.Locale, &p.PasswordSet, &p.ProfilePrivate, &p.PersonalizationOptOut)

	return p, err
}

// PasswordSet reports whether the user chose a password (false for accounts
// created through Google, whose password is random).
func (a *Accounts) PasswordSet(ctx context.Context, userID int64) (bool, error) {
	p, err := a.Profile(ctx, userID)

	return p.PasswordSet, err
}

// UpdateDetails saves the display name and interface locale.
func (a *Accounts) UpdateDetails(ctx context.Context, userID int64, name, locale string) error {
	p := a.DB.Dialect().Placeholder
	_, err := a.DB.ExecContext(ctx,
		"UPDATE users SET name = "+p(1)+", locale = "+p(2)+", updated_at = "+p(3)+" WHERE id = "+p(4),
		name, locale, a.now().UTC(), userID)

	return err
}

// UpdatePrivacy saves the privacy switches. personalizationOptOut is stored
// now and read by the recommendation work (SDD 08).
func (a *Accounts) UpdatePrivacy(ctx context.Context, userID int64, private, optOut bool) error {
	p := a.DB.Dialect().Placeholder

	return a.DB.Tx(ctx, func(tx *database.Tx) error {
		if _, err := tx.ExecContext(ctx,
			"UPDATE users SET profile_private = "+p(1)+", personalization_opt_out = "+p(2)+", updated_at = "+p(3)+
				" WHERE id = "+p(4), private, optOut, a.now().UTC(), userID); err != nil {
			return err
		}

		// Opting out removes the stored "For you" rows (SDD 08 FR-08-10, D-29); the table exists on Postgres only.
		if optOut && a.DB.Dialect().Name() == "postgres" {
			_, err := tx.ExecContext(ctx, "DELETE FROM user_recommendations WHERE user_id = "+p(1), userID)

			return err
		}

		return nil
	})
}

// ChangePassword sets a new password for u. The current password is
// required only when the user already chose one. Every other session,
// remember-me token and API token ends; keep is the current session id,
// which the caller must refresh with Web.RefreshSession.
func (a *Accounts) ChangePassword(ctx context.Context, u models.User, current, next, keep string) error {
	set, err := a.PasswordSet(ctx, u.ID)
	if err != nil {
		return err
	}

	if set {
		if ok, err := a.Hasher.Verify(u.Password, current); err != nil || !ok {
			return errs.NewValidation().Add("currentPassword", a.T(ctx, "auth.two_factor.wrong_password"))
		}
	}

	err = a.DB.Tx(ctx, func(tx *database.Tx) error {
		if err := a.revokeAccess(ctx, tx, u.ID); err != nil {
			return err
		}

		if err := a.updatePassword(ctx, tx, u.ID, next); err != nil {
			return err
		}

		return recordEvent(ctx, tx, u.ID, nil, "password_changed", map[string]any{"via": "profile"}, a.now().UTC())
	})
	if err != nil {
		return err
	}

	if err := a.revokeServerSessions(ctx, u.ID, keep); err != nil {
		return err
	}

	a.sendPasswordChanged(ctx, u.ID)

	return nil
}

// SessionInfo is one active browser session; Handle is safe to show.
type SessionInfo = session.Info

// Sessions lists the live browser sessions of userID, the current one marked.
// Cookie-only session drivers cannot list: the result is empty then.
func (a *Accounts) Sessions(ctx context.Context, userID int64, currentID string) ([]SessionInfo, error) {
	m, err := a.sessions()
	if err != nil {
		return nil, err
	}

	list, err := m.UserSessions(ctx, a.Web.SessionUserID(strconv.FormatInt(userID, 10)), currentID)
	if errors.Is(err, session.ErrListUnsupported) {
		return []SessionInfo{}, nil
	}

	return list, err
}

// RevokeSession ends one session of userID by handle; a handle of another
// user's session answers not found.
func (a *Accounts) RevokeSession(ctx context.Context, userID int64, handle string) error {
	m, err := a.sessions()
	if err != nil {
		return err
	}

	err = m.RevokeSession(ctx, a.Web.SessionUserID(strconv.FormatInt(userID, 10)), handle)
	if errors.Is(err, session.ErrSessionNotFound) {
		return errs.ErrNotFound
	}

	if err != nil {
		return err
	}

	// Remember-me tokens are not tied to a session, so a revoked device could
	// sign itself back in with one; drop them all. Live sessions survive and
	// re-issue a token on the next "remember me" sign-in (security review, P1.C.2.2).
	_, err = a.DB.ExecContext(ctx, "DELETE FROM remember_tokens WHERE user_id = "+a.DB.Dialect().Placeholder(1),
		strconv.FormatInt(userID, 10))

	return err
}
