package admin

import "github.com/AlexGalhardo/vukem/web"

// ModerationHandlers serve the moderation pages (SDD 04 §6): server-rendered views inside the panel layout, gated by
// the moderate ability rather than the panel's admin gate.
type ModerationHandlers struct {
	Pages CatalogHandlers
	// Authorize checks the moderate ability.
	Authorize func(c *web.Ctx) error
}

// Index renders GET /admin/moderation. The queue itself arrives with P4.B.1.2; this is the page shell.
func (h ModerationHandlers) Index(c *web.Ctx) error {
	if err := h.Authorize(c); err != nil {
		return err
	}

	return h.Pages.render(c, "moderation", h.Pages.newPage("Moderation"))
}
