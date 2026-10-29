package publicapi

import (
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// EntityJSON is one entity on /api/v1: licence-filtered fields, provenance and meta (SDD 01 §6.2).
type EntityJSON struct {
	ID          string                         `json:"id"                    validate:"required"`
	Kind        string                         `json:"kind"                  validate:"required"`
	Slug        string                         `json:"slug"                  validate:"required"`
	WikidataQID string                         `json:"wikidataQid"`
	Code        string                         `json:"code,omitempty"`
	Locale      string                         `json:"locale"                validate:"required"`
	Name        string                         `json:"name,omitempty"`
	Description *catalog.Synopsis              `json:"description,omitempty"`
	BornOn      string                         `json:"bornOn,omitempty"`
	DiedOn      string                         `json:"diedOn,omitempty"`
	FoundedYear *int                           `json:"foundedYear,omitempty"`
	ExternalIDs map[string]string              `json:"externalIds,omitempty"`
	Provenance  map[string]catalog.FieldSource `json:"provenance"`
	Meta        ItemMeta                       `json:"meta"`
}

// GetEntityInput is the entity path and query shared by getEntity, getPerson, getOrganization, getFranchise
// and getPlatform.
type GetEntityInput struct {
	ID     string `json:"id"`
	Locale string `json:"locale" doc:"en, pt-BR or es"`
}

// ListEntityItemsInput is the listEntityItems path and query.
type ListEntityItemsInput struct {
	ID     string `json:"id"`
	Type   string `json:"type"   doc:"games, movies, series or books"`
	Role   string `json:"role"   doc:"relation role, e.g. developer, director or genre"`
	Sort   string `json:"sort"   doc:"[-]popularity, score, release_date or rating_count"`
	Limit  string `json:"limit"  doc:"1 to the plan page size, default 10"`
	Cursor string `json:"cursor" doc:"opaque next_cursor of the previous page"`
	Locale string `json:"locale" doc:"en, pt-BR or es"`
}

// GetEntity is getEntity: GET /api/v1/entities/{id}.
func (h *Controller) GetEntity(c *web.Ctx, in GetEntityInput) (EntityJSON, error) {
	return h.entity(c, in, "")
}

// GetPerson is getPerson: GET /api/v1/people/{id}.
func (h *Controller) GetPerson(c *web.Ctx, in GetEntityInput) (EntityJSON, error) {
	return h.entity(c, in, "person")
}

// GetOrganization is getOrganization: GET /api/v1/organizations/{id}.
func (h *Controller) GetOrganization(c *web.Ctx, in GetEntityInput) (EntityJSON, error) {
	return h.entity(c, in, "organization")
}

// GetFranchise is getFranchise: GET /api/v1/franchises/{id}.
func (h *Controller) GetFranchise(c *web.Ctx, in GetEntityInput) (EntityJSON, error) {
	return h.entity(c, in, "franchise")
}

// GetPlatform is getPlatform: GET /api/v1/platforms/{id}.
func (h *Controller) GetPlatform(c *web.Ctx, in GetEntityInput) (EntityJSON, error) {
	return h.entity(c, in, "platform")
}

// entity answers 404 for unknown ids and, when kind is set, for entities of another kind.
func (h *Controller) entity(c *web.Ctx, in GetEntityInput, kind string) (EntityJSON, error) {
	if err := checkQuery(c, "locale"); err != nil {
		return EntityJSON{}, err
	}
	locale, err := localeOf(in.Locale)
	if err != nil {
		return EntityJSON{}, err
	}
	if !uuidRe.MatchString(in.ID) {
		return EntityJSON{}, errs.ErrNotFound
	}

	e, ok, err := catalog.LoadAPIEntity(c.Context(), h.DB, in.ID, kind, locale, h.Policy)
	if err != nil {
		return EntityJSON{}, err
	}
	if !ok {
		return EntityJSON{}, errs.ErrNotFound
	}

	out := EntityJSON{
		ID: e.ID, Kind: e.Kind, Slug: e.Slug, WikidataQID: e.WikidataQID, Code: e.Code, Locale: locale,
		Name: e.Name, Description: e.Description, BornOn: e.BornOn, DiedOn: e.DiedOn, FoundedYear: e.FoundedYear,
		ExternalIDs: e.ExternalIDs, Provenance: e.Provenance, Meta: ItemMeta{OmittedFields: []string{}},
	}
	if e.Meta != nil {
		out.Meta.OmittedFields = e.Meta.OmittedFields
	}

	return out, notModified(c, out)
}

// ListEntityItems is listEntityItems: GET /api/v1/entities/{id}/items.
func (h *Controller) ListEntityItems(c *web.Ctx, in ListEntityItemsInput) (ItemPageJSON, error) {
	if err := checkQuery(c, "type", "role", "sort", "limit", "cursor", "locale"); err != nil {
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
	if !uuidRe.MatchString(in.ID) {
		return ItemPageJSON{}, errs.ErrNotFound
	}

	params := catalog.ListParams{EntityID: in.ID, Role: in.Role, Sort: in.Sort, Cursor: in.Cursor, Limit: limit}
	if in.Type != "" {
		typ, ok := typeOf[in.Type]
		if !ok {
			return ItemPageJSON{}, invalid("unknown type")
		}
		params.Type = typ
	}
	if in.Role != "" && !catalog.KnownRole(in.Role) {
		return ItemPageJSON{}, invalid("unknown role")
	}

	if _, ok, err := catalog.LoadAPIEntity(c.Context(), h.DB, in.ID, "", locale, h.Policy); err != nil {
		return ItemPageJSON{}, err
	} else if !ok {
		return ItemPageJSON{}, errs.ErrNotFound
	}

	return h.itemPage(c, params, locale, limit)
}
