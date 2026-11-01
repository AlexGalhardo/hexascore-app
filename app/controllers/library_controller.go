package controllers

import (
	"errors"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"
	"github.com/AlexGalhardo/vukem/z"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// Library serves the personal library and the book edition choice for one guard (SDD 14 §6.2, §6.3).
type Library struct {
	db    *database.DB
	now   func() time.Time
	guard string
	api   bool
}

// NewLibrary builds the handlers; api selects JSON answers.
func NewLibrary(db *database.DB, now func() time.Time, guard string, api bool) *Library {
	return &Library{db: db, now: now, guard: guard, api: api}
}

// EditionForm is the body of PUT /items/{id}/edition.
type EditionForm struct {
	EditionID string `json:"editionId" form:"editionId"`
}

// EditionSchema validates EditionForm; the edition's ownership is checked against the database.
var EditionSchema = z.Struct[EditionForm](z.Shape{"editionId": z.String().Max(64)})

// Index renders Library/Index, or answers the same groups as JSON.
func (h *Library) Index(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	f := library.Filter{
		Status: c.Query("status"), Type: c.Query("type"), Sort: c.Query("sort"),
		Offset: library.ParseOffset(c.Query("cursor")),
	}

	view, err := library.Library(c.Context(), h.db, u.ID, i18n.Locale(c.Context()), f)
	if err != nil {
		return err
	}

	if h.api {
		return c.JSON(200, view)
	}

	meta := inertia.ScrollMeta{PageName: "cursor"}
	if f.Type != "" && len(view.Groups) == 1 && view.Groups[0].Next >= 0 {
		meta.NextPage = strconv.Itoa(view.Groups[0].Next)
	}

	return inertia.Render(c, "Library/Index", inertia.Props{
		"counts":  view.Counts,
		"groups":  inertia.Scroll(map[string]any{"data": view.Groups}, meta),
		"filters": map[string]string{"status": f.Status, "type": f.Type, "sort": f.Sort},
	})
}

// PutEdition sets the signed-in user's edition of the book in {id}.
func (h *Library) PutEdition(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	in, err := web.Parse(c, EditionSchema)
	if err != nil {
		return err
	}

	if err := library.SetEdition(c.Context(), h.db, u.ID, c.Param("id"), in.EditionID, h.now().UTC()); err != nil {
		return err
	}

	if !h.api {
		return c.Back("/library")
	}

	return c.JSON(200, map[string]any{"itemId": c.Param("id"), "editionId": in.EditionID})
}

// Editions lists the editions of the book in {id} for the picker.
func (h *Library) Editions(c *web.Ctx) error {
	eds, next, err := catalog.LoadBookEditions(c.Context(), h.db, c.Param("id"), library.ParseOffset(c.Query("cursor")))
	if errors.Is(err, catalog.ErrItemNotFound) {
		return NotFound(c)
	}
	if err != nil {
		return err
	}

	return c.JSON(200, map[string]any{"data": eds, "next": next})
}

// DeleteEdition forgets the signed-in user's edition of the book in {id}.
func (h *Library) DeleteEdition(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	if err := library.ClearEdition(c.Context(), h.db, u.ID, c.Param("id")); err != nil {
		return err
	}

	if !h.api {
		return c.Back("/library")
	}

	return c.NoContent()
}
