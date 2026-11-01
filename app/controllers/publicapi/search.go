package publicapi

import (
	"errors"
	"net/http"
	"slices"
	"strconv"

	"github.com/AlexGalhardo/vukem/api"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

// lookupSources are the ids lookupItem accepts as input (legacy parity). Restricted ones are matched but
// never echoed: the item serializer filters external ids by source (FR-05-16).
var lookupSources = []string{"imdb", "steam", "igdb", "tmdb", "isbn13", "openlibrary", "wikidata"}

// SearchItemsInput is the searchItems query; values are strings so bad input is a 400 invalid-query.
type SearchItemsInput struct {
	Q      string `json:"q"      doc:"1 to 100 characters"`
	Type   string `json:"type"   doc:"games, movies, series or books"`
	Locale string `json:"locale" doc:"en, pt-BR or es"`
	Limit  string `json:"limit"  doc:"1 to 50, default 20"`
	Page   string `json:"page"   doc:"1 to 10, default 1"`
}

// SearchEntitiesInput is the searchEntities query.
type SearchEntitiesInput struct {
	Q      string `json:"q"      doc:"1 to 100 characters"`
	Kind   string `json:"kind"   doc:"person, organization or franchise"`
	Locale string `json:"locale" doc:"en, pt-BR or es"`
	Limit  string `json:"limit"  doc:"1 to 50, default 20"`
	Page   string `json:"page"   doc:"1 to 10, default 1"`
}

// EntitySummaryJSON is one searchEntities hit.
type EntitySummaryJSON struct {
	ID        string `json:"id"        validate:"required"`
	Kind      string `json:"kind"      validate:"required"`
	Slug      string `json:"slug"      validate:"required"`
	Name      string `json:"name"`
	ItemCount int    `json:"itemCount"`
}

// EntityPageJSON is one page of entity hits in the vukem envelope.
type EntityPageJSON struct {
	Data []EntitySummaryJSON `json:"data"`
	Meta api.Meta            `json:"meta"`
}

// LookupInput is the lookupItem query.
type LookupInput struct {
	Source string `json:"source" doc:"imdb, steam, igdb, tmdb, isbn13, openlibrary or wikidata"`
	ID     string `json:"id"     doc:"the id in that source, 1 to 100 characters"`
	Locale string `json:"locale" doc:"en, pt-BR or es"`
}

// SearchItems is searchItems: GET /api/v1/search. It ranks allowlisted-source text only (open issue 07-7).
func (h *Controller) SearchItems(c *web.Ctx, in SearchItemsInput) (ItemPageJSON, error) {
	if err := checkQuery(c, "q", "type", "locale", "limit", "page"); err != nil {
		return ItemPageJSON{}, err
	}
	if in.Q == "" {
		return ItemPageJSON{}, invalid("q is required")
	}
	locale, err := localeOf(in.Locale)
	if err != nil {
		return ItemPageJSON{}, err
	}
	page, perPage, err := h.paging(c, in.Page, in.Limit)
	if err != nil {
		return ItemPageJSON{}, err
	}
	q := search.Query{Text: in.Q, Locale: locale, Page: page, PerPage: perPage, Audience: search.AudienceAPI}
	if in.Type != "" {
		typ, ok := typeOf[in.Type]
		if !ok {
			return ItemPageJSON{}, invalid("unknown type")
		}
		q.Types = []string{typ}
	}

	res, err := h.Search.Search(c.Context(), q)
	if err != nil {
		return ItemPageJSON{}, searchErr(err)
	}
	ids := make([]string, len(res.Hits))
	for i, hit := range res.Hits {
		ids[i] = hit.ItemID
	}
	items, err := catalog.LoadAPIItems(c.Context(), h.DB, ids, locale, h.Policy, false)
	if err != nil {
		return ItemPageJSON{}, err
	}

	total := int64(res.Total)
	out := ItemPageJSON{Data: make([]ItemJSON, len(items)), Meta: api.Meta{Page: page, PerPage: perPage, Total: &total}}
	for i, it := range items {
		out.Data[i] = itemJSON(it, locale)
	}
	out.Links.Self = c.Request().URL.RequestURI()

	return out, notModified(c, out)
}

// SearchEntities is searchEntities: GET /api/v1/search/entities.
func (h *Controller) SearchEntities(c *web.Ctx, in SearchEntitiesInput) (EntityPageJSON, error) {
	if err := checkQuery(c, "q", "kind", "locale", "limit", "page"); err != nil {
		return EntityPageJSON{}, err
	}
	if in.Q == "" {
		return EntityPageJSON{}, invalid("q is required")
	}
	locale, err := localeOf(in.Locale)
	if err != nil {
		return EntityPageJSON{}, err
	}
	page, perPage, err := h.paging(c, in.Page, in.Limit)
	if err != nil {
		return EntityPageJSON{}, err
	}
	q := search.EntityQuery{Text: in.Q, Locale: locale, Page: page, PerPage: perPage, Audience: search.AudienceAPI}
	if in.Kind != "" {
		q.Kinds = []string{in.Kind}
	}

	res, err := h.Search.SearchEntities(c.Context(), q)
	if err != nil {
		return EntityPageJSON{}, searchErr(err)
	}

	total := int64(res.Total)
	out := EntityPageJSON{
		Data: make([]EntitySummaryJSON, len(res.Hits)),
		Meta: api.Meta{Page: page, PerPage: perPage, Total: &total},
	}
	for i, e := range res.Hits {
		out.Data[i] = EntitySummaryJSON{ID: e.ID, Kind: e.Kind, Slug: e.Slug, Name: e.Name, ItemCount: e.ItemCount}
	}

	return out, notModified(c, out)
}

// LookupItem is lookupItem: GET /api/v1/lookup. The source id is input only; the body is the getItem
// serializer, which never echoes a restricted source id (FR-05-16).
func (h *Controller) LookupItem(c *web.Ctx, in LookupInput) (ItemJSON, error) {
	if err := checkQuery(c, "source", "id", "locale"); err != nil {
		return ItemJSON{}, err
	}
	if !slices.Contains(lookupSources, in.Source) {
		return ItemJSON{}, invalid("unknown source")
	}
	if in.ID == "" || len(in.ID) > 100 {
		return ItemJSON{}, invalid("id must be 1 to 100 characters")
	}
	locale, err := localeOf(in.Locale)
	if err != nil {
		return ItemJSON{}, err
	}

	id, ok, err := catalog.LookupItemID(c.Context(), h.DB, in.Source, in.ID)
	if err != nil {
		return ItemJSON{}, err
	}
	if !ok {
		return ItemJSON{}, errs.ErrNotFound
	}
	items, err := catalog.LoadAPIItems(c.Context(), h.DB, []string{id}, locale, h.Policy, true)
	if err != nil {
		return ItemJSON{}, err
	}
	if len(items) == 0 {
		return ItemJSON{}, errs.ErrNotFound
	}

	out := itemJSON(items[0], locale)

	return out, notModified(c, out)
}

// paging parses page (default 1) and limit (default 20, at most 50 and the plan page size); the searcher
// enforces the 200 reachable results.
func (h *Controller) paging(c *web.Ctx, rawPage, rawLimit string) (page, perPage int, err error) {
	page, perPage = 1, 20
	if rawPage != "" {
		if page, err = strconv.Atoi(rawPage); err != nil {
			return 0, 0, invalid("page must be a positive integer")
		}
	}
	if rawLimit != "" {
		if perPage, err = strconv.Atoi(rawLimit); err != nil {
			return 0, 0, invalid("limit must be a positive integer")
		}
		plan, err := h.plan(c)
		if err != nil {
			return 0, 0, err
		}
		if perPage > plan.MaxPageSize {
			return 0, 0, invalid("limit exceeds the " + plan.ID + " plan page size")
		}
	}

	return page, perPage, nil
}

func searchErr(err error) error {
	switch {
	case errors.Is(err, search.ErrInvalidQuery):
		return errs.New(http.StatusBadRequest, err.Error())
	case errors.Is(err, search.ErrTimeout):
		return errs.New(http.StatusServiceUnavailable, "search timeout")
	}

	return err
}
