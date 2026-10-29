package account

import (
	"regexp"
	"strings"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
)

// Nickname error codes, returned as the `nickname` field error; the page maps
// them to catalogue keys.
const (
	errNicknameTaken   = "nickname-taken"
	errNicknameInvalid = "invalid-nickname"
)

var nicknameRe = regexp.MustCompile(`^[a-z]{4,16}$`)

func (h *Handlers) hasNickname(c *web.Ctx, id int64) (bool, error) {
	var has bool

	err := h.a.DB.QueryRowContext(c.Context(), `SELECT username IS NOT NULL FROM users WHERE id = $1`, id).Scan(&has)

	return has, err
}

// ShowOnboarding asks a signed-in, verified user without a nickname for one;
// users who have it go to the dashboard.
func (h *Handlers) ShowOnboarding(c *web.Ctx) error {
	u, err := auth.User[models.User](c, "web")
	if err != nil {
		return err
	}

	if has, err := h.hasNickname(c, u.ID); err != nil || has {
		if err != nil {
			return err
		}

		return c.Redirect("/dashboard")
	}

	return inertia.Render(c, "Auth/Onboarding", nil)
}

// Onboarding stores the nickname once (it never changes afterwards, SDD 03).
// A nickname someone else holds answers 422 `nickname-taken`, also when the
// unique index decides a race.
func (h *Handlers) Onboarding(c *web.Ctx) error {
	u, err := auth.User[models.User](c, "web")
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.OnboardingSchema)
	if err != nil {
		return err
	}

	nick := strings.ToLower(in.Nickname)
	if !nicknameRe.MatchString(nick) {
		return errs.NewValidation().Add("nickname", errNicknameInvalid)
	}

	_, err = h.a.DB.ExecContext(c.Context(),
		`UPDATE users SET username = $1, display_username = $1, updated_at = now() WHERE id = $2 AND username IS NULL`,
		nick, u.ID)
	if err != nil {
		if strings.Contains(err.Error(), "users_username_uq") {
			return errs.NewValidation().Add("nickname", errNicknameTaken)
		}

		return err
	}

	// Zero rows: the user already had a nickname, which never changes.
	return c.Redirect("/dashboard")
}
