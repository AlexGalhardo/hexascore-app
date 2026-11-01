// Package publicapi serves catalogue reads with typed handlers, so the OpenAPI
// document describes every route (https://vukem.dev/docs/rest-apis/typed-handlers).
// /api/v1 uses the "api" audience (ADR 0006, SDD 01 §6.1). /app/v1 reuses these
// handlers with the "site" audience (SDD 05 §6.4).
package publicapi

import (
	"context"
	"errors"
	"maps"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/vukem/api"
	"github.com/AlexGalhardo/vukem/auth"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	hxapi "github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	appmiddleware "github.com/AlexGalhardo/hexascore-app/app/http/middleware"
	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// Score provenance: HexaScore aggregates reach customers under the API terms (A-07).
const (
	scoreSource  = "hexascore"
	scoreLicence = "HEXASCORE-API-TERMS"
	defaultLimit = 10
)

var (
	typeOf  = map[string]string{"games": "game", "movies": "movie", "series": "series", "books": "book"}
	locales = []string{"en", "pt-BR", "es"}
	uuidRe  = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// errNotModified ends a typed handler after c.NotModified wrote the 304: the error handler sees the response
// written and only logs it at debug level.
// workaround(vukem): VF-049, typed handlers have no documented way to answer 304.
var errNotModified = errs.New(http.StatusNotModified, "not modified")

// Controller holds the catalogue read handlers shared by /api/v1 and /app/v1.
type Controller struct {
	DB     catalog.Querier
	Policy catalog.Policy
	Codec  *catalog.CursorCodec
	Meter  *hxapi.Meter
	Plans  hxapi.PlanResolver
	Search search.Searcher
	// Redirects resolves merged item ids (catalog.ResolveID needs a database.Executor).
	Redirects func(ctx context.Context, typ, id string) (catalog.Target, bool, error)
	// Now is the clock for the changes window. Nil means time.Now.
	Now func() time.Time
	// Mount is this audience's URL prefix. Empty stays /api/v1 so a merged id
	// on the paid API does not start pointing at the app.
	Mount string
	// PageCap, when set, replaces the caller's API plan page size. The
	// first-party app is not plan-metered (SDD 05 §6.4).
	PageCap int
}

// ItemMeta lists the fields the licence filter removed (SDD 05 FR-05-6).
type ItemMeta struct {
	OmittedFields []string `json:"omittedFields"`
}

// ItemJSON is one item on /api/v1: licence-filtered fields, SDD 02 §6 score fields, provenance and meta.
type ItemJSON struct {
	ID               string                         `json:"id"                         validate:"required"`
	Type             string                         `json:"type"                       validate:"required"`
	Slug             string                         `json:"slug"                       validate:"required"`
	WikidataQID      string                         `json:"wikidataQid"`
	Locale           string                         `json:"locale"                     validate:"required"`
	Title            string                         `json:"title,omitempty"`
	Synopsis         *catalog.Synopsis              `json:"synopsis,omitempty"`
	ReleaseDate      string                         `json:"releaseDate,omitempty"`
	ReleasePrecision string                         `json:"releasePrecision,omitempty"`
	RuntimeMinutes   *int                           `json:"runtimeMinutes,omitempty"`
	PageCount        *int                           `json:"pageCount,omitempty"`
	ExternalIDs      map[string]string              `json:"externalIds,omitempty"`
	Entities         []catalog.EntityRef            `json:"entities,omitempty"`
	Score            *float64                       `json:"score"`
	ScoreByCategory  map[string]float64             `json:"scoreByCategory"`
	RatingCount      int                            `json:"ratingCount"`
	Confidence       *string                        `json:"confidence"`
	FormulaVersion   *int                           `json:"formulaVersion"`
	Provenance       map[string]catalog.FieldSource `json:"provenance"`
	Meta             ItemMeta                       `json:"meta"`
}

// ItemPageJSON is one keyset page of items in the vukem envelope.
type ItemPageJSON struct {
	Data  []ItemJSON `json:"data"`
	Meta  api.Meta   `json:"meta"`
	Links api.Links  `json:"links"`
}

// ListItemsInput documents the listItems query; values are parsed by catalog.ParseListQuery so a bad value is
// a 400 invalid-query (not counted, FR-05-4) instead of a 422. `min_<category>` filters are dynamic per type.
type ListItemsInput struct {
	Type     string `json:"type"      doc:"games, movies, series or books"`
	Genre    string `json:"genre"     doc:"genre entity slug"`
	Platform string `json:"platform"  doc:"platform entity slug (games)"`
	Year     string `json:"year"      doc:"release year, 1800-2100"`
	MinScore string `json:"min_score" doc:"0-10 in steps of 0.5"`
	MaxScore string `json:"max_score" doc:"0-10 in steps of 0.5"`
	Sort     string `json:"sort"      doc:"[-]popularity, score, release_date or rating_count"`
	Limit    string `json:"limit"     doc:"1 to the plan page size, default 10"`
	Cursor   string `json:"cursor"    doc:"opaque next_cursor of the previous page"`
	Locale   string `json:"locale"    doc:"en, pt-BR or es"`
}

// unknownType names the accepted path segments, so a client can fix the request without reading the docs.
const unknownType = "unknown type: use games, movies, series or books"

// GetItemInput is the getItem path and query.
type GetItemInput struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	Locale string `json:"locale" doc:"en, pt-BR or es"`
}

// ListItems is listItems: GET /api/v1/{type}.
func (h *Controller) ListItems(c *web.Ctx, in ListItemsInput) (ItemPageJSON, error) {
	typ, ok := typeOf[in.Type]
	if !ok {
		return ItemPageJSON{}, invalid(unknownType)
	}
	allowed := []string{"genre", "platform", "year", "min_score", "max_score", "sort", "limit", "cursor", "locale"}
	for _, k := range catalog.CategoryKeys[typ] {
		allowed = append(allowed, "min_"+k)
	}
	if err := checkQuery(c, allowed...); err != nil {
		return ItemPageJSON{}, err
	}
	locale, err := localeOf(in.Locale)
	if err != nil {
		return ItemPageJSON{}, err
	}
	limit, err := h.limit(c, in.Limit)
	if err != nil {
		return ItemPageJSON{}, err
	}

	params, err := catalog.ParseListQuery(c.Query, typ)
	if err != nil {
		return ItemPageJSON{}, badQuery(err)
	}
	params.Limit = limit

	return h.itemPage(c, params, locale, limit)
}

// itemPage runs one keyset page of params and renders it in the vukem envelope.
func (h *Controller) itemPage(c *web.Ctx, params catalog.ListParams, locale string, limit int) (ItemPageJSON, error) {
	params.RelationSources = h.Policy.RelationSources()
	page, err := catalog.List(c.Context(), h.DB, h.Codec, params)
	if err != nil {
		return ItemPageJSON{}, badQuery(err)
	}

	ids := make([]string, len(page.Items))
	for i, it := range page.Items {
		ids[i] = it.ID
	}
	items, err := catalog.LoadAPIItems(c.Context(), h.DB, ids, locale, h.Policy, false)
	if err != nil {
		return ItemPageJSON{}, err
	}

	out := ItemPageJSON{Data: make([]ItemJSON, len(items)), Meta: api.Meta{PerPage: limit, NextCursor: page.NextCursor}}
	for i, it := range items {
		out.Data[i] = itemJSON(it, locale)
	}
	out.Links.Self = c.Request().URL.RequestURI()
	if page.NextCursor != "" {
		q := c.Request().URL.Query()
		q.Set("cursor", page.NextCursor)
		next := c.Request().URL.Path + "?" + q.Encode()
		out.Links.Next = &next
	}

	return out, notModified(c, out)
}

// GetItem is getItem: GET /api/v1/{type}/{id}. Merged ids answer 301 to the surviving item; hidden or unknown
// ids 404 (SDD 01 FR-01-14).
func (h *Controller) GetItem(c *web.Ctx, in GetItemInput) (ItemJSON, error) {
	typ, ok := typeOf[in.Type]
	if !ok {
		return ItemJSON{}, invalid(unknownType)
	}
	if err := checkQuery(c, "locale"); err != nil {
		return ItemJSON{}, err
	}
	locale, err := localeOf(in.Locale)
	if err != nil {
		return ItemJSON{}, err
	}
	if !uuidRe.MatchString(in.ID) {
		return ItemJSON{}, errs.ErrNotFound
	}

	items, err := catalog.LoadAPIItems(c.Context(), h.DB, []string{in.ID}, locale, h.Policy, true)
	if err != nil {
		return ItemJSON{}, err
	}
	if len(items) == 0 {
		return ItemJSON{}, h.merged(c, typ, in)
	}
	if items[0].Detail.Type != typ {
		return ItemJSON{}, errs.ErrNotFound
	}

	out := itemJSON(items[0], locale)

	return out, notModified(c, out)
}

// merged answers 301 to the item a merged-away id now points at, else 404.
func (h *Controller) merged(c *web.Ctx, typ string, in GetItemInput) error {
	t, ok, err := h.Redirects(c.Context(), typ, in.ID)
	if err != nil {
		return err
	}
	if !ok {
		return errs.ErrNotFound
	}

	return h.movedTo(c, "/"+in.Type+"/"+t.ID)
}

// publicPrefix is Mount, or /api/v1 when the paid API left it empty.
func (h *Controller) publicPrefix() string {
	if h.Mount != "" {
		return h.Mount
	}

	return "/api/v1"
}

// movedTo answers 301 to path under this controller's prefix, keeping the query.
func (h *Controller) movedTo(c *web.Ctx, path string) error {
	loc := h.publicPrefix() + path
	if q := c.Request().URL.RawQuery; q != "" {
		loc += "?" + q
	}

	return errs.New(http.StatusMovedPermanently, "item merged").WithHeader("Location", loc)
}

func itemJSON(it catalog.APIItem, locale string) ItemJSON {
	d, s := it.Detail, it.Score
	out := ItemJSON{
		ID: d.ID, Type: sectionOf(d.Type), Slug: d.Slug, WikidataQID: d.WikidataQID, Locale: locale,
		Title: d.Title, Synopsis: d.Synopsis, ReleaseDate: d.ReleaseDate, ReleasePrecision: d.ReleasePrecision,
		RuntimeMinutes: d.RuntimeMinutes, PageCount: d.PageCount, ExternalIDs: d.ExternalIDs, Entities: d.Entities,
		Score: s.Score, ScoreByCategory: s.ScoreByCategory, RatingCount: s.RatingCount, Confidence: s.Confidence,
		FormulaVersion: s.FormulaVersion,
		Provenance:     map[string]catalog.FieldSource{"score": {Source: scoreSource, Licence: scoreLicence}},
		Meta:           ItemMeta{OmittedFields: []string{}},
	}
	maps.Copy(out.Provenance, d.Provenance)
	if d.Meta != nil {
		out.Meta.OmittedFields = append(out.Meta.OmittedFields, d.Meta.OmittedFields...)
		slices.Sort(out.Meta.OmittedFields)
	}

	return out
}

func sectionOf(typ string) string {
	for section, t := range typeOf {
		if t == typ {
			return section
		}
	}

	return typ
}

// notModified sets the strong ETag and Cache-Control of v (SDD 05 §6.1) and, when If-None-Match matches, has
// written a 304 and returns errNotModified (FR-05-8; the quota still counts it, FR-05-4).
func notModified(c *web.Ctx, v any) error {
	etag, err := web.ETagOf(v)
	if err != nil {
		return err
	}
	c.SetHeader("Cache-Control", "private, no-cache")
	if c.NotModified(etag, time.Time{}) {
		return errNotModified
	}

	return nil
}

// limit parses `limit`: default 10, at most the plan page size (SDD 05 FR-05-3).
func (h *Controller) limit(c *web.Ctx, raw string) (int, error) {
	if raw == "" {
		return defaultLimit, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return 0, invalid("limit must be a positive integer")
	}
	if h.PageCap > 0 {
		if n > h.PageCap {
			return 0, invalid("limit exceeds the page size " + strconv.Itoa(h.PageCap))
		}

		return n, nil
	}
	plan, err := h.plan(c)
	if err != nil {
		return 0, err
	}
	if n > plan.MaxPageSize {
		return 0, invalid("limit exceeds the " + plan.ID + " plan page size " + strconv.Itoa(plan.MaxPageSize))
	}

	return n, nil
}

func (h *Controller) plan(c *web.Ctx) (hxapi.Plan, error) {
	u, err := auth.User[models.User](c, hxapi.GuardName)
	if err != nil {
		return hxapi.Plan{}, err
	}
	p, err := h.Plans(c.Context(), u.ID)
	if err != nil {
		return hxapi.Plan{}, hxapi.Unavailable(err)
	}

	return p, nil
}

// checkQuery rejects unknown, repeated or empty query parameters (legacy lib/api/handler.ts).
func checkQuery(c *web.Ctx, allowed ...string) error {
	return checkValues(c.Request().URL.Query(), allowed)
}

func checkValues(q url.Values, allowed []string) error {
	for k, vs := range q {
		if !slices.Contains(allowed, k) {
			return invalid("unknown parameter " + k)
		}
		if len(vs) != 1 || strings.TrimSpace(vs[0]) == "" {
			return invalid("parameter " + k + " must be given once and not empty")
		}
	}

	return nil
}

func localeOf(raw string) (string, error) {
	if raw == "" {
		return "en", nil
	}
	if !slices.Contains(locales, raw) {
		return "", invalid("locale must be en, pt-BR or es")
	}

	return raw, nil
}

// invalid is the 400 invalid-query problem; the quota refunds it (FR-05-4).
func invalid(detail string) error {
	return errs.New(http.StatusBadRequest, "invalid-query: "+detail)
}

func badQuery(err error) error {
	if errors.Is(err, catalog.ErrInvalidCursor) {
		return errs.New(http.StatusBadRequest, err.Error()).
			WithHeader(appmiddleware.ProblemCodeHeader, "invalid-cursor")
	}

	if errors.Is(err, catalog.ErrInvalidQuery) {
		return errs.New(http.StatusBadRequest, err.Error())
	}

	return err
}
