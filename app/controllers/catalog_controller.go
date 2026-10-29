package controllers

import (
	"context"
	"errors"
	"net/http"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// CatalogPageSize is the item count of one Catalog/Index scroll page (SDD 09 §7.3).
const CatalogPageSize = 24

// Catalog serves the section listings Catalog/Index.
type Catalog struct {
	app   *vukem.App
	codec *catalog.CursorCodec
}

// NewCatalog builds the handlers; the cursor key derives from APP_KEY.
func NewCatalog(app *vukem.App) (*Catalog, error) {
	key, err := app.DeriveKey("catalog.cursor", 32)
	if err != nil {
		return nil, err
	}
	codec, err := catalog.NewCursorCodec(key)
	if err != nil {
		return nil, err
	}

	return &Catalog{app: app, codec: codec}, nil
}

// Index renders Catalog/Index for the item type typ. `items` is a Scroll prop
// whose next page is an opaque signed cursor; `facets` loads on demand.
func (h *Catalog) Index(typ string) web.HandlerFunc {
	return func(c *web.Ctx) error {
		db, err := h.app.DB()
		if err != nil {
			return err
		}

		params, err := catalog.ParseListQuery(c.Query, typ)
		params.Limit = CatalogPageSize
		if err != nil {
			return badListing(err)
		}
		page, err := catalog.List(c.Context(), db, h.codec, params)
		if err != nil {
			return badListing(err)
		}

		ids := make([]string, len(page.Items))
		for i, it := range page.Items {
			ids[i] = it.ID
		}
		cards, err := catalog.Summaries(c.Context(), db, ids, i18n.Locale(c.Context()))
		if err != nil {
			return err
		}

		meta := inertia.ScrollMeta{PageName: "cursor"}
		if page.NextCursor != "" {
			meta.NextPage = page.NextCursor
		}

		props := inertia.Props{
			"type":  typ,
			"items": inertia.Scroll(map[string]any{"data": cards}, meta),
			"filters": map[string]any{
				"genre": params.Genre, "platform": params.Platform, "year": params.Year,
				"minScore": params.MinScore, "sort": params.Sort,
			},
			"facets": inertia.Optional(func(ctx context.Context) (any, error) {
				genres, platforms, err := catalog.Facets(ctx, db, typ, nil)

				return map[string]any{"genres": genres, "platforms": platforms}, err
			}),
		}
		if page.Total != nil {
			props["total"] = *page.Total
		}

		return inertia.Render(c, "Catalog/Index", props)
	}
}

// badListing maps bad filters and forged cursors to 400.
func badListing(err error) error {
	if errors.Is(err, catalog.ErrInvalidQuery) || errors.Is(err, catalog.ErrInvalidCursor) {
		return errs.New(http.StatusBadRequest, err.Error())
	}

	return err
}
