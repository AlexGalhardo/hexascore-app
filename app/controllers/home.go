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
	"github.com/AlexGalhardo/hexascore-app/app/recommend/trending"
	"github.com/AlexGalhardo/hexascore-app/app/social"
)

// recentTTL bounds how stale the "recently rated" strip may be.
const recentTTL = 60 * time.Second

// Home trending strip (SDD 08 section 7.2 "Home window"): 3 per type over the last 7 days, held 5 minutes.
const (
	trendingTTL  = 5 * time.Minute
	trendingDays = 7
	trendingTop  = 3
)

var homeTypes = []string{"game", "movie", "series", "book"}

// Home renders Home/Index with one highlights section per item type (SDD 09
// §6.2). Props are eager: this is public SEO content (open issue 08-8).
func Home(app *vukem.App) web.HandlerFunc {
	var (
		once sync.Once
		acc  *accounts.Accounts
		svc  *foryou.Service
		sc   *social.Follows
	)

	return func(c *web.Ctx) error {
		sections := make([]catalog.SectionHighlights, 0, len(homeTypes))
		recent := []catalog.RecentRating{}
		trend := map[string]trendingSection{}

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

			trend, err = cache.Remember(c.Context(), store, "home:trending7:v1:"+locale, trendingTTL,
				func(ctx context.Context) (map[string]trendingSection, error) {
					return homeTrending(ctx, db.SQL(), app.Now(), locale)
				})
			if err != nil {
				return err
			}
		}

		props := inertia.Props{"sections": sections, "recent": recent, "trending": trend}

		if db.Dialect().Name() == "postgres" {
			once.Do(func() {
				if a, err := accounts.New(app); err == nil {
					acc, svc = a, &foryou.Service{DB: db, Now: app.Now}
				}

				if ff, err := app.Flags(); err == nil {
					sc = &social.Follows{DB: db, Flags: ff}
				}
			})

			if acc != nil {
				if u, err := acc.Web.User(c); err == nil {
					props["forYou"] = inertia.Defer(func(ctx context.Context) (any, error) {
						return homeForYou(ctx, svc, u.ID, i18n.Locale(ctx))
					}).Group("personal")

					if sc.Enabled(c.Context()) {
						props["friends"] = inertia.Defer(func(ctx context.Context) (any, error) {
							return sc.FriendsRated(ctx, u.ID, i18n.Locale(ctx))
						}).Group("personal")
					}
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

type trendingEntry struct {
	Rank  int    `json:"rank"`
	Basis string `json:"basis"`
	catalog.Highlight
}

type trendingSection struct {
	ComputedAt string          `json:"computedAt"`
	Items      []trendingEntry `json:"items"`
}

// homeTrending ranks each type over the last trendingDays and hydrates the entries; empty types are kept empty so the
// client hides them.
func homeTrending(
	ctx context.Context,
	db catalog.Querier,
	now time.Time,
	locale string,
) (map[string]trendingSection, error) {
	out := make(map[string]trendingSection, len(homeTypes))

	for _, typ := range homeTypes {
		sec := trendingSection{ComputedAt: now.UTC().Format(time.RFC3339), Items: []trendingEntry{}}

		entries, err := trending.Window(ctx, db, typ, now, trendingDays, trendingTop)
		if err != nil {
			return nil, err
		}

		ids := make([]string, len(entries))
		for i, e := range entries {
			ids[i] = e.ID
		}

		cards, err := catalog.HighlightsByID(ctx, db, ids, locale)
		if err != nil {
			return nil, err
		}

		byID := make(map[string]catalog.Highlight, len(cards))
		for _, h := range cards {
			byID[h.ID] = h
		}

		for _, e := range entries {
			if h, ok := byID[e.ID]; ok {
				sec.Items = append(sec.Items, trendingEntry{Rank: len(sec.Items) + 1, Basis: e.Basis, Highlight: h})
			}
		}

		out[typ] = sec
	}

	return out, nil
}
