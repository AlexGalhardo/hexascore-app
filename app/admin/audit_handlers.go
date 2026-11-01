package admin

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/export"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/app/moderation"
)

var (
	auditSources = []string{"community", "admin", "revert", "merge", "create"}
	auditTargets = []string{"item", "edition", "entity"}
)

// Reverter undoes a catalog_audit row (community.Applier).
type Reverter interface {
	Revert(ctx context.Context, id int64, mode, reason string, actor community.Actor) error
}

type auditView struct {
	Page    moderation.AuditPage
	Filter  moderation.AuditFilter
	NextURL string
	Sources []string
	Targets []string
}

// Audit renders GET /admin/catalog-audit (viewCatalogAudit): the audit log with the revert form (SDD 04 section 6).
func (h ModerationHandlers) Audit(c *web.Ctx) error {
	if err := authorizeOr403(h.AuthorizeAudit, c); err != nil {
		return err
	}

	f := auditFilter(c)

	page, err := h.Queue.Audit(c.Context(), f)
	if err := mapError(err); err != nil {
		return err
	}

	v := auditView{Page: page, Filter: f, Sources: auditSources, Targets: auditTargets}
	if page.Next > 0 {
		q := c.Request().URL.Query()
		q.Set("before", strconv.FormatInt(page.Next, 10))
		v.NextURL = "/admin/catalog-audit?" + q.Encode()
	}

	d := h.Pages.newPage("Catalogue audit")
	d.Audit = v
	d.CanRevert = h.AuthorizeRevert != nil && h.AuthorizeRevert(c) == nil
	d.RevertReasons = community.RevertReasons

	return h.Pages.render(c, "catalog_audit", d)
}

// auditFilter reads the list filters shared by the page and the exports.
func auditFilter(c *web.Ctx) moderation.AuditFilter {
	before, _ := strconv.ParseInt(c.Query("before"), 10, 64)

	return moderation.AuditFilter{
		TargetType: oneOf(c.Query("target"), auditTargets),
		TargetID:   c.Query("target_id"),
		Field:      c.Query("field"),
		Source:     oneOf(c.Query("source"), auditSources),
		Before:     before,
	}
}

// AuditExport serves GET /admin/catalog-audit/export.csv and export.ndjson (viewCatalogAudit, FR-04-15): every row
// matching the page filters, streamed in the request (vukem export; the CSV formula guard stays on).
func (h ModerationHandlers) AuditExport(format string) web.HandlerFunc {
	return func(c *web.Ctx) error {
		if err := authorizeOr403(h.AuthorizeAudit, c); err != nil {
			return err
		}

		f := auditFilter(c)
		if f.TargetID != "" {
			if _, err := uuid.Parse(f.TargetID); err != nil {
				return errs.NewValidation().Add("target_id", "The target id must be a uuid.")
			}
		}

		rows := h.Queue.AuditRows(c.Context(), f)
		w, r := c.Writer(), c.Request()

		if format == "csv" {
			return export.CSV(w, r, "catalog-audit.csv", moderation.AuditExportHeader, rows, moderation.AuditRecord,
				export.FlushEvery(1000))
		}

		return export.NDJSON(w, r, "catalog-audit.ndjson", rows, export.FlushEvery(1000))
	}
}

// Revert handles POST /admin/catalog-audit/{id}/revert (revertCatalog, admins only).
func (h ModerationHandlers) Revert(c *web.Ctx) error {
	if err := authorizeOr403(h.AuthorizeRevert, c); err != nil {
		return err
	}

	u, err := h.Pages.actor(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return errs.ErrNotFound
	}

	if h.Reverter == nil {
		return errs.New(http.StatusServiceUnavailable, "Reverting is not available.")
	}

	err = h.Reverter.Revert(c.Context(), id, c.FormValue("mode"), c.FormValue("reason"), actorOf(u))

	var rc *community.RevertConflictError

	switch {
	case errors.As(err, &rc):
		return errs.New(http.StatusConflict, rc.Error())
	case errors.Is(err, community.ErrAuditNotFound):
		return errs.ErrNotFound
	case errors.Is(err, community.ErrAdminOnly):
		return errs.New(http.StatusForbidden, "Only admins can revert catalogue changes.")
	}

	if err := mapError(err); err != nil {
		return err
	}

	return c.Redirect(auditBack(c))
}

// authorizeOr403 runs an ability check and fails closed when it is not wired.
func authorizeOr403(check func(*web.Ctx) error, c *web.Ctx) error {
	if check == nil {
		return errs.New(http.StatusForbidden, "Forbidden.")
	}

	return check(c)
}

// auditBack returns the audit list URL the form came from (a same-site path only).
func auditBack(c *web.Ctx) string {
	u, err := url.Parse(c.FormValue("return"))
	if err != nil || u.Host != "" || u.Scheme != "" || u.Path != "/admin/catalog-audit" {
		return "/admin/catalog-audit"
	}

	return u.RequestURI()
}
