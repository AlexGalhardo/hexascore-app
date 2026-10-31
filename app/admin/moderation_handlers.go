package admin

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/moderation"
)

// ModerationHandlers serve the moderation pages (SDD 04 §6): server-rendered views inside the panel layout, gated by
// the moderate ability rather than the panel's admin gate.
type ModerationHandlers struct {
	Pages CatalogHandlers
	Queue moderation.Queue
	// Authorize checks the moderate ability.
	Authorize func(c *web.Ctx) error
}

type queueView struct {
	Rows       []moderation.Row
	Filter     moderation.Filter
	NextURL    string
	UserID     int64
	Statuses   []string
	Kinds      []string
	Domains    []string
	AgeBuckets []string
}

var (
	queueStatuses = []string{"submitted", "in_review"}
	queueKinds    = []string{"edit", "create", "merge"}
	queueDomains  = []string{"game", "movie", "series", "book"}
	queueAges     = []string{moderation.AgeGreen, moderation.AgeAmber, moderation.AgeRed}
)

// oneOf returns v when it is in allowed, else "" (an unknown filter value is ignored, not an error).
func oneOf(v string, allowed []string) string {
	if slices.Contains(allowed, v) {
		return v
	}

	return ""
}

func (h ModerationHandlers) filter(c *web.Ctx) moderation.Filter {
	num := func(k string) int64 {
		n, _ := strconv.ParseInt(c.Query(k), 10, 64)

		return n
	}

	return moderation.Filter{
		Status:     oneOf(c.Query("status"), queueStatuses),
		Kind:       oneOf(c.Query("kind"), queueKinds),
		TargetType: oneOf(c.Query("target"), []string{"item", "edition", "entity"}),
		Domain:     oneOf(c.Query("domain"), queueDomains),
		Locale:     oneOf(c.Query("locale"), []string{"en", "pt-BR", "es"}),
		Age:        oneOf(c.Query("age"), queueAges),
		Field:      c.Query("field"),
		Author:     num("author"),
		Assignee:   num("assignee"),
		Competing:  c.Query("competing") == "1",
		Stale:      c.Query("stale") == "1",
		After:      c.Query("after"),
	}
}

// Index renders GET /admin/moderation: the queue, oldest first.
func (h ModerationHandlers) Index(c *web.Ctx) error {
	u, err := h.actor(c)
	if err != nil {
		return err
	}

	f := h.filter(c)

	page, err := h.Queue.List(c.Context(), f)
	if errors.Is(err, moderation.ErrBadCursor) {
		return errs.ErrNotFound
	}

	if err != nil {
		return err
	}

	q := queueView{
		Rows: page.Rows, Filter: f, UserID: u.ID,
		Statuses: queueStatuses, Kinds: queueKinds, Domains: queueDomains, AgeBuckets: queueAges,
	}

	if page.Next != "" {
		v := c.Request().URL.Query()
		v.Set("after", page.Next)
		q.NextURL = "/admin/moderation?" + v.Encode()
	}

	d := h.Pages.newPage("Moderation")
	d.Queue = q
	d.RejectReasons = moderation.RejectReasons

	return h.Pages.render(c, "moderation", d)
}

func (h ModerationHandlers) actor(c *web.Ctx) (models.User, error) {
	if err := h.Authorize(c); err != nil {
		return models.User{}, err
	}

	return auth.UserFrom[models.User](c.Context(), h.Pages.Accounts.Web.Name())
}

func actorOf(u models.User) community.Actor {
	role := "admin"
	if u.Role == models.RoleModerator {
		role = "moderator"
	}

	return community.Actor{ID: sql.NullInt64{Int64: u.ID, Valid: true}, Role: role}
}

// Claim handles POST /admin/moderation/{id}/claim.
func (h ModerationHandlers) Claim(c *web.Ctx) error {
	return h.move(c, h.Queue.Claim)
}

// Release handles POST /admin/moderation/{id}/release.
func (h ModerationHandlers) Release(c *web.Ctx) error {
	return h.move(c, h.Queue.Release)
}

// Approve handles POST /admin/moderation/{id}/approve.
func (h ModerationHandlers) Approve(c *web.Ctx) error {
	return h.move(c, func(ctx context.Context, id string, version int, a community.Actor) error {
		return h.Queue.Approve(ctx, id, version, community.ApproveOptions{
			ConfirmStale: c.FormValue("confirm_stale") == "1", ConfirmTitle: c.FormValue("confirm_title"),
		}, a)
	})
}

