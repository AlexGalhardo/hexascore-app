package controllers

import (
	"context"
	"fmt"
	"strings"
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
	"github.com/AlexGalhardo/hexascore-app/app/social"
)

// Home trending carousel (owner round 2026-11-01 item 26): the 12 most rated items per type over the last 30 days,
// filled with the most rated items overall when fewer were rated in the window, held 5 minutes.
const (
	trendingTTL  = 5 * time.Minute
	trendingDays = 30
	trendingTop  = 12
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
		trend := map[string]trendingSection{}

		db, err := app.DB()
		if err != nil {
			return err
		}

		for _, typ := range homeTypes {
			s := catalog.SectionHighlights{Type: typ, Pool: []catalog.Highlight{}}
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

			trend, err = cache.Remember(c.Context(), store, "home:trending30:v1:"+locale, trendingTTL,
				func(ctx context.Context) (map[string]trendingSection, error) {
					return homeTrending(ctx, db.SQL(), app.Now(), locale)
				})
			if err != nil {
				return err
			}
		}

		props := inertia.Props{"sections": sections, "trending": trend}

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
	Rank  int     `json:"rank"`
	Basis string  `json:"basis"`
	Genre *string `json:"genre"`
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

		entries, err := mostRatedWindow(ctx, db, typ, now, trendingDays, trendingTop)
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

		genres, err := primaryGenres(ctx, db, ids, locale)
		if err != nil {
			return nil, err
		}

		for _, e := range entries {
			if h, ok := byID[e.ID]; ok {
				var genre *string
				if g, ok := genres[e.ID]; ok {
					genre = &g
				}

				sec.Items = append(
					sec.Items,
					trendingEntry{Rank: len(sec.Items) + 1, Basis: e.Basis, Genre: genre, Highlight: h},
				)
			}
		}

		out[typ] = sec
	}

	return out, nil
}

// primaryGenres maps each of ids to the localized name of its first genre (lowest position); items without one are
// absent.
func primaryGenres(ctx context.Context, db catalog.Querier, ids []string, locale string) (map[string]string, error) {
	out := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}

	rows, err := db.QueryContext(ctx, `SELECT DISTINCT ON (ie.item_id) ie.item_id::text,
  coalesce(t.name, te.name, e.default_name)
FROM item_entities ie
JOIN entities e ON e.id = ie.entity_id
LEFT JOIN entity_translations t ON t.entity_id = e.id AND t.locale = $2
LEFT JOIN entity_translations te ON te.entity_id = e.id AND te.locale = 'en'
WHERE ie.item_id = ANY($1::uuid[]) AND ie.role = 'genre'
ORDER BY ie.item_id, ie.position, e.slug`, "{"+strings.Join(ids, ",")+"}", locale)
	if err != nil {
		return nil, fmt.Errorf("home primary genres: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("scan primary genre: %w", err)
		}

		out[id] = name
	}

	return out, rows.Err()
}

type windowEntry struct {
	ID    string
	Basis string
}

// mostRatedWindow ranks typ by new ratings over the last days UTC days ending at the day of now ("trend"), then
// fills up to n with the most rated public items overall ("prior").
func mostRatedWindow(
	ctx context.Context,
	db catalog.Querier,
	typ string,
	now time.Time,
	days, n int,
) ([]windowEntry, error) {
	day := now.UTC().Truncate(24 * time.Hour)
	out := []windowEntry{}
	seen := map[string]bool{}

	for _, q := range []struct {
		basis, sql string
		args       []any
	}{
		{"trend", `SELECT a.item_id::text
FROM item_activity_daily a
JOIN items i ON i.id = a.item_id AND i.visibility = 'public'
WHERE a.item_type = $1 AND a.day > $2::date - $3::int AND a.day <= $2::date
GROUP BY a.item_id, i.rating_count
HAVING sum(a.ratings_new_w) > 0
ORDER BY sum(a.ratings_new_w) DESC, i.rating_count DESC, a.item_id
LIMIT $4`, []any{typ, day, days, n}},
		{"prior", `SELECT id::text FROM items WHERE type = $1 AND visibility = 'public'
ORDER BY rating_count DESC, id LIMIT $2`, []any{typ, n}},
	} {
		if len(out) == n {
			break
		}

		ids, err := scanIDs(ctx, db, q.sql, q.args...)
		if err != nil {
			return nil, fmt.Errorf("home trending %s %s: %w", typ, q.basis, err)
		}

		for _, id := range ids {
			if len(out) < n && !seen[id] {
				seen[id] = true
				out = append(out, windowEntry{ID: id, Basis: q.basis})
			}
		}
	}

	return out, nil
}

func scanIDs(ctx context.Context, db catalog.Querier, query string, args ...any) ([]string, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}
