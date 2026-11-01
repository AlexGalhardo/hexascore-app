package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/flash"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Admin session limits (SDD 13 §7.5).
const (
	AdminIdleTimeout = 30 * time.Minute
	AdminMaxLifetime = 12 * time.Hour
)

// Session keys of the admin session clock, bound to the user id so a key
// left by another account in the same browser session never counts.
const (
	adminUIDKey   = "admin.uid"
	adminSinceKey = "admin.since"
	adminSeenKey  = "admin.seen"
)

// AdminGuardConfig wires AdminGuard to the accounts service.
type AdminGuardConfig struct {
	// User returns the signed-in user, or an error for a guest.
	User func(c *web.Ctx) (models.User, error)
	// TwoFactorEnabled reports whether the user confirmed TOTP.
	TwoFactorEnabled func(ctx context.Context, userID int64) (bool, error)
	// Logout ends the session (guard logout: session and remember-me).
	Logout func(c *web.Ctx) error
	// Allowlist limits /admin to these networks; empty means off.
	Allowlist []netip.Prefix
	Now       func() time.Time
	// SetupPath is the 2FA setup page; LoginPath the login page.
	SetupPath, LoginPath string
}

// ParseAllowlist parses ADMIN_IP_ALLOWLIST: comma-separated CIDRs or single
// addresses. Empty means off; a bad entry is a boot error (fail closed).
func ParseAllowlist(s string) ([]netip.Prefix, error) {
	var out []netip.Prefix

	for f := range strings.SplitSeq(s, ",") {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}

		p, err := netip.ParsePrefix(f)
		if err != nil {
			a, aerr := netip.ParseAddr(f)
			if aerr != nil {
				return nil, fmt.Errorf("ADMIN_IP_ALLOWLIST: %q is not a CIDR or address", f)
			}

			p = netip.PrefixFrom(a, a.BitLen())
		}

		out = append(out, p.Masked())
	}

	return out, nil
}

// AdminGuard is the router middleware of SDD 13 FR-13-5 (workaround(vukem):
// VF-NEW-ADMIN-GUARD-SCOPE, role-scoped 2FA, session lifetime and IP allowlist):
//   - /admin from an address outside a non-empty allowlist: 404, like a
//     non-admin, so the panel's existence is not revealed;
//   - any request of a signed-in admin or moderator: idle 30 min and absolute 12 h, then
//     logout and a redirect to the login page;
//   - /admin by an admin or moderator without confirmed TOTP: redirect to the 2FA setup.
//
// Role and ban checks stay with the panel gate (admin.Gate).
func AdminGuard(cfg AdminGuardConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := web.NewCtx(w, r, web.ServicesFrom(r.Context()))
			adminPath := r.URL.Path == "/admin" || strings.HasPrefix(r.URL.Path, "/admin/")

			if adminPath && !allowed(cfg.Allowlist, c.IP()) {
				web.SecurityEvent(r, "admin.ip_denied", "ip", c.IP())
				http.NotFound(w, r)

				return
			}

			u, err := cfg.User(c)
			if err != nil || u.Role != models.RoleAdmin && u.Role != models.RoleModerator {
				next.ServeHTTP(w, r)
				return
			}

			if expired, why := touchAdminSession(c.Session(), u.ID, cfg.Now()); expired {
				web.SecurityEvent(r, "admin.session_expired", "user_id", u.AuthID(), "reason", why)

				if err := cfg.Logout(c); err != nil {
					c.HandleError(err)
					return
				}

				http.Redirect(w, r, cfg.LoginPath, http.StatusSeeOther)

				return
			}

			if adminPath {
				on, err := cfg.TwoFactorEnabled(r.Context(), u.ID)
				if err != nil {
					c.HandleError(err)
					return
				}

				if !on {
					flash.Status(c, "flash.adminTwoFactorRequired")
					http.Redirect(w, r, cfg.SetupPath, http.StatusSeeOther)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func allowed(list []netip.Prefix, ip string) bool {
	if len(list) == 0 {
		return true
	}

	a, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}

	a = a.Unmap()
	for _, p := range list {
		if p.Contains(a) {
			return true
		}
	}

	return false
}

// touchAdminSession checks and advances the admin session clock. Without a
// session (misconfiguration) it reports expired: fail closed.
func touchAdminSession(s web.Session, userID int64, now time.Time) (bool, string) {
	if s == nil {
		return true, "no_session"
	}

	uid := strconv.FormatInt(userID, 10)
	if s.Get(adminUIDKey) != uid {
		s.Put(adminUIDKey, uid)
		s.Put(adminSinceKey, now.Unix())
		s.Put(adminSeenKey, now.Unix())

		return false, ""
	}

	since, okSince := unix(s.Get(adminSinceKey))
	seen, okSeen := unix(s.Get(adminSeenKey))

	switch {
	case !okSince || !okSeen:
		return true, "clock_missing"
	case now.Sub(since) > AdminMaxLifetime:
		return true, "absolute"
	case now.Sub(seen) > AdminIdleTimeout:
		return true, "idle"
	}

	s.Put(adminSeenKey, now.Unix())

	return false, ""
}

// unix reads a session value written as Unix seconds; stores may hand it
// back as int64, float64 (JSON) or a string.
func unix(v any) (time.Time, bool) {
	var n int64

	switch x := v.(type) {
	case int64:
		n = x
	case int:
		n = int64(x)
	case float64:
		n = int64(x)
	case string:
		var err error
		if n, err = strconv.ParseInt(x, 10, 64); err != nil {
			return time.Time{}, false
		}
	default:
		if s, ok := v.(fmt.Stringer); ok {
			var err error
			if n, err = strconv.ParseInt(s.String(), 10, 64); err != nil {
				return time.Time{}, false
			}
		} else {
			return time.Time{}, false
		}
	}

	return time.Unix(n, 0), true
}