// Reject handles POST /admin/moderation/{id}/reject.
func (h ModerationHandlers) Reject(c *web.Ctx) error {
	return h.move(c, func(ctx context.Context, id string, version int, a community.Actor) error {
		return h.Queue.Reject(ctx, id, version, c.FormValue("reason"), c.FormValue("note"), a)
	})
}

func (h ModerationHandlers) move(
	c *web.Ctx,
	fn func(ctx context.Context, id string, version int, actor community.Actor) error,
) error {
	u, err := h.actor(c)
	if err != nil {
		return err
	}

	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		return errs.ErrNotFound
	}

	version, err := strconv.Atoi(c.FormValue("version"))
	if err != nil {
		return errs.NewValidation().Add("version", "The version is required.")
	}

	if err := mapError(fn(c.Context(), id, version, actorOf(u))); err != nil {
		return err
	}

	return c.Redirect(backTo(c))
}

func mapError(err error) error {
	var (
		te *community.TransitionError
		pe *community.Problem
		ve *community.ValidationError
		ie *moderation.InvalidError
	)

	switch {
	case err == nil:
		return nil
	case errors.Is(err, moderation.ErrNotFound):
		return errs.ErrNotFound
	case errors.Is(err, moderation.ErrNotAssignee):
		return errs.New(http.StatusForbidden, "This claim belongs to another moderator.")
	case errors.Is(err, moderation.ErrApplyUnavailable):
		return errs.New(http.StatusServiceUnavailable, "Applying suggestions is not available yet.")
	case errors.Is(err, moderation.ErrNoAuthor):
		return errs.New(http.StatusConflict, "no-author")
	case errors.Is(err, moderation.ErrBulkSize):
		return errs.NewValidation().Add("ids", "Select between 1 and 50 suggestions.")
	case errors.Is(err, moderation.ErrBulkAction):
		return errs.NewValidation().Add("action", "Unknown bulk action.")
	case errors.As(err, &ie):
		return errs.NewValidation().Add(ie.Field, ie.Msg)
	case errors.Is(err, community.ErrApplyDeadlock):
		return errs.New(http.StatusServiceUnavailable, "The catalogue was busy. Try the approval again.")
	case errors.As(err, &pe):
		return errs.New(pe.Status, pe.Code)
	case errors.As(err, &ve):
		return errs.NewValidation().Add(ve.Field, ve.Msg)
	case errors.As(err, &te):
		return errs.New(te.Status, te.Code)
	default:
		return err
	}
}

// Detail renders GET /admin/moderation/{id}.
func (h ModerationHandlers) Detail(c *web.Ctx) error {
	if _, err := h.actor(c); err != nil {
		return err
	}

	d, err := h.Queue.Detail(c.Context(), c.Param("id"))
	if err := mapError(err); err != nil {
		return err
	}

	p := h.Pages.newPage("Suggestion")
	p.Detail = d
	p.RejectReasons = moderation.RejectReasons

	return h.Pages.render(c, "moderation_detail", p)
}

// Clone handles POST /admin/moderation/{id}/clone: admins only.
func (h ModerationHandlers) Clone(c *web.Ctx) error {
	u, err := h.actor(c)
	if err != nil {
		return err
	}

	if u.Role != models.RoleAdmin {
		return errs.New(http.StatusForbidden, "Only admins can clone a suggestion.")
	}

	if _, err := h.Queue.CloneAsDraft(c.Context(), c.Param("id"), actorOf(u)); err != nil {
		return mapError(err)
	}

	return c.Redirect("/admin/moderation")
}

// Bulk handles POST /admin/moderation/bulk and renders the per-id outcomes.
func (h ModerationHandlers) Bulk(c *web.Ctx) error {
	u, err := h.actor(c)
	if err != nil {
		return err
	}

	r := c.Request()
	if err := r.ParseForm(); err != nil {
		return errs.NewValidation().Add("ids", "Invalid form.")
	}

	res, err := h.Queue.Bulk(c.Context(), r.PostForm.Get("action"), r.PostForm["ids"], r.PostForm.Get("reason"),
		r.PostForm.Get("note"), r.PostForm.Get("confirm_stale") == "1", actorOf(u))
	if err := mapError(err); err != nil {
		return err
	}

	p := h.Pages.newPage("Bulk result")
	p.BulkResults = res

	return h.Pages.render(c, "moderation_bulk", p)
}

// backTo returns the queue URL the form came from (a same-site path only), else the bare queue.
func backTo(c *web.Ctx) string {
	u, err := url.Parse(c.FormValue("return"))
	if err != nil || u.Host != "" || u.Scheme != "" || u.Path != "/admin/moderation" {
		return "/admin/moderation"
	}

	return u.RequestURI()
}
