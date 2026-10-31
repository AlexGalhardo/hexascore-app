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
	"github.com/AlexGalhardo/hexascore-app/app/recommend/trending"
)

// trendingTTL is a backstop: the hourly ComputeTrending job forgets the entries after each snapshot.
const trendingTTL = time.Hour

// TrendingEntry is one rank of the home trending list.
type TrendingEntry struct {
	Rank  int    `json:"rank"`
	Basis string `json:"basis"`
	catalog.Highlight
}

// TrendingList is the home payload of one type; Items is empty before the first snapshot.
type TrendingList struct {
	ComputedAt *string         `json:"computedAt"`
	Items      []TrendingEntry `json:"items"`
}

func homeTrending(ctx context.Context, db catalog.Querier, typ, locale string) (TrendingList, error) {
	out := TrendingList{Items: []TrendingEntry{}}

	at, rows, err := trending.Latest(ctx, db, typ)
	if err != nil {
		return out, err
	}

	ids := make([]string, len(rows))
	for i, r := range rows {
		ids[i] = r.ItemID
	}

	items, err := catalog.HighlightsByID(ctx, db, ids, locale)
	if err != nil {
		return out, err
	}

	byID := make(map[string]catalog.Highlight, len(items))
	for _, it := range items {
		byID[it.ID] = it
	}

	for _, r := range rows {
		if it, ok := byID[r.ItemID]; ok {
			out.Items = append(out.Items, TrendingEntry{Rank: r.Rank, Basis: r.Basis, Highlight: it})
		}
	}

	if !at.IsZero() {
		s := at.Format(time.RFC3339)
		out.ComputedAt = &s
	}

	return out, nil
}

// recentTTL bounds how stale the "recently rated" strip may be.
const recentTTL = 60 * time.Second

var homeTypes = []string{"game", "movie", "series", "book"}

// Home renders Home/Index with one highlights section per item type (SDD 09
// §6.2). Props are eager: this is public SEO content (open issue 08-8).
func Home(app *vukem.App) web.HandlerFunc {
	return func(c *web.Ctx) error {
		sections := make([]catalog.SectionHighlights, 0, len(homeTypes))
		recent := []catalog.RecentRating{}
		trendingProp := map[string]TrendingList{}

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

		if db.Dialect().Name() == "postgres" {
			locale := i18n.Locale(c.Context())

			store, err := app.Cache()
			if err != nil {
				return err
			}

			for _, typ := range homeTypes {
				list, err := cache.Remember(c.Context(), store, trending.CacheKey(typ, locale), trendingTTL,
					func(ctx context.Context) (TrendingList, error) {
						return homeTrending(ctx, db.SQL(), typ, locale)
					})
				if err != nil {
					return err
				}

				trendingProp[typ] = list
			}
		}

		return inertia.Render(c, "Home/Index",
			inertia.Props{"sections": sections, "recent": recent, "trending": trendingProp})
	}
}
