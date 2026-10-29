package admin

import (
	"context"
	"time"

	vadmin "github.com/AlexGalhardo/vukem/admin"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/orm"
	"github.com/AlexGalhardo/vukem/z"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// userEdit lists the only user fields the panel can write. Role, email,
// password and two-factor columns are not schema keys, so a crafted POST
// cannot reach them (mass-assignment protection).
type userEdit struct {
	Banned       bool      `json:"banned"`
	BanReason    string    `json:"banReason"`
	BanExpiresAt time.Time `json:"banExpiresAt"`
}

var userEditSchema = z.Struct[userEdit](z.Shape{
	"banned":       z.Bool().Optional(),
	"banReason":    z.String().Trim().Max(500).Optional(),
	"banExpiresAt": z.Time().Optional(),
})

// UsersResource is the users panel resource. The virtual "banned" checkbox
// bans or unbans through accounts.BanTx and accounts.UnbanTx in the panel
// transaction (AfterSave), so sessions, the account event and the
// user.banned dispatch share the save's commit (VF-046). The routes
// POST /admin/users/{id}/ban and /unban stay for the moderation pages.
func UsersResource(acc *accounts.Accounts) *vadmin.Resource[models.User] {
	return vadmin.NewResource[models.User](userEditSchema).
		Slug("users").
		Search("email", "name").
		Filter("role", "user", "admin").
		List("id", "email", "name", "role", "banned_at", "created_at").
		ReadOnly("role", "email").
		Sensitive("password").
		Virtual("banned", func(u *models.User) any { return u.BannedAt.Valid }).
		BeforeSave(func(_ context.Context, u *models.User) error {
			if u.BanReason != nil && *u.BanReason == "" {
				u.BanReason = nil
			}

			if u.BanExpiresAt.Valid && !u.BanExpiresAt.V.After(time.Now()) {
				return errs.NewValidation().Add("banExpiresAt", "The expiry must be in the future.")
			}

			return nil
		}).
		AfterSave(func(ctx context.Context, saved vadmin.Saved[models.User]) error {
			return applyBanChange(ctx, acc, saved)
		})
}

// applyBanChange turns the submitted "banned" checkbox into a ban or unban.
// An absent key (a programmatic edit of the reason or expiry) changes nothing.
// A failure rolls the whole save back, including the reason and expiry the
// panel already wrote.
func applyBanChange(ctx context.Context, acc *accounts.Accounts, saved vadmin.Saved[models.User]) error {
	rec := saved.Record
	want, submitted := saved.Virtual["banned"].(bool)
	wasBanned := saved.Before != nil && saved.Before.BannedAt.Valid

	switch {
	case submitted && want && !wasBanned:
		actor, err := auth.UserFrom[models.User](ctx, acc.Web.Name())
		if err != nil {
			return err
		}

		var expires *time.Time
		if rec.BanExpiresAt.Valid {
			expires = &rec.BanExpiresAt.V
		}

		reason := ""
		if rec.BanReason != nil {
			reason = *rec.BanReason
		}

		if err := acc.BanTx(ctx, saved.Tx, actor.ID, rec.ID, reason, expires); err != nil {
			return mapBanErr(err)
		}
	case submitted && !want && wasBanned:
		actor, err := auth.UserFrom[models.User](ctx, acc.Web.Name())
		if err != nil {
			return err
		}

		if err := acc.UnbanTx(ctx, saved.Tx, actor.ID, rec.ID); err != nil {
			return mapBanErr(err)
		}
	default:
		if !rec.BannedAt.Valid && (rec.BanReason != nil || rec.BanExpiresAt.Valid) {
			return errs.NewValidation().
				Add("banReason", "Ban the user first; reason and expiry only apply to a ban.")
		}

		return nil
	}

	fresh, err := orm.QueryOf[models.User](saved.Tx).Where("id", rec.ID).First(ctx)
	if err != nil {
		return err
	}

	*rec = fresh

	return nil
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
