package publicapi

import (
	"net/http"
	"time"

	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/web"

	hxapi "github.com/AlexGalhardo/hexascore-app/app/api"
	"github.com/AlexGalhardo/hexascore-app/app/catalog"
	"github.com/AlexGalhardo/hexascore-app/app/recommend/trending"
)

// Activity limits (SDD 08 §6): at most 90 days, within the last 365; the default is the last 30.
const (
	activityMaxDays     = 90
	activityHorizonDays = 365
	activityDefaultDays = 30
)

// TrendingEntryJSON is one rank of listTrending. rank, trendScore and basis are HexaScore-owned (A-07).
type TrendingEntryJSON struct {
	Rank       int      `json:"rank"       validate:"required"`
	Basis      string   `json:"basis"      validate:"required,oneof=trend prior"`
	TrendScore float64  `json:"trendScore"`
	Item       ItemJSON `json:"item"`
}

// TrendingJSON is listTrending: the top 10 of a type over the last 30 days. computedAt is null before the first
// snapshot.
type TrendingJSON struct {
	ComputedAt *string             `json:"computedAt"`
	Data       []TrendingEntryJSON `json:"data"`
}

// TrendingInput is the listTrending query.
type TrendingInput struct {
	Locale string `json:"locale" doc:"en, pt-BR or es"`
}

// ListTrending is listTrending for one type: GET /{games,movies,series,books}/trending. Literal sections, because
// "/{type}/trending" overlaps "/entities/{id}" on the ServeMux (as for genres).
func (h *Controller) ListTrending(section string) func(*web.Ctx, TrendingInput) (TrendingJSON, error) {
	return func(c *web.Ctx, in TrendingInput) (TrendingJSON, error) { return h.trending(c, typeOf[section], in) }
}

func (h *Controller) trending(c *web.Ctx, typ string, in TrendingInput) (TrendingJSON, error) {
	if err := checkQuery(c, "locale"); err != nil {
		return TrendingJSON{}, err
	}

	locale, err := localeOf(in.Locale)
	if err != nil {
		return TrendingJSON{}, err
	}

	at, rows, err := trending.Latest(c.Context(), h.DB, typ)
	if err != nil {
		return TrendingJSON{}, err
	}

	ids := make([]string, len(rows))
	for i, r := range rows {
		ids[i] = r.ItemID
	}

	items, err := catalog.LoadAPIItems(c.Context(), h.DB, ids, locale, h.Policy, false)
	if err != nil {
		return TrendingJSON{}, err
	}

	byID := make(map[string]catalog.APIItem, len(items))
	for _, it := range items {
		byID[it.Detail.ID] = it
	}

	out := TrendingJSON{Data: []TrendingEntryJSON{}}

	for _, r := range rows {
		if it, ok := byID[r.ItemID]; ok {
			out.Data = append(out.Data,
				TrendingEntryJSON{Rank: r.Rank, Basis: r.Basis, TrendScore: r.Score, Item: itemJSON(it, locale)})
		}
	}

	if !at.IsZero() {
		s := at.Format(time.RFC3339)
		out.ComputedAt = &s
	}

	etag, err := web.ETagOf(out)
	if err != nil {
		return TrendingJSON{}, err
	}

	c.SetHeader("Cache-Control", "public, max-age=300")

	if c.NotModified(etag, at) {
		return TrendingJSON{}, errNotModified
	}

	return out, nil
}

// ActivityDayJSON is the raw, unweighted activity of one UTC day.
type ActivityDayJSON struct {
	Day              string `json:"day"              validate:"required"`
	Views            int    `json:"views"`
	SearchClicks     int    `json:"searchClicks"`
	RatingsNew       int    `json:"ratingsNew"`
	RatingsUpdated   int    `json:"ratingsUpdated"`
	StatusWant       int    `json:"statusWant"`
	StatusInProgress int    `json:"statusInProgress"`
	StatusCompleted  int    `json:"statusCompleted"`
	StatusDropped    int    `json:"statusDropped"`
	ListAdds         int    `json:"listAdds"`
	Suggestions      int    `json:"suggestions"`
}

// ActivityJSON is getItemActivity.
type ActivityJSON struct {
	Data []ActivityDayJSON `json:"data"`
}

// ActivityInput is the getItemActivity path and query.
type ActivityInput struct {
	Type string `json:"type" validate:"oneof=games movies series books" doc:"games, movies, series or books"`
	ID   string `json:"id"`
	From string `json:"from"                                            doc:"first UTC day, YYYY-MM-DD; default 30 days before to"`
	To   string `json:"to"                                              doc:"last UTC day, YYYY-MM-DD; default today"`
}

// GetItemActivity is getItemActivity: GET /api/v1/{type}/{id}/activity, business plan only (SDD 05 FR-05-17). It
// is one request on the quota, like any read.
func (h *Controller) GetItemActivity(c *web.Ctx, in ActivityInput) (ActivityJSON, error) {
	typ, ok := typeOf[in.Type]
	if !ok {
		return ActivityJSON{}, invalid(unknownType)
	}

	if err := checkQuery(c, "from", "to"); err != nil {
		return ActivityJSON{}, err
	}

	plan, err := h.plan(c)
	if err != nil {
		return ActivityJSON{}, err
	}

	if plan.ID != hxapi.PlanBusiness {
		return ActivityJSON{}, errs.New(http.StatusForbidden, "plan required").WithExtension("code", "plan-required")
	}

	from, to, err := h.activityRange(in)
	if err != nil {
		return ActivityJSON{}, err
	}

	if !uuidRe.MatchString(in.ID) {
		return ActivityJSON{}, errs.ErrNotFound
	}

	var public bool
	if err := h.DB.QueryRowContext(c.Context(),
		`SELECT true FROM items WHERE id = $1 AND type = $2 AND visibility = 'public'`, in.ID, typ).
		Scan(&public); err != nil {
		return ActivityJSON{}, errs.ErrNotFound
	}

	days, err := trending.Activity(c.Context(), h.DB, in.ID, from, to)
	if err != nil {
		return ActivityJSON{}, err
	}

	out := ActivityJSON{Data: make([]ActivityDayJSON, len(days))}
	for i, d := range days {
		out.Data[i] = ActivityDayJSON(d)
	}

	return out, nil
}

func (h *Controller) activityRange(in ActivityInput) (from, to time.Time, err error) {
	now := time.Now()
	if h.Now != nil {
		now = h.Now()
	}

	today := now.UTC().Truncate(24 * time.Hour)
	to = today

	if in.To != "" {
		if to, err = time.Parse(time.DateOnly, in.To); err != nil {
			return from, to, errs.New(http.StatusUnprocessableEntity, "to must be a YYYY-MM-DD date")
		}
	}

	from = to.AddDate(0, 0, -(activityDefaultDays - 1))

	if in.From != "" {
		if from, err = time.Parse(time.DateOnly, in.From); err != nil {
			return from, to, errs.New(http.StatusUnprocessableEntity, "from must be a YYYY-MM-DD date")
		}
	}

	switch {
	case to.Before(from):
		err = errs.New(http.StatusUnprocessableEntity, "from must not be after to")
	case to.Sub(from) >= activityMaxDays*24*time.Hour:
		err = errs.New(http.StatusUnprocessableEntity, "the range is at most 90 days")
	case from.Before(today.AddDate(0, 0, -activityHorizonDays)) || to.After(today):
		err = errs.New(http.StatusUnprocessableEntity, "the range must lie within the last 365 days")
	}

	return from, to, err
}
