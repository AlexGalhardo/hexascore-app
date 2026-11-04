package admin

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/flags"
	"github.com/AlexGalhardo/vukem/web"
	"github.com/google/uuid"
)

// The two writes the framework panel used to offer and the React screens still need (ADR 0026): changing the global
// value of a feature flag and marking a contact message as handled. Both keep the panel's audit trail: one row in
// admin_audit_logs, in the shape the panel wrote, so the audit log screen reads old and new entries alike.

const (
	flagsAuditResource   = "_flags"
	contactAuditResource = "contact-messages"
	auditUpdate          = "update"
)

// WriteHandlers serve POST /admin/r/flags/{name} and POST /admin/r/contact-messages/{id}/handled.
type WriteHandlers struct {
	DB    *database.DB
	Flags *flags.Flags
	// Authorize is the panel gate check (404 for non-admins).
	Authorize func(c *web.Ctx) error
	// Actor is the signed-in admin's id as the audit log stores it.
	Actor func(ctx context.Context) string
}

type auditChange struct {
	Before any `json:"before,omitempty"`
	After  any `json:"after,omitempty"`
}

func (h WriteHandlers) audit(ctx context.Context, resource, recordID string, changes map[string]auditChange) error {
	raw, err := json.Marshal(changes)
	if err != nil {
		return fmt.Errorf("admin audit: %w", err)
	}

	_, err = h.DB.SQL().ExecContext(ctx,
		`INSERT INTO admin_audit_logs (actor_id, resource, record_id, action, changes) VALUES ($1, $2, $3, $4, $5)`,
		h.Actor(ctx), resource, recordID, auditUpdate, string(raw))

	return err
}

// auditValues is what the log keeps of a flag change: the values of bool and int flags, and only the length and a
// SHA-256 prefix of a string value, which may hold anything.
func auditValues(ch flags.Change) (before, after string) {
	if !ch.Sensitive {
		return ch.Old, ch.New
	}

	digest := func(v string) string {
		sum := sha256.Sum256([]byte(v))

		return fmt.Sprintf("%d bytes, sha256:%x", len(v), sum[:4])
	}

	return digest(ch.Old), digest(ch.New)
}

// UpdateFlag sets (form field "value") or resets (field "reset") the global value of one flag.
func (h WriteHandlers) UpdateFlag(c *web.Ctx) error {
	if err := h.Authorize(c); err != nil {
		return err
	}

	if h.Flags == nil {
		return errs.ErrNotFound
	}

	ctx, name := c.Context(), c.Param("name")

	var (
		ch  flags.Change
		err error
	)

	if c.FormValue("reset") != "" {
		ch, err = h.Flags.Reset(ctx, flags.Global, name)
	} else {
		ch, err = h.Flags.Set(ctx, flags.Global, name, c.FormValue("value"))
	}

	switch {
	case errors.Is(err, flags.ErrUnknown):
		return errs.New(http.StatusNotFound, "Unknown flag.")
	case errors.Is(err, flags.ErrLocked):
		web.SecurityEvent(c.Request(), "flags.locked_write", "actor_id", h.Actor(ctx), "flag", name)

		return errs.New(http.StatusForbidden, "This flag is locked in code.")
	case errors.Is(err, flags.ErrInvalid):
		return errs.New(http.StatusUnprocessableEntity, "Invalid value for this flag.")
	case err != nil:
		return err
	}

	if ch.Old != ch.New {
		before, after := auditValues(ch)
		web.AuditEvent(c.Request(), "flags.changed", "actor_id", h.Actor(ctx), "flag", ch.Flag,
			"changed", true, "old", before, "new", after)

		// The store write and this row are not atomic: the flag already changed, so a failed row is logged.
		if err := h.audit(ctx, flagsAuditResource, ch.Flag, map[string]auditChange{
			"value": {Before: before, After: after},
		}); err != nil {
			c.Log().ErrorContext(ctx, "admin: flag changed but its audit-log row failed", "flag", ch.Flag, "err", err)
		}
	}

	return c.Redirect("/admin/r/flags")
}

// SetContactHandled sets the handled flag of a contact message from the form field "handled" (true or false).
func (h WriteHandlers) SetContactHandled(c *web.Ctx) error {
	if err := h.Authorize(c); err != nil {
		return err
	}

	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return errs.ErrNotFound
	}

	want := c.FormValue("handled") == "true"
	ctx := c.Context()

	var before bool

	err := h.DB.SQL().QueryRowContext(ctx, `WITH old AS (SELECT handled FROM contact_messages WHERE id = $1)
UPDATE contact_messages SET handled = $2 WHERE id = $1 RETURNING (SELECT handled FROM old)`, id, want).Scan(&before)
	if err != nil {
		// No row updated: the message does not exist.
		return errs.ErrNotFound
	}

	if before != want {
		if err := h.audit(ctx, contactAuditResource, id, map[string]auditChange{
			"handled": {Before: before, After: want},
		}); err != nil {
			c.Log().
				ErrorContext(ctx, "admin: contact message changed but its audit-log row failed", "id", id, "err", err)
		}
	}

	return c.Redirect("/admin/r/contact-messages/" + id)
}
