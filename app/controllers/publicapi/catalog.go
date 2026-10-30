package publicapi

import (
	"errors"
	"net/http"
	"time"

	"github.com/AlexGalhardo/vukem/api"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	hxapi "github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/models"
)

// editionCursorSort tags edition cursors so a listing cursor is rejected here.
const editionCursorSort = "editions"

// TypeInput is a path with only {type}.
type TypeInput struct {
	Type string `json:"type"`
}

// GenreJSON is one genre with its public item count (legacy parity).
type GenreJSON struct {
	Slug  string `json:"slug"  validate:"required"`
	Count int    `json:"count"`
}

// GenreListJSON is listGenres.
type GenreListJSON struct {
	Data []GenreJSON `json:"data"`
}

// ListGenres is listGenres for one type: GET /api/v1/{games,movies,series,books}/genres. Genre slugs are HexaScore
// identifiers; names stay out because default_name carries no source here. The type is part of the path
// literal because "/{type}/genres" overlaps "/entities/{id}" in the ServeMux.
func (h *Controller) ListGenres(section string) func(*web.Ctx, struct{}) (GenreListJSON, error) {
	return func(c *web.Ctx, _ struct{}) (GenreListJSON, error) { return h.genres(c, typeOf[section]) }
}

func (h *Controller) genres(c *web.Ctx, typ string) (GenreListJSON, error) {
	if err := checkQuery(c); err != nil {
		return GenreListJSON{}, err
	}

	genres, _, err := catalog.Facets(c.Context(), h.DB, typ, h.Policy.RelationSources())
	if err != nil {
		return GenreListJSON{}, err
	}

	out := GenreListJSON{Data: make([]GenreJSON, len(genres))}
	for i, g := range genres {
		out.Data[i] = GenreJSON{Slug: g.Slug, Count: g.Count}
	}

	return out, notModified(c, out)
}

// EditionJSON is one book edition; omitted fields are absent and listed in meta (FR-05-6).
type EditionJSON struct {
	ID              string   `json:"id"                        validate:"required"`
	Title           *string  `json:"title,omitempty"`
	Subtitle        *string  `json:"subtitle,omitempty"`
	Language        string   `json:"language"`
	Publisher       *string  `json:"publisher,omitempty"`
	PublicationYear *int     `json:"publicationYear,omitempty"`
	PageCount       *int     `json:"pageCount,omitempty"`
	Format          *string  `json:"format,omitempty"`
	ISBN13          *string  `json:"isbn13,omitempty"`
	Meta            ItemMeta `json:"meta"`
}

// EditionPageJSON is one page of editions.
type EditionPageJSON struct {
	Data  []EditionJSON `json:"data"`
	Meta  api.Meta      `json:"meta"`
	Links api.Links     `json:"links"`
}

// ListEditionsInput is the listEditions path and query.
type ListEditionsInput struct {
	ID     string `json:"id"`
	Limit  string `json:"limit"  doc:"1 to the plan page size, default 10"`
	Cursor string `json:"cursor" doc:"opaque next_cursor of the previous page"`
}

