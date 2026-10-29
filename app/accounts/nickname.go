package accounts

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// HasNickname reports whether the user has chosen a nickname (SDD 03 FR-03-9).
func (a *Accounts) HasNickname(ctx context.Context, userID int64) (bool, error) {
	var has bool

	err := a.DB.QueryRowContext(ctx, `SELECT username IS NOT NULL FROM users WHERE id = $1`, userID).Scan(&has)

	return has, err
}

// RedirectAfterLogin ends a successful login (password, two-factor or
// provider). A verified user without a nickname goes to /onboarding; anybody
// else goes back to the page that asked for the login, such as a
// verification link, or to the dashboard.
func (a *Accounts) RedirectAfterLogin(c *web.Ctx, u models.User) error {
	target, _ := c.Session().Pull(auth.IntendedKey).(string)

	if u.EmailVerified() {
		has, err := a.HasNickname(c.Context(), u.ID)
		if err != nil {
			return err
		}

		if !has {
			return c.Redirect("/onboarding")
		}
	}

	return c.Redirect(web.SafeRedirect(c.Request(), target, "/dashboard"))
}

// EmailForIdentifier maps a login identifier to an email: anything with "@"
// is an email, anything else a nickname. An unknown nickname maps to a value
// that is not an email, so the guard still spends a password hash and
// answers like a wrong password (SDD 03 FR-03-3).
func (a *Accounts) EmailForIdentifier(ctx context.Context, identifier string) (string, error) {
	id := strings.ToLower(strings.TrimSpace(identifier))
	if strings.Contains(id, "@") {
		return id, nil
	}

	var email string

	err := a.DB.QueryRowContext(ctx, `SELECT email FROM users WHERE username = $1`, id).Scan(&email)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "unknown-nickname", nil
	case err != nil:
		return "", err
	}

	return email, nil
}
