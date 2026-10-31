package controllers

import (
	"context"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/cache"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/catalog"
)

// recentTTL bounds how stale the "recently rated" strip may be.
const recentTTL = 60 * time.Second

var homeTypes = []string{"game", "movie", "series", "book"}

// Home renders Home/Index with one highlights section per item type (SDD 09
// §6.2). Props are eager: this is public SEO content (open issue 08-8).
func Home(app *vukem.App) web.HandlerFunc {
	return func(c *web.Ctx) error {
		sections := make([]catalog.SectionHighlights, 0, len(homeTypes))
		recent := []catalog.RecentRating{}

		db, err := app.DB()
		if err != nil {
			return err
		}

		for _, typ := range homeTypes {
			s := catalog.SectionHighlights{
				Type:      typ,
				TopRated:  []catalog.Highlight{},
				Newest:    []catalog.Highlight{},
				MostRated: []catalog.Highlight{},
			}
			// catalog tables exist only on Postgres (VF-033).
			if db.Dialect().Name() == "postgres" {
				if s, err = catalog.Highlights(c.Context(), db, typ, i18n.Locale(c.Context())); err != nil {
					return err
				}
			}
			sections = append(sections, s)
		}

		if db.Dialect().Name() == "postgres" {
			locale := i18n.Locale(c.Context())

			store, err := app.Cache()
			if err != nil {
				return err
			}

			// Shared across anonymous and signed-in visitors: the strip holds only public data.
			recent, err = cache.Remember(c.Context(), store, "home:recent:v1:"+locale, recentTTL,
				func(ctx context.Context) ([]catalog.RecentRating, error) {
					return catalog.RecentRatings(ctx, db.SQL(), locale)
				})
			if err != nil {
				return err
			}
		}

		return inertia.Render(c, "Home/Index", inertia.Props{"sections": sections, "recent": recent})
	}
}