// ListEditions is listEditions: GET /api/v1/books/{id}/editions (D-10), paged by edition id.
func (h *Controller) ListEditions(c *web.Ctx, in ListEditionsInput) (EditionPageJSON, error) {
	if err := checkQuery(c, "limit", "cursor"); err != nil {
		return EditionPageJSON{}, err
	}
	limit, err := h.limit(c, in.Limit)
	if err != nil {
		return EditionPageJSON{}, err
	}
	if !uuidRe.MatchString(in.ID) {
		return EditionPageJSON{}, errs.ErrNotFound
	}

	after := ""
	if in.Cursor != "" {
		cur, err := h.Codec.Decode(in.Cursor, editionCursorSort)
		if err != nil {
			return EditionPageJSON{}, badQuery(err)
		}
		after = cur.ID
	}

	editions, more, err := catalog.LoadAPIEditions(c.Context(), h.DB, in.ID, after, limit, h.Policy)
	if errors.Is(err, catalog.ErrItemNotFound) {
		return EditionPageJSON{}, h.mergedEditions(c, in)
	}
	if err != nil {
		return EditionPageJSON{}, err
	}

	out := EditionPageJSON{Data: make([]EditionJSON, len(editions)), Meta: api.Meta{PerPage: limit}}
	for i, e := range editions {
		out.Data[i] = EditionJSON{
			ID: e.ID, Subtitle: e.Subtitle, Language: e.Language, Publisher: e.Publisher,
			PublicationYear: e.PublicationYear, PageCount: e.PageCount, Format: e.Format, ISBN13: e.ISBN13,
			Meta: ItemMeta{OmittedFields: append([]string{}, e.OmittedFields...)},
		}
		if e.Title != "" {
			out.Data[i].Title = &e.Title
		}
	}
	out.Links.Self = c.Request().URL.RequestURI()
	if more {
		out.Meta.NextCursor = h.Codec.Encode(
			catalog.Cursor{S: editionCursorSort, ID: editions[len(editions)-1].ID, D: catalog.DirNext},
		)
		q := c.Request().URL.Query()
		q.Set("cursor", out.Meta.NextCursor)
		next := c.Request().URL.Path + "?" + q.Encode()
		out.Links.Next = &next
	}

	return out, notModified(c, out)
}

// mergedEditions sends a merged-away book id to the surviving work's edition list (SDD 01).
func (h *Controller) mergedEditions(c *web.Ctx, in ListEditionsInput) error {
	target, ok, err := h.Redirects(c.Context(), "book", in.ID)
	if err != nil {
		return err
	}
	if !ok {
		return errs.ErrNotFound
	}
	loc := "/api/v1/books/" + target.ID + "/editions"
	if q := c.Request().URL.RawQuery; q != "" {
		loc += "?" + q
	}

	return errs.New(http.StatusMovedPermanently, "item merged").WithHeader("Location", loc)
}

// UsageQuota is the quota block of getUsage (legacy lib/api/schemas.ts usageSchema, snake_case kept).
type UsageQuota struct {
	Limit     int     `json:"limit"`
	Used      int     `json:"used"`
	Remaining int     `json:"remaining"`
	Period    string  `json:"period"    validate:"oneof=lifetime month"`
	ResetsAt  *string `json:"resets_at"`
}

// UsageRateLimit is the per-second throttle of the plan.
type UsageRateLimit struct {
	RequestsPerSecond int `json:"requests_per_second"`
}

// UsageJSON is getUsage: the legacy shape without the credits block (D-23 credits are not shown).
type UsageJSON struct {
	Plan      string         `json:"plan"       validate:"required"`
	Quota     UsageQuota     `json:"quota"`
	RateLimit UsageRateLimit `json:"rate_limit"`
}

// GetUsage is getUsage: GET /api/v1/me/usage. It counts like any request, so `used` includes this one.
func (h *Controller) GetUsage(c *web.Ctx, _ struct{}) (UsageJSON, error) {
	if err := checkQuery(c); err != nil {
		return UsageJSON{}, err
	}
	u, err := auth.User[models.User](c, hxapi.GuardName)
	if err != nil {
		return UsageJSON{}, err
	}
	plan, err := h.plan(c)
	if err != nil {
		return UsageJSON{}, err
	}
	used, err := h.Meter.Used(c.Context(), u.ID, plan)
	if err != nil {
		return UsageJSON{}, hxapi.Unavailable(err)
	}
	used = min(used, plan.Quota)

	out := UsageJSON{
		Plan:      plan.ID,
		Quota:     UsageQuota{Limit: plan.Quota, Used: used, Remaining: plan.Quota - used, Period: "lifetime"},
		RateLimit: UsageRateLimit{RequestsPerSecond: plan.RPS},
	}
	if !plan.Lifetime {
		now := h.Meter.Now().UTC()
		reset := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
		out.Quota.Period, out.Quota.ResetsAt = "month", &reset
	}

	return out, nil
}
