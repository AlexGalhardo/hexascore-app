package controllers

import (
	"errors"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"
	"github.com/AlexGalhardo/vukem/z"

	"github.com/AlexGalhardo/hexascore-app/app/library"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/notifications"
)

// Notifications serves the inbox for one guard: the session guard renders the
// Inertia page and answers 303 back, the token guard answers JSON (SDD 14 §6.2, §6.3).
type Notifications struct {
	db    *database.DB
	guard string
	api   bool
}

// NewNotifications builds the handlers; api selects JSON answers.
func NewNotifications(db *database.DB, guard string, api bool) *Notifications {
	return &Notifications{db: db, guard: guard, api: api}
}

// MarkReadForm is the API body: ids (at most 100) or all.
type MarkReadForm struct {
	IDs []string `json:"ids" form:"ids"`
	All bool     `json:"all" form:"all"`
}

// MarkReadSchema validates MarkReadForm.
var MarkReadSchema = z.Struct[MarkReadForm](z.Shape{
	"ids": z.Slice(z.String().Max(64)).Max(notifications.MaxReadIDs).Default([]string{}),
	"all": z.Bool().Default(false),
}).Refine(func(f MarkReadForm) error {
	if !f.All && len(f.IDs) == 0 {
		return z.FieldError("ids", "Send ids or all.")
	}

	return nil
})

// Index lists the signed-in user's notifications, newest first.
func (h *Notifications) Index(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	limit := notifications.PageSize
	page, err := notifications.List(c.Context(), h.db, u.ID, c.Query("unread") != "", c.Query("cursor"), limit)
	if err != nil {
		return err
	}

	if h.api {
		unread, err := notifications.Unread(c.Context(), h.db, u.ID)
		if err != nil {
			return err
		}

		return c.JSON(200, map[string]any{
			"data": page.Data, "meta": map[string]any{"next_cursor": page.Next, "unread": unread},
		})
	}

	meta := inertia.ScrollMeta{PageName: "cursor"}
	if page.Next != "" {
		meta.NextPage = page.Next
	}

	return inertia.Render(c, "Account/Notifications", inertia.Props{
		"notifications": inertia.Scroll(map[string]any{"data": page.Data}, meta),
	})
}

// Read marks the notification in {id} read when it is the user's.
func (h *Notifications) Read(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	if err := notifications.MarkRead(c.Context(), h.db, u.ID, []string{c.Param("id")}); err != nil {
		return notFound(err)
	}

	return c.Back("/me/notifications")
}

// ReadAll marks every notification of the user read.
func (h *Notifications) ReadAll(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	if err := notifications.MarkAllRead(c.Context(), h.db, u.ID); err != nil {
		return err
	}

	return c.Back("/me/notifications")
}

// ReadMany is the API body form: {ids} or {all:true}.
func (h *Notifications) ReadMany(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	in, err := web.Parse(c, MarkReadSchema)
	if err != nil {
		return err
	}

	if in.All {
		err = notifications.MarkAllRead(c.Context(), h.db, u.ID)
	} else {
		err = notifications.MarkRead(c.Context(), h.db, u.ID, in.IDs)
	}

	if err != nil {
		return notFound(err)
	}

	return c.NoContent()
}

func notFound(err error) error {
	if errors.Is(err, notifications.ErrNotFound) {
		return library.ErrNotFound()
	}

	return err
}
