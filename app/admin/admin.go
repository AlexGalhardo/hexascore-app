// Package admin defines the admin panel's gate, resources and policies
// (vukem install:admin, adapted: SDD 03 section 6.3).
package admin

import (
	"context"
	"time"

	"github.com/AlexGalhardo/vukem/authz"

	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// IsAdmin is the admin gate: role admin, not banned, email verified. It
// replaces the scaffold's ADMIN_EMAILS list; roles change only through the
// admin:grant and admin:revoke commands.
func IsAdmin(u *models.User, now time.Time) bool {
	return u != nil && u.Role == models.RoleAdmin && !u.IsBanned(now) && u.EmailVerified()
}

// Gate returns the gate result for the panel abilities: non-admins get 404,
// so the panel's existence is not revealed (legacy admin.spec.ts).
func Gate(now func() time.Time) func(ctx context.Context, u *models.User) authz.Result {
	return func(_ context.Context, u *models.User) authz.Result {
		if IsAdmin(u, now()) {
			return authz.Allow()
		}

		return authz.DenyAsNotFound()
	}
}
