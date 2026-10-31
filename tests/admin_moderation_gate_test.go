//go:build integration

package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/AlexGalhardo/vukem/authz"
	"github.com/AlexGalhardo/vukem/drivers/postgres/pgtest"
	"github.com/AlexGalhardo/vukem/vukemtest"

	"github.com/AlexGalhardo/hexascore-app/app/admin"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/bootstrap"
)

func moderationApp(t *testing.T, trusted bool) *vukemtest.App {
	t.Helper()

	env := map[string]string{"DATABASE_URL": pgtest.Database(t, "")}
	if trusted {
		env["FLAG_COMMUNITY_TRUSTED_MODERATORS"] = "true"
	}

	return vukemtest.New(t, bootstrap.New, vukemtest.Options{Env: env})
}

func TestModerationGate(t *testing.T) {
	t.Parallel()

	off := moderationApp(t, false)
	on := moderationApp(t, true)

	for name, tc := range map[string]struct {
		app  *vukemtest.App
		role string
		want int
	}{
		"admin":              {off, "admin", 200},
		"user":               {off, "user", 403},
		"moderator flag off": {off, "moderator", 403},
		"moderator flag on":  {on, "moderator", 200},
		"user with flag on":  {on, "user", 403},
		"admin with flag on": {on, "admin", 200},
	} {
		c := browserFor(t, tc.app, makeUser(t, tc.app, name+"@example.com", tc.role, true))
		c.Get("/admin/moderation").AssertStatus(tc.want)
	}

	off.Client(t).Get("/admin/moderation").AssertRedirect("/login")

	body := browserFor(t, off, makeUser(t, off, "shell@example.com", "admin", true)).
		Get("/admin/moderation").AssertOK().Body()
	if want := `<nav class="sidebar"`; !strings.Contains(body, want) ||
		!strings.Contains(body, `href="/admin/moderation"`) {
		t.Fatalf("moderation page is not inside the admin layout with a sidebar link: %s", body)
	}
}

func TestSecurity_Community_A01_PolicyFailsClosed(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	verified := models.User{Role: models.RoleAdmin}
	verified.EmailVerifiedAt.V, verified.EmailVerifiedAt.Valid = now, true

	with := func(mut func(u *models.User)) *models.User {
		u := verified
		mut(&u)

		return &u
	}

	banned := with(func(u *models.User) { u.BannedAt.V, u.BannedAt.Valid = now.Add(-time.Hour), true })

	for _, flag := range []bool{false, true} {
		var user *models.User
		gate := authz.New(func(context.Context) (*models.User, error) { return user, nil })
		admin.DefineModeration(gate, func() time.Time { return now }, func(context.Context) bool { return flag })

		cases := map[string]*models.User{
			"nil user":         nil,
			"banned admin":     banned,
			"unverified admin": with(func(u *models.User) { u.EmailVerifiedAt.Valid = false }),
			"empty role":       with(func(u *models.User) { u.Role = "" }),
			"unknown role":     with(func(u *models.User) { u.Role = "owner" }),
			"unverified mod": with(
				func(u *models.User) { u.Role, u.EmailVerifiedAt.Valid = models.RoleModerator, false },
			),
			"banned moderator": with(
				func(u *models.User) { u.Role, u.BannedAt = models.RoleModerator, banned.BannedAt },
			),
		}
		for name, u := range cases {
			user = u
			for _, ab := range []string{
				admin.AbilityModerate, admin.AbilityViewCatalogAudit, admin.AbilityRevertCatalog, admin.AbilityLockCatalog,
			} {
				if gate.Allows(t.Context(), u, ab, nil) {
					t.Fatalf("flag=%v: %s allowed %s", flag, name, ab)
				}
			}
		}

		// Revert, lock and audit stay admin-only whatever the flag says.
		mod := with(func(u *models.User) { u.Role = models.RoleModerator })
		user = mod
		for _, ab := range []string{admin.AbilityViewCatalogAudit, admin.AbilityRevertCatalog, admin.AbilityLockCatalog} {
			if gate.Allows(t.Context(), mod, ab, nil) {
				t.Fatalf("flag=%v: moderator allowed %s", flag, ab)
			}
		}
		if got := gate.Allows(t.Context(), mod, admin.AbilityModerate, nil); got != flag {
			t.Fatalf("moderator moderate = %v with flag %v", got, flag)
		}
	}
}
