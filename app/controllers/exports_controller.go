package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Exports requests the signed-in user's data export (SDD 14 §6.2, §6.3): the session guard answers 303 back to
// the profile, the token guard 202 {exportId, status}.
type Exports struct {
	svc         *library.Exports
	guard       string
	api         bool
	problemBase string
}

// NewExports builds the handlers; api selects JSON answers, problemBase prefixes problem types.
func NewExports(svc *library.Exports, guard string, api bool, problemBase string) *Exports {
	return &Exports{svc: svc, guard: guard, api: api, problemBase: problemBase}
}

// Create queues the export, or answers 429 export-cooldown.
func (h *Exports) Create(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	id, err := h.svc.Request(c.Context(), u.ID)
	if cd, ok := errors.AsType[library.ExportCooldownError](err); ok {
		return h.cooldown(c, cd)
	}

	if err != nil {
		return err
	}

	if !h.api {
		return c.Back("/profile")
	}

	return c.JSON(http.StatusAccepted, map[string]any{"exportId": id, "status": "queued"})
}

// cooldown writes the 429 export-cooldown problem. workaround(vukem): VF-NEW-PROBLEM-TYPE: errs.HTTPError cannot
// set the problem type or title, so the problem is written by hand like the app token problems.
func (h *Exports) cooldown(c *web.Ctx, cd library.ExportCooldownError) error {
	raw, err := json.Marshal(map[string]any{
		"type":       strings.TrimSuffix(h.problemBase, "/") + "/" + library.CodeExportCooldown,
		"title":      "You can request one export every 7 days.",
		"status":     http.StatusTooManyRequests,
		"code":       library.CodeExportCooldown,
		"retryAfter": cd.RetryAfter,
	})
	if err != nil {
		return err
	}

	c.SetHeader("Retry-After", strconv.Itoa(cd.RetryAfter))

	return c.Blob(http.StatusTooManyRequests, "application/problem+json", raw)
}
