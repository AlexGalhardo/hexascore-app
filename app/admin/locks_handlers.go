package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/moderation"
)

const maxLockDays = 365

// Locks renders GET /admin/locks (lockCatalog, admins only): the active locks and the lock form (SDD 04 section 6).
func (h ModerationHandlers) Locks(c *web.Ctx) error {
	if err := authorizeOr403(h.AuthorizeLock, c); err != nil {
		return err
	}

	locks, err := h.Queue.Locks(c.Context())
	if err != nil {
		return err
	}

	d := h.Pages.newPage("Catalogue locks")
	d.Locks = locks
	d.LockTargets = auditTargets

	return h.Pages.render(c, "locks", d)
}

// CreateLock handles POST /admin/locks: target, target_id, field (empty or "*" locks the record), reason and an
// optional expiry in days.
func (h ModerationHandlers) CreateLock(c *web.Ctx) error {
	if err := authorizeOr403(h.AuthorizeLock, c); err != nil {
		return err
	}

	u, err := auth.UserFrom[models.User](c.Context(), h.Pages.Accounts.Web.Name())
	if err != nil {
		return err
	}

	in := moderation.NewLock{
		TargetType: c.FormValue("target"), TargetID: strings.ToLower(strings.TrimSpace(c.FormValue("target_id"))),
		Field: c.FormValue("field"), Reason: c.FormValue("reason"), By: u.ID,
	}

	if raw := strings.TrimSpace(c.FormValue("days")); raw != "" {
		days, err := strconv.Atoi(raw)
		if err != nil || days < 1 || days > maxLockDays {
			return errs.NewValidation().Add("days", "Expiry is 1 to 365 days, or empty for none.")
		}

		in.TTL = time.Duration(days) * 24 * time.Hour
	}

	if _, err := h.Queue.CreateLock(c.Context(), in); err != nil {
		return mapError(err)
	}

	return c.Redirect("/admin/locks")
}

// DeleteLock handles POST /admin/locks/{id}/delete (HTML forms cannot send DELETE).
func (h ModerationHandlers) DeleteLock(c *web.Ctx) error {
	if err := authorizeOr403(h.AuthorizeLock, c); err != nil {
		return err
	}

	if err := h.Queue.DeleteLock(c.Context(), c.Param("id")); err != nil {
		return mapError(err)
	}

	return c.Redirect("/admin/locks")
}
