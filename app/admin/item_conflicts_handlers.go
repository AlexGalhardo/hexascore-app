package admin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/resolve"
)

// ItemConflictHandlers serve the queue of items the normaliser held back because they duplicate another (RFC-001 section
// 5 C): runIngestion, admin-only, CSRF on every POST.
type ItemConflictHandlers struct {
	Pages CatalogHandlers
	DB    *database.DB
	// Authorize checks the runIngestion ability.
	Authorize func(c *web.Ctx) error
}

// List renders GET /admin/ingest/conflicts with the oldest open conflicts.
func (h ItemConflictHandlers) List(c *web.Ctx) error {
	if err := authorizeOr403(h.Authorize, c); err != nil {
		return err
	}

	rows, err := resolve.Open(c.Context(), h.DB, reviewPageSize)
	if err != nil {
		return err
	}

	return h.Pages.react(c, "Admin/IngestConflicts", map[string]any{"conflicts": rows})
}

// Resolve handles POST /admin/ingest/conflicts/{id}: allow the item in or dismiss it.
func (h ItemConflictHandlers) Resolve(c *web.Ctx) error {
	if err := authorizeOr403(h.Authorize, c); err != nil {
		return err
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return errs.ErrNotFound
	}

	u, err := auth.UserFrom[models.User](c.Context(), h.Pages.Accounts.Web.Name())
	if err != nil {
		return err
	}

	var state string

	switch c.FormValue("action") {
	case "allow":
		state = resolve.StateAllowed
	case "dismiss":
		state = resolve.StateDismissed
	default:
		return errs.New(http.StatusUnprocessableEntity, "The action does not fit this conflict.")
	}

	switch err := resolve.Decide(c.Context(), h.DB, id, state, u.ID); {
	case errors.Is(err, resolve.ErrNotFound):
		return errs.ErrNotFound
	case errors.Is(err, resolve.ErrResolved):
		return errs.New(http.StatusConflict, "This conflict was already resolved.")
	case err != nil:
		return err
	}

	return c.Redirect("/admin/ingest/conflicts")
}
