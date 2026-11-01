package admin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/entities"
)

const reviewPageSize = 50

// IngestReviewHandlers serve the entity match queue (SDD 06 §6): runIngestion, admin-only, CSRF on every POST.
type IngestReviewHandlers struct {
	Pages   CatalogHandlers
	Reviews entities.ReviewQueue
	// Authorize checks the runIngestion ability.
	Authorize func(c *web.Ctx) error
}

// List renders GET /admin/ingest/reviews with the oldest open reviews.
func (h IngestReviewHandlers) List(c *web.Ctx) error {
	if err := authorizeOr403(h.Authorize, c); err != nil {
		return err
	}

	rows, err := h.Reviews.Open(c.Context(), reviewPageSize)
	if err != nil {
		return err
	}

	d := h.Pages.newPage("Match reviews")
	d.Reviews = rows

	return h.Pages.render(c, "ingest_reviews", d)
}

// Resolve handles POST /admin/ingest/reviews/{id}: action link (with entity_id), create or dismiss.
func (h IngestReviewHandlers) Resolve(c *web.Ctx) error {
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

	err = h.Reviews.Resolve(c.Context(), id, c.FormValue("action"), c.FormValue("entity_id"), u.ID)

	switch {
	case errors.Is(err, entities.ErrReviewNotFound):
		return errs.ErrNotFound
	case errors.Is(err, entities.ErrReviewResolved):
		return errs.New(http.StatusConflict, "This review was already resolved.")
	case errors.Is(err, entities.ErrInvalidReview):
		return errs.New(http.StatusUnprocessableEntity, "The action does not fit this review.")
	case err != nil:
		return err
	}

	return c.Redirect("/admin/ingest/reviews")
}
