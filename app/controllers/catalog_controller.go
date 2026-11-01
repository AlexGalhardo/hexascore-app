package controllers

import (
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

// CatalogPageSize is the default item count of one Catalog/Index page; ?per= picks 25, 50, 75 or 100
// (owner round 2026-11-01 item 20).
const CatalogPageSize = 25

// catalogPageSize reads ?per=; any value but the four offered sizes is the default.
func catalogPageSize(raw string) int {
	switch n, _ := strconv.Atoi(raw); n {
	case 25, 50, 75, 100:
		return n
	default:
		return CatalogPageSize
	}
}

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

// Index renders Catalog/Index for the item type typ: one page of items plus the opaque signed cursors of the
// pages before and after it ("" at either end), which the page turns into buttons and rel=prev/next links.
// Facets are in the first paint.
func (h *Catalog) Index(typ string) web.HandlerFunc {
	return func(c *web.Ctx) error {
		db, err := h.app.DB()
		if err != nil {
			return err
		}

		params, err := catalog.ParseListQuery(c.Query, typ)
		params.Limit = catalogPageSize(c.Query("per"))
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

		// Facets ride the first response. A follow-up reload used to widen the
		// selects after paint and shift the list (CLS on /games).
		genres, platforms, err := catalog.Facets(c.Context(), db, typ, nil)
		if err != nil {
			return err
		}

		props := inertia.Props{
			"type":       typ,
			"items":      map[string]any{"data": cards},
			"pagination": map[string]string{"prev": page.PrevCursor, "next": page.NextCursor},
			"filters": map[string]any{
				"genre": params.Genre, "platform": params.Platform, "year": params.Year,
				"minScore": params.MinScore, "sort": params.Sort, "per": params.Limit,
			},
			"facets": map[string]any{"genres": genres, "platforms": platforms},
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
