package admin

import (
	"errors"
	"net/http"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
)

// ReportHandlers serve the report resolve route behind the panel gate and CSRF.
type ReportHandlers struct {
	Accounts *accounts.Accounts
	Reports  *library.Reports
	// Authorize is the panel gate check (404 for non-admins).
	Authorize func(c *web.Ctx) error
}

// Resolve applies the outcome of POST /admin/reports/{id}/resolve (SDD 14 §6.4).
func (h ReportHandlers) Resolve(c *web.Ctx) error {
	if err := h.Authorize(c); err != nil {
		return err
	}

	u, err := auth.UserFrom[models.User](c.Context(), h.Accounts.Web.Name())
	if err != nil {
		return err
	}

	in, err := web.Parse(c, requests.ReportResolveSchema)
	if err != nil {
		return err
	}

	err = h.Reports.Resolve(c.Context(), u.ID, c.Param("id"), library.ResolveInput{
		Outcome: in.Outcome, AdminNote: in.AdminNote,
	})
	if err != nil {
		return mapBanErr(err)
	}

	return c.Redirect("/admin/r/reports/" + c.Param("id"))
}

// mapBanErr gives the ban rules of accounts.Ban their HTTP statuses, as BanHandlers does.
func mapBanErr(err error) error {
	switch {
	case errors.Is(err, accounts.ErrBanSelf), errors.Is(err, accounts.ErrBanAdmin):
		return errs.New(http.StatusForbidden, err.Error())
	case errors.Is(err, accounts.ErrAlreadyBanned), errors.Is(err, accounts.ErrNotBanned):
		return errs.New(http.StatusConflict, err.Error())
	case errors.Is(err, accounts.ErrReasonTooLong), errors.Is(err, accounts.ErrExpiryPast):
		return errs.New(http.StatusUnprocessableEntity, err.Error())
	}

	return err
}
