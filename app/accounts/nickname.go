package accounts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Nickname error codes. The page and the app map them to catalogue keys.
const (
	NicknameTaken   = "nickname-taken"
	NicknameInvalid = "invalid-nickname"
)

var nicknameRe = regexp.MustCompile(`^[a-z]{4,16}$`)

// SetNickname stores a nickname once (it never changes afterwards, SDD 03).
// field is the validation key (`nickname` on the web form, `username` on
// /app/v1/me). A nickname someone else holds answers 422 nickname-taken,
// also when the unique index decides a race. A user who already has one is
// left unchanged.
func (a *Accounts) SetNickname(ctx context.Context, userID int64, raw, field string) error {
	nick := strings.ToLower(strings.TrimSpace(raw))
	if !nicknameRe.MatchString(nick) {
		return errs.NewValidation().Add(field, NicknameInvalid)
	}

	p := a.DB.Dialect().Placeholder
	_, err := a.DB.ExecContext(ctx,
		"UPDATE users SET username = "+p(1)+", display_username = "+p(2)+", updated_at = "+p(3)+
			" WHERE id = "+p(4)+" AND username IS NULL",
		nick, nick, a.now().UTC(), userID)
	if err != nil && strings.Contains(err.Error(), "users_username_uq") {
		return errs.NewValidation().Add(field, NicknameTaken)
	}

	return err
}

// checkSignupNickname normalises the optional sign-up nickname and rejects it
// when it breaks the rule or is taken. An empty one is allowed: the account
// then picks it at /onboarding.
func (a *Accounts) checkSignupNickname(ctx context.Context, raw string) (string, error) {
	nick := strings.ToLower(strings.TrimSpace(raw))
	if nick == "" {
		return "", nil
	}

	if !nicknameRe.MatchString(nick) {
		return "", errs.NewValidation().Add("nickname", NicknameInvalid)
	}

	var taken bool
	if err := a.DB.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)`, nick).
		Scan(&taken); err != nil {
		return "", fmt.Errorf("check nickname: %w", err)
	}

	if taken {
		return "", errs.NewValidation().Add("nickname", NicknameTaken)
	}

	return nick, nil
}

func isNicknameTaken(err error) bool {
	v, ok := errors.AsType[*errs.ValidationError](err)

	return ok && v.First("nickname") == NicknameTaken
}

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
