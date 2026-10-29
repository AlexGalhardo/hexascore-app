package admin

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/web"
	"github.com/AlexGalhardo/vukem/z"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// banForm is the body of POST /admin/users/{id}/ban.
type banForm struct {
	Reason    string    `json:"reason"     form:"reason"`
	ExpiresAt time.Time `json:"expires_at" form:"expires_at"`
}

var banFormSchema = z.Struct[banForm](z.Shape{
	"reason":     z.String().Trim().Max(500).Optional(),
	"expires_at": z.Time().Optional(),
})

// BanHandlers serve the ban and unban actions shared with the moderation
// pages (SDD 03 section 6.3). They sit behind the panel gate and CSRF.
type BanHandlers struct {
	Accounts *accounts.Accounts
	// Authorize is the panel gate check (404 for non-admins).
	Authorize func(c *web.Ctx) error
}

func (h BanHandlers) actorAndTarget(c *web.Ctx) (actor, target int64, err error) {
	if err := h.Authorize(c); err != nil {
		return 0, 0, err
	}

	u, err := auth.UserFrom[models.User](c.Context(), h.Accounts.Web.Name())
	if err != nil {
		return 0, 0, err
	}

	target, err = strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return 0, 0, errs.ErrNotFound
	}

	return u.ID, target, nil
}

// Ban bans the user with an optional reason and expiry.
func (h BanHandlers) Ban(c *web.Ctx) error {
	actor, target, err := h.actorAndTarget(c)
	if err != nil {
		return err
	}

	in, err := web.Parse(c, banFormSchema)
	if err != nil {
		return err
	}

	var expires *time.Time
	if !in.ExpiresAt.IsZero() {
		expires = &in.ExpiresAt
	}

	return h.done(c, target, h.Accounts.Ban(c.Context(), actor, target, in.Reason, expires))
}

// Unban lifts the ban.
func (h BanHandlers) Unban(c *web.Ctx) error {
	actor, target, err := h.actorAndTarget(c)
	if err != nil {
		return err
	}

	return h.done(c, target, h.Accounts.Unban(c.Context(), actor, target))
}

func (BanHandlers) done(c *web.Ctx, target int64, err error) error {
	switch {
	case err == nil:
		return c.Redirect("/admin/users/" + strconv.FormatInt(target, 10))
	case errors.Is(err, orm.ErrNotFound):
		return errs.ErrNotFound
	case errors.Is(err, accounts.ErrAlreadyBanned), errors.Is(err, accounts.ErrNotBanned):
		return errs.New(http.StatusConflict, err.Error())
	case errors.Is(err, accounts.ErrBanSelf), errors.Is(err, accounts.ErrBanAdmin):
		return errs.New(http.StatusForbidden, err.Error())
	case errors.Is(err, accounts.ErrReasonTooLong), errors.Is(err, accounts.ErrExpiryPast):
		return errs.New(http.StatusUnprocessableEntity, err.Error())
	}

	return err
}
