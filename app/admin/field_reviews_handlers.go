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
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
)

// FieldReviewHandlers serve the low-confidence field queue (RFC-001 section 5 D): runIngestion, admin-only, CSRF on
// every POST.
type FieldReviewHandlers struct {
	Pages CatalogHandlers
	DB    *database.DB
	// Authorize checks the runIngestion ability.
	Authorize func(c *web.Ctx) error
}

// List renders GET /admin/ingest/field-reviews with the oldest open reviews.
func (h FieldReviewHandlers) List(c *web.Ctx) error {
	if err := authorizeOr403(h.Authorize, c); err != nil {
		return err
	}

	rows, err := store.OpenReviews(c.Context(), h.DB, reviewPageSize)
	if err != nil {
		return err
	}

	d := h.Pages.newPage("Field reviews")
	d.FieldReviews = rows

	return h.Pages.render(c, "ingest_field_reviews", d)
}

// Resolve handles POST /admin/ingest/field-reviews/{id}: accept or reject.
func (h FieldReviewHandlers) Resolve(c *web.Ctx) error {
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

	resolution := store.ResolutionRejected
	if c.FormValue("action") == "accept" {
		resolution = store.ResolutionAccepted
	} else if c.FormValue("action") != "reject" {
		return errs.New(http.StatusUnprocessableEntity, "The action does not fit this review.")
	}

	err = store.ResolveReview(c.Context(), h.DB, id, resolution, u.ID)

	switch {
	case errors.Is(err, store.ErrReviewNotFound):
		return errs.ErrNotFound
	case errors.Is(err, store.ErrReviewResolved):
		return errs.New(http.StatusConflict, "This review was already resolved.")
	case errors.Is(err, store.ErrInvalidReview):
		return errs.New(http.StatusUnprocessableEntity, "The value cannot be applied to this item.")
	case err != nil:
		return err
	}

	return c.Redirect("/admin/ingest/field-reviews")
}
