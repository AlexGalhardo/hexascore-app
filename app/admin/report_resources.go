package admin

import (
	"context"
	"errors"
	"net/http"
	"time"

	vadmin "github.com/AlexGalhardo/vukem/admin"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"
	"github.com/AlexGalhardo/vukem/z"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
)

// ReportsResource lists member reports; resolving one goes through POST /admin/reports/{id}/resolve because
// panel row actions are not released (VF-008).
func ReportsResource() *vadmin.Resource[models.Report] {
	return vadmin.NewResource[models.Report](noInputSchema).
		Slug("reports").
		Search("target_id").
		Filter("status", "open", "resolved", "dismissed").
		List("id", "status", "target_type", "target_id", "reason", "outcome", "created_at")
}

type contactEdit struct {
	Handled bool `json:"handled"`
}

var contactEditSchema = z.Struct[contactEdit](z.Shape{"handled": z.Bool().Default(false)})

// ContactMessagesResource lists contact messages; the only writable field is handled.
func ContactMessagesResource() *vadmin.Resource[models.ContactMessage] {
	return vadmin.NewResource[models.ContactMessage](contactEditSchema).
		Slug("contact-messages").
		Search("message", "name").
		Filter("subject", "question", "feedback", "bug").
		List("id", "subject", "name", "email", "handled", "created_at").
		ReadOnly("name", "email", "subject", "message", "user_id")
}

// contactPolicy lets admins list, view and edit (toggle handled); creating and deleting are denied (retention
// runs as a job, SDD 14 NFR-14-6).
type contactPolicy struct{ now func() time.Time }

func (p contactPolicy) ViewAny(_ context.Context, u *models.User) bool { return IsAdmin(u, p.now()) }

func (p contactPolicy) View(_ context.Context, u *models.User, _ *models.ContactMessage) bool {
	return IsAdmin(u, p.now())
}

func (p contactPolicy) Update(_ context.Context, u *models.User, _ *models.ContactMessage) bool {
	return IsAdmin(u, p.now())
}

func (contactPolicy) Create(context.Context, *models.User) bool { return false }

func (contactPolicy) Delete(context.Context, *models.User, *models.ContactMessage) bool { return false }

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

	return c.Redirect("/admin/reports/" + c.Param("id"))
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
