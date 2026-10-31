package admin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/community"
)

const maxEditFields = 20

var editTargets = map[string]bool{"item": true, "entity": true, "edition": true}

// EditFields serves POST /admin/catalog/{type}/{id}/fields: paired field/value form inputs (suggestion field names,
// plain text; empty clears) written through community.Applier.AdminEdit with one catalog_audit row each. It is a
// catalogue write, so it is admin-only (applySuggestion, owner Q-A15) on top of the panel gate.
func (h CatalogHandlers) EditFields(c *web.Ctx) error {
	actor, err := h.actor(c)
	if err != nil {
		return err
	}

	if err := authorizeOr403(h.AuthorizeApply, c); err != nil {
		return err
	}

	targetType := c.Param("type")
	parsed, err := uuid.Parse(c.Param("id"))
	if !editTargets[targetType] || err != nil {
		return errs.ErrNotFound
	}

	targetID := parsed.String()

	if err := c.Request().ParseForm(); err != nil {
		return errs.NewValidation().Add("field", "The form could not be read.")
	}

	names, raw := c.Request().PostForm["field"], c.Request().PostForm["value"]
	if len(names) == 0 || len(names) != len(raw) || len(names) > maxEditFields {
		return errs.NewValidation().Add("field", "Send one value per field.")
	}

	values := make(map[string]json.RawMessage, len(names))
	for i, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || name == "cover" {
			return errs.NewValidation().Add("field", "This field cannot be edited here.")
		}

		v := strings.TrimSpace(raw[i])
		if _, err := time.Parse(time.DateOnly, v); name == "release_date" && v != "" && err != nil {
			return errs.NewValidation().Add("value", "Dates use YYYY-MM-DD.")
		}

		values[name] = editValue(name, v)
	}

	if h.Editor == nil {
		return errs.New(http.StatusServiceUnavailable, "Catalogue edits are not available.")
	}

	if _, err := h.Editor.AdminEdit(c.Context(), targetType, targetID, values, community.Actor{
		ID: sql.NullInt64{Int64: actor.ID, Valid: true}, Role: "admin",
	}); err != nil {
		return mapEditErr(err)
	}

	if targetType == "item" {
		return c.Redirect("/admin/items/" + targetID + "/manage")
	}

	return c.Redirect("/admin/catalog-audit")
}

// editValue encodes a form value as the JSON the write path expects: null when empty, a day-precision date object
// for release_date, a string otherwise.
func editValue(field, v string) json.RawMessage {
	if v == "" {
		return json.RawMessage("null")
	}

	if field == "release_date" {
		b, _ := json.Marshal(map[string]string{"value": v, "precision": "day"})
		return b
	}

	b, _ := json.Marshal(v)

	return b
}

func mapEditErr(err error) error {
	var te *community.TransitionError
	if errors.As(err, &te) && te.Code == "target-gone" {
		return errs.ErrNotFound
	}

	if errors.Is(err, catalog.ErrUnsupportedField) {
		return errs.NewValidation().Add("field", "This field cannot be edited here.")
	}

	return err
}
