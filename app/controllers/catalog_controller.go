package controllers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

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

		params, err := listParams(c, typ)
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
				genres, platforms, err := catalog.Facets(ctx, db, typ)

				return map[string]any{"genres": genres, "platforms": platforms}, err
			}),
		}
		if page.Total != nil {
			props["total"] = *page.Total
		}

		return inertia.Render(c, "Catalog/Index", props)
	}
}

func listParams(c *web.Ctx, typ string) (catalog.ListParams, error) {
	p := catalog.ListParams{
		Type: typ, Genre: c.Query("genre"), Platform: c.Query("platform"),
		Sort: c.Query("sort"), Cursor: c.Query("cursor"), Limit: CatalogPageSize,
	}

	if v := c.Query("year"); v != "" {
		year, err := strconv.Atoi(v)
		if err != nil {
			return p, catalog.ErrInvalidQuery
		}
		p.Year = year
	}
	stars := func(name string) (*float64, error) {
		v := c.Query(name)
		if v == "" {
			return nil, nil
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, catalog.ErrInvalidQuery
		}

		return &f, nil
	}

	var err error
	if p.MinScore, err = stars("min_score"); err != nil {
		return p, err
	}
	if p.MaxScore, err = stars("max_score"); err != nil {
		return p, err
	}
	for _, key := range catalog.CategoryKeys[typ] {
		v, err := stars("min_" + key)
		if err != nil {
			return p, err
		}
		if v != nil {
			if p.MinCategory == nil {
				p.MinCategory = map[string]float64{}
			}
			p.MinCategory[key] = *v
		}
	}

	return p, nil
}

// badListing maps bad filters and forged cursors to 400.
func badListing(err error) error {
	if errors.Is(err, catalog.ErrInvalidQuery) || errors.Is(err, catalog.ErrInvalidCursor) {
		return errs.New(http.StatusBadRequest, err.Error())
	}

	return err
}
