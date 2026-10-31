// Package community serves the community suggestion routes (SDD 04 §6).
package community

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/database"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/flags"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/community"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

const maxBody = 256 << 10

// Suggest serves the suggest page and the draft save routes.
type Suggest struct {
	db    *database.DB
	flags *flags.Flags
	guard string
	now   func() time.Time
}

// NewSuggest builds the handlers.
func NewSuggest(db *database.DB, ff *flags.Flags, guard string, now func() time.Time) *Suggest {
	return &Suggest{db: db, flags: ff, guard: guard, now: now}
}

type changeForm struct {
	Before json.RawMessage        `json:"before"`
	After  json.RawMessage        `json:"after"`
	Add    []community.RelationOp `json:"add"`
	Remove []community.RelationOp `json:"remove"`
}

type draftForm struct {
	TargetKind string                `json:"targetKind"`
	TargetID   string                `json:"targetId"`
	Domain     string                `json:"domain"`
	Version    int                   `json:"version"`
	Locale     string                `json:"locale"`
	Note       string                `json:"note"`
	SourceURLs []string              `json:"sourceUrls"`
	Changes    map[string]changeForm `json:"changes"`
	Website    string                `json:"website"`
}

func (h *Suggest) open(c *web.Ctx) bool {
	return h.flags != nil && h.flags.Enabled(c.Context(), community.SubmissionsFlag)
}

// Edit renders Community/SuggestEdit for the target in the path; with the flag off it renders the notice only.
func (h *Suggest) Edit(c *web.Ctx) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}

	if !h.open(c) {
		return inertia.Render(c, "Community/SuggestEdit", inertia.Props{"open": false})
	}

	t := community.Target{Kind: c.Param("target_type"), ID: c.Param("id")}
	page, err := community.LoadEditPage(c.Context(), h.db, u.ID, t, c.Query("domain"), h.now())
	if errors.Is(err, community.ErrTargetNotFound) {
		return errs.ErrNotFound
	}
	if err != nil {
		return err
	}

	return inertia.Render(c, "Community/SuggestEdit", inertia.Props{
		"open":   true,
		"target": page.Target,
		"fields": page.Fields,
		"locks":  page.Locks,
		"quota":  page.Quota,
		"draft":  page.Draft,
	})
}

// Create saves a new draft (POST /suggestions).
func (h *Suggest) Create(c *web.Ctx) error { return h.save(c, "") }

// Update saves the draft in {id} (PATCH /suggestions/{id}).
func (h *Suggest) Update(c *web.Ctx) error { return h.save(c, c.Param("id")) }

func (h *Suggest) save(c *web.Ctx, id string) error {
	u, err := auth.User[models.User](c, h.guard)
	if err != nil {
		return err
	}
	if !h.open(c) {
		return errs.New(http.StatusForbidden, "Suggestions are closed").
			WithExtension("code", "community-submissions-closed")
	}

	var f draftForm
	if err := json.NewDecoder(http.MaxBytesReader(c.Writer(), c.Request().Body, maxBody)).Decode(&f); err != nil {
		return errs.NewValidation().Add("body", "invalid")
	}

	now := h.now()
	if err := community.RejectHoneypot(c.Context(), h.db, c.Request(), u.ID, f.Website, now); err != nil {
		return problemError(err)
	}

	t := community.Target{Kind: f.TargetKind, ID: f.TargetID}
	domain, err := community.TargetDomain(c.Context(), h.db, t, f.Domain)
	if errors.Is(err, community.ErrTargetNotFound) {
		return errs.ErrNotFound
	}
	if err != nil {
		return err
	}

	in := community.SuggestionInput{
		Kind: "edit", TargetType: domain, Locale: f.Locale, Note: f.Note, SourceURLs: f.SourceURLs,
		Changes: make(map[string]community.FieldChange, len(f.Changes)),
	}
	for k, ch := range f.Changes {
		in.Changes[k] = community.FieldChange{Before: ch.Before, After: ch.After, Add: ch.Add, Remove: ch.Remove}
	}

	if _, _, err := community.SaveDraft(c.Context(), h.db, u.ID, id, f.Version, t, in, now); err != nil {
		if errors.Is(err, community.ErrDraftNotFound) {
			return errs.ErrNotFound
		}

		return problemError(err)
	}

	return c.Back("/suggest/" + t.Kind + "/" + t.ID)
}

// problemError maps community errors onto the SDD 04 §6 problem table; unknown errors pass through.
func problemError(err error) error {
	if v, ok := errors.AsType[*community.ValidationError](err); ok {
		return errs.NewValidation().Add(v.Field, v.Msg)
	}
	if p, ok := errors.AsType[*community.Problem](err); ok {
		e := errs.New(p.Status, "").WithExtension("code", p.Code)
		if p.Field != "" {
			e = e.WithExtension("field", p.Field).WithExtension("lockReason", p.LockReason)
		}
		if p.ExistingID != "" {
			e = e.WithExtension("existingId", p.ExistingID)
		}
		if p.EligibleAt != nil {
			e = e.WithExtension("eligibleAt", p.EligibleAt)
		}
		if p.CooldownUntil != nil {
			e = e.WithExtension("cooldownUntil", p.CooldownUntil)
		}
		if p.RetryAfter > 0 {
			e = e.WithExtension("limit", p.Limit).WithExtension("used", p.Used).
				WithExtension("resetAt", p.ResetAt).WithHeader("Retry-After", strconv.Itoa(p.RetryAfter))
		}

		return e
	}
	if t, ok := errors.AsType[*community.TransitionError](err); ok {
		e := errs.New(t.Status, "").WithExtension("code", t.Code)
		if t.Code == "stale-before" {
			return e.WithExtension("field", t.Field).WithExtension("current", t.Current)
		}

		return e.WithExtension("from", t.From).WithExtension("to", t.To)
	}

	return err
}
