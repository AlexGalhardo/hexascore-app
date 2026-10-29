package accounts

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/ratelimit"
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

// LoginThrottleKeys are auth.ThrottleKeys keyed by the resolved account: an
// email and the nickname of the same user share one budget, so alternating
// identifiers cannot double the guesses. Unknown nicknames keep their own
// lowercased key; non-POST requests are not counted.
func (a *Accounts) LoginThrottleKeys(field string) (perIP, perAccount ratelimit.KeyFunc) {
	baseIP, baseAccount := auth.ThrottleKeys(field)
	account := func(r *http.Request) (string, bool) {
		if r.Method != http.MethodPost {
			return "", false
		}
		id := auth.RequestEmail(r, field)
		if id == "" {
			return "", false
		}
		email, err := a.EmailForIdentifier(r.Context(), id)
		if err != nil || email == "unknown-nickname" {
			email = "nick:" + id
		}
		sum := sha256.Sum256([]byte(email))
		return hex.EncodeToString(sum[:16]), true
	}
	perIP = func(r *http.Request) string {
		if key, ok := account(r); ok {
			return "login:" + key + ":" + ratelimit.ByIP(r)
		}
		return baseIP(r)
	}
	perAccount = func(r *http.Request) string {
		if key, ok := account(r); ok {
			return "login.account:" + key
		}
		return baseAccount(r)
	}
	return perIP, perAccount
}
