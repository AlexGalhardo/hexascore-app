package admin

import (
	"context"
	"time"

	vadmin "github.com/AlexGalhardo/vukem/admin"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/z"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// userEdit lists the only user fields the panel can write. Role, email,
// password and two-factor columns are not schema keys, so a crafted POST
// cannot reach them (mass-assignment protection).
type userEdit struct {
	BanReason    string    `json:"banReason"`
	BanExpiresAt time.Time `json:"banExpiresAt"`
}

var userEditSchema = z.Struct[userEdit](z.Shape{
	"banReason":    z.String().Trim().Max(500).Optional(),
	"banExpiresAt": z.Time().Optional(),
})

// UsersResource is the users panel resource. Ban and unban go through
// accounts.Ban and accounts.Unban (routes POST /admin/users/{id}/ban and
// /unban) because BeforeSave has no transaction handle or after-commit hook
// (VF-NEW-P1.C.3.1): the form only edits the reason and expiry of an active ban.
func UsersResource() *vadmin.Resource[models.User] {
	return vadmin.NewResource[models.User](userEditSchema).
		Slug("users").
		Search("email", "name").
		Filter("role", "user", "admin").
		List("id", "email", "name", "role", "banned_at", "created_at").
		ReadOnly("role", "email").
		Sensitive("password").
		BeforeSave(func(_ context.Context, u *models.User) error {
			if u.BanReason != nil && *u.BanReason == "" {
				u.BanReason = nil
			}

			if !u.BannedAt.Valid && (u.BanReason != nil || u.BanExpiresAt.Valid) {
				return errs.NewValidation().
					Add("banReason", "Ban the user first; reason and expiry only apply to a ban.")
			}

			if u.BanExpiresAt.Valid && !u.BanExpiresAt.V.After(time.Now()) {
				return errs.NewValidation().Add("banExpiresAt", "The expiry must be in the future.")
			}

			return nil
		})
}

// UserPolicy lets admins list, view and edit users; creating and deleting
// are denied (deletion goes through SDD 03 section 7.5).
type UserPolicy struct{ now func() time.Time }

// ViewAny allows the list.
func (p UserPolicy) ViewAny(_ context.Context, u *models.User) bool { return IsAdmin(u, p.now()) }

// View allows the detail page.
func (p UserPolicy) View(_ context.Context, u *models.User, _ *models.User) bool {
	return IsAdmin(u, p.now())
}

// Update allows editing the ban reason and expiry.
func (p UserPolicy) Update(_ context.Context, u *models.User, _ *models.User) bool {
	return IsAdmin(u, p.now())
}

// Create is always denied.
func (UserPolicy) Create(context.Context, *models.User) bool { return false }

// Delete is always denied.
func (UserPolicy) Delete(context.Context, *models.User, *models.User) bool { return false }
