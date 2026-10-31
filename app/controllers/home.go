package controllers

import (
	"context"
	"sync"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/cache"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/accounts"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/foryou"
)

// recentTTL bounds how stale the "recently rated" strip may be.
const recentTTL = 60 * time.Second

var homeTypes = []string{"game", "movie", "series", "book"}

// Home renders Home/Index with one highlights section per item type (SDD 09
// §6.2). Props are eager: this is public SEO content (open issue 08-8).
func Home(app *vukem.App) web.HandlerFunc {
	var (
		once sync.Once
		acc  *accounts.Accounts
		svc  *foryou.Service
	)

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

		props := inertia.Props{"sections": sections, "recent": recent}

		if db.Dialect().Name() == "postgres" {
			once.Do(func() {
				if a, err := accounts.New(app); err == nil {
					acc, svc = a, &foryou.Service{DB: db, Now: app.Now}
				}
			})

			if acc != nil {
				if u, err := acc.Web.User(c); err == nil {
					props["forYou"] = inertia.Defer(func(ctx context.Context) (any, error) {
						return homeForYou(ctx, svc, u.ID, i18n.Locale(ctx))
					}).Group("personal")
				}
			}
		}

		return inertia.Render(c, "Home/Index", props)
	}
}

// homeForYou is the home strip: the personal lists of the types that have one. Onboarding types are left out, so the
// strip is empty (and hidden) for opted-out users and users with too few positive items.
func homeForYou(ctx context.Context, svc *foryou.Service, userID int64, locale string) (map[string]foryou.List, error) {
	out := map[string]foryou.List{}

	for _, typ := range homeTypes {
		l, err := svc.Get(ctx, userID, typ, locale)
		if err != nil {
			return nil, err
		}

		if l.Basis == foryou.BasisPersonal {
			out[typ] = l
		}
	}

	return out, nil
}
