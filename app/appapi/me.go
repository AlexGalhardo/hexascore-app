package appapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"
	"github.com/AlexGalhardo/vukem/z"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/controllers/account"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Me serves GET/PATCH /app/v1/me and POST/DELETE /app/v1/me/deletion.
// Export stays on routes.Exports. The caller puts catalog:read on the read
// and profile:write plus idempotency on the writes (SDD 03 §6.2).
type Me struct {
	a     *accounts.Accounts
	guard string
}

// NewMe builds the profile handlers for one guard.
func NewMe(a *accounts.Accounts, guard string) *Me {
	return &Me{a: a, guard: guard}
}

type meUpdate struct {
	Name                  string `json:"name"`
	Locale                string `json:"locale"`
	ProfilePrivate        bool   `json:"profilePrivate"`
	PersonalizationOptOut bool   `json:"personalizationOptOut"`
	Username              string `json:"username"`
}

var meUpdateSchema = z.Struct[meUpdate](z.Shape{
	"name":                  z.String().Trim().Min(1).Max(100),
	"locale":                z.Enum("en", "pt-BR", "es"),
	"profilePrivate":        z.Bool(),
	"personalizationOptOut": z.Bool(),
	"username":              z.String().Trim().Optional(),
})

// Show is GET /app/v1/me.
func (h *Me) Show(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	body, err := h.load(c, u.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, body)
}

// Update is PATCH /app/v1/me. Username is stored only while unset. A nickname
// someone else holds is 422 nickname-taken and leaves the other fields alone.
func (h *Me) Update(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	in, err := web.Parse(c, meUpdateSchema)
	if err != nil {
		return err
	}

	if in.Username != "" {
		if err := h.a.SetNickname(c.Context(), u.ID, in.Username, "username"); err != nil {
			return err
		}
	}

	if err := h.a.UpdateDetails(c.Context(), u.ID, in.Name, in.Locale); err != nil {
		return err
	}

	if err := h.a.UpdatePrivacy(c.Context(), u.ID, in.ProfilePrivate, in.PersonalizationOptOut); err != nil {
		return err
	}

	body, err := h.load(c, u.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, body)
}

// Schedule is POST /app/v1/me/deletion. The service revokes sessions and tokens.
func (h *Me) Schedule(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	if _, err := web.Parse(c, account.DeletionSchema); err != nil {
		return err
	}

	at, err := h.a.ScheduleDeletion(c.Context(), u.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]any{"deletesAt": at.UTC().Format(time.RFC3339)})
}

// Cancel is DELETE /app/v1/me/deletion.
func (h *Me) Cancel(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	if err := h.a.CancelDeletion(c.Context(), u.ID); err != nil {
		if errors.Is(err, accounts.ErrNotScheduled) {
			return errs.New(http.StatusConflict, h.a.T(c.Context(), "auth.deletion.not_scheduled"))
		}

		return err
	}

	return c.NoContent()
}

func (h *Me) load(c *web.Ctx, userID int64) (map[string]any, error) {
	p := h.a.DB.Dialect().Placeholder

	var name, email, locale string
	var private, optOut bool
	var username sql.NullString
	var deletion sql.NullTime
	err := h.a.DB.QueryRowContext(
		c.Context(),
		"SELECT name, email, locale, profile_private, personalization_opt_out, username, deletion_scheduled_at "+
			"FROM users WHERE id = "+p(1),
		userID,
	).Scan(&name, &email, &locale, &private, &optOut, &username, &deletion)
	if err != nil {
		return nil, err
	}

	// id is a string so JavaScript never rounds it; the app's offline queue binds writes to it (SDD 10 §7.3).
	out := map[string]any{
		"id": strconv.FormatInt(userID, 10), "name": name, "email": email, "locale": locale,
		"profilePrivate": private, "personalizationOptOut": optOut, "deletionScheduledAt": nil,
	}
	if deletion.Valid {
		out["deletionScheduledAt"] = deletion.Time.UTC().Format(time.RFC3339)
	}
	if username.Valid {
		out["username"] = username.String
	}

	return out, nil
}
