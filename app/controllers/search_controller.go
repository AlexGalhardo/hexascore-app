package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	vukem "github.com/AlexGalhardo/vukem"
	"github.com/AlexGalhardo/vukem/cache"
	"github.com/AlexGalhardo/vukem/errs"
	"github.com/AlexGalhardo/vukem/i18n"
	"github.com/AlexGalhardo/vukem/inertia"
	"github.com/AlexGalhardo/vukem/uuid"
	"github.com/AlexGalhardo/vukem/web"

	"github.com/AlexGalhardo/hexascore-app/app/models"
	"github.com/AlexGalhardo/hexascore-app/app/requests"
	"github.com/AlexGalhardo/hexascore-app/internal/search"
)

const (
	searchTTL     = 120 * time.Second
	suggestTTL    = 300 * time.Second
	searchPerPage = 20
	entityLimit   = 5
)

// Search serves /search, /search/suggest and POST /search/events (SDD 07 §6).
type Search struct {
	app      *vukem.App
	searcher search.Searcher
	// user returns the signed-in user, or an error for anonymous requests.
	user func(c *web.Ctx) (models.User, error)
}

// NewSearch builds the handlers.
func NewSearch(app *vukem.App, searcher search.Searcher, user func(c *web.Ctx) (models.User, error)) *Search {
	return &Search{app: app, searcher: searcher, user: user}
}

// SearchHit is one result row; the catalogue read model adds score and cover later.
type SearchHit struct {
	ID             string               `json:"id"`
	Type           string               `json:"type"`
	Slug           string               `json:"slug"`
	Title          string               `json:"title"`
	Path           string               `json:"path"`
	MatchedTitle   string               `json:"matchedTitle,omitempty"`
	MatchedEdition *search.EditionMatch `json:"matchedEdition,omitempty"`
}

// SearchGroup is one type's block of the grouped page.
type SearchGroup struct {
	Type     string      `json:"type"`
	Items    []SearchHit `json:"items"`
	Total    int         `json:"total"`
	MoreHref string      `json:"moreHref"`
}

// SearchEntity is one matching person, company or franchise.
type SearchEntity struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	ItemCount int    `json:"itemCount"`
}

type searchQueryProp struct {
	Q      string `json:"q"`
	Type   string `json:"type"`
	Locale string `json:"locale"`
	Page   int    `json:"page"`
}

// searchPage is the cached part of the page: it depends only on the query.
type searchPage struct {
	Groups     []SearchGroup  `json:"groups"`
	Results    []SearchHit    `json:"results"`
	Total      int            `json:"total"`
	Entities   []SearchEntity `json:"entities"`
	DidYouMean *string        `json:"didYouMean"`
	TooShort   bool           `json:"tooShort"`
}

var (
	itemSection   = map[string]string{"game": "games", "movie": "movies", "series": "series", "book": "books"}
	entitySection = map[string]string{"person": "people", "organization": "companies", "franchise": "franchises"}
)

// searchParams reads q, type (comma separated) and locale; a missing locale
// means the interface locale, an empty one means every locale.
func searchParams(c *web.Ctx) (text, locale string, types []string) {
	q := c.Request().URL.Query()
	locale = i18n.Locale(c.Context())

	if q.Has("locale") {
		locale = q.Get("locale")
	}

	for t := range strings.SplitSeq(q.Get("type"), ",") {
		if t = strings.TrimSpace(t); t != "" {
			types = append(types, t)
		}
	}

	return q.Get("q"), locale, types
}

func (h *Search) signedIn(c *web.Ctx) (models.User, bool) {
	u, err := h.user(c)

	return u, err == nil
}

// cacheOff skips the anonymous result cache so k6 measures the database path (SDD 07 NFR-07-1, NFR-07-2).
func cacheOff() bool { return os.Getenv("SEARCH_CACHE") == "off" }

func cacheKey(kind string, parts ...any) string {
	var b strings.Builder
	for _, p := range parts {
		switch v := p.(type) {
		case []string:
			b.WriteString(strings.Join(v, ","))
		case int:
			b.WriteString(strconv.Itoa(v))
		case string:
			b.WriteString(v)
		}

		b.WriteByte('|')
	}

	sum := sha256.Sum256([]byte(b.String()))

	return "search:v1:" + kind + ":" + strconv.FormatInt(search.Epoch(), 10) + ":" + hex.EncodeToString(sum[:])
}

func itemPath(prefix, typ, slug string) string { return prefix + "/" + itemSection[typ] + "/" + slug }

func toHit(prefix string, h search.Hit) SearchHit {
	return SearchHit{
		ID: h.ItemID, Type: h.Type, Slug: h.Slug, Title: h.Title, Path: itemPath(prefix, h.Type, h.Slug),
		MatchedTitle: h.MatchedTitle, MatchedEdition: h.MatchedEdition,
	}
}

// Index renders Search/Index. Anonymous requests share a cached result for 120 s;
// anonymous Inertia visits carry an ETag over the JSON page (09-2), never the HTML.
func (h *Search) Index(c *web.Ctx) error {
	text, locale, types := searchParams(c)
	page := max(1, c.QueryInt("page", 1))
	prefix := i18n.PathPrefix(c.Context())

	user, signed := h.signedIn(c)
	query := search.Query{
		Text:     text,
		Locale:   locale,
		Types:    types,
		Page:     page,
		PerPage:  searchPerPage,
		Audience: search.AudienceSite,
	}
	start := time.Now()

	compute := func(ctx context.Context) (searchPage, error) { return h.run(ctx, query, prefix) }

	var (
		out searchPage
		err error
	)

	if signed || cacheOff() {
		out, err = compute(c.Context())
	} else {
		store, cerr := h.app.Cache()
		if cerr != nil {
			return cerr
		}

		out, err = cache.Remember(
			c.Context(),
			store,
			cacheKey("page", text, locale, prefix, types, page),
			searchTTL,
			compute,
		)
	}

	if err != nil {
		return badQuery(err)
	}

	searchID := uuid.NewV4().String()

	props := inertia.Props{
		"query":      searchQueryProp{Q: text, Type: strings.Join(types, ","), Locale: locale, Page: page},
		"groups":     out.Groups,
		"results":    out.Results,
		"total":      out.Total,
		"entities":   out.Entities,
		"didYouMean": out.DidYouMean,
		"searchId":   searchID,
		"tooShort":   out.TooShort,
	}

	if signed {
		c.SetHeader("Cache-Control", "private, no-store")
	} else if inertia.IsInertia(c.Request()) {
		// searchId differs per request and must not defeat the 304.
		etag, err := web.ETagOf(map[string]any{"ui": i18n.Locale(c.Context()), "query": props["query"], "page": out})
		if err != nil {
			return err
		}

		c.SetHeader("Cache-Control", "private, no-cache")

		if c.NotModified(etag, time.Time{}) {
			return nil
		}
	}

	if text != "" && !out.TooShort {
		h.recordPerformed(c, searchID, query, user, signed, out, time.Since(start), "web")
	}

	return inertia.Render(c, "Search/Index", props)
}

// AppIndex is GET /app/v1/search. Auth is optional (SDD 10). The audience is
// site, and the stored surface is app so web analytics stay separate.
func (h *Search) AppIndex(c *web.Ctx) error {
	text, locale, types := searchParams(c)
	page := max(1, c.QueryInt("page", 1))
	query := search.Query{
		Text: text, Locale: locale, Types: types, Page: page, PerPage: searchPerPage,
		Audience: search.AudienceSite,
	}
	user, signed := h.signedIn(c)
	start := time.Now()

	out, err := h.run(c.Context(), query, "")
	if err != nil {
		return badQuery(err)
	}

	searchID := uuid.NewV4().String()
	if text != "" && !out.TooShort {
		h.recordPerformed(c, searchID, query, user, signed, out, time.Since(start), "app")
	}

	hits := out.Results
	if len(hits) == 0 {
		for _, g := range out.Groups {
			hits = append(hits, g.Items...)
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"searchId": searchID, "data": hits, "total": out.Total, "tooShort": out.TooShort,
	})
}

// badQuery maps the searcher's validation failures to 400 (SDD 07 §6).
func badQuery(err error) error {
	if errors.Is(err, search.ErrInvalidQuery) {
		return errs.New(http.StatusBadRequest, err.Error())
	}

	return err
}

// run executes the searches behind one page.
func (h *Search) run(ctx context.Context, q search.Query, prefix string) (searchPage, error) {
	res, err := h.searcher.Search(ctx, q)
	if err != nil {
		return searchPage{}, err
	}

	out := searchPage{
		Total: res.Total, TooShort: res.TooShort, DidYouMean: res.DidYouMean,
		Groups: []SearchGroup{}, Results: []SearchHit{}, Entities: []SearchEntity{},
	}

	for _, hit := range res.Hits {
		out.Results = append(out.Results, toHit(prefix, hit))
	}

	for _, g := range res.Groups {
		sg := SearchGroup{
			Type: g.Type, Total: g.Total, Items: []SearchHit{},
			MoreHref: prefix + "/search?" + url.Values{"q": {q.Text}, "type": {g.Type}}.Encode(),
		}
		for _, hit := range g.Hits {
			sg.Items = append(sg.Items, toHit(prefix, hit))
		}

		out.Groups = append(out.Groups, sg)
	}

	if q.Page == 1 && len(q.Types) == 0 && !res.TooShort {
		ents, err := h.searcher.SearchEntities(ctx, search.EntityQuery{
			Text: q.Text, Locale: q.Locale, PerPage: entityLimit, Audience: search.AudienceSite,
		})
		if err != nil {
			return searchPage{}, err
		}

		for _, e := range ents.Hits {
			out.Entities = append(out.Entities, SearchEntity{
				ID: e.ID, Kind: e.Kind, Name: e.Name, ItemCount: e.ItemCount,
				Path: prefix + "/" + entitySection[e.Kind] + "/" + e.Slug,
			})
		}
	}

	return out, nil
}

func (h *Search) recordPerformed(
	c *web.Ctx,
	id string,
	q search.Query,
	u models.User,
	signed bool,
	out searchPage,
	took time.Duration,
	surface string,
) {
	db, err := h.app.DB()
	if err != nil {
		return
	}

	e := search.Performed{
		SearchID: id, Surface: surface, Text: q.Text, Locale: q.Locale, Types: q.Types,
		ResultCount: out.Total, Latency: took, Page: q.Page,
	}
	for _, hit := range out.Results {
		e.ResultRefs = append(e.ResultRefs, hit.ID)
	}
	for _, g := range out.Groups {
		for _, hit := range g.Items {
			e.ResultRefs = append(e.ResultRefs, hit.ID)
		}
	}
	if signed {
		e.UserID = &u.ID
	}

	// Analytics must never fail a search.
	if err := search.RecordPerformed(c.Context(), db, e); err != nil {
		c.Log().WarnContext(c.Context(), "search event", "err", err)
	}
}

// SuggestItem is one autocomplete entry (legacy shape plus kind, SDD 07 §6).
type SuggestItem struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	Kind          string `json:"kind"`
	Title         string `json:"title"`
	Year          *int   `json:"year"`
	CoverThumbURL string `json:"coverThumbUrl,omitempty"`
	Path          string `json:"path"`
}

// Suggest answers /search/suggest with up to 8 entries; anonymous answers are
// cached for 300 s and carry an ETag.
func (h *Search) Suggest(c *web.Ctx) error {
	text, locale, types := searchParams(c)
	prefix := i18n.PathPrefix(c.Context())

	compute := func(ctx context.Context) ([]SuggestItem, error) {
		rows, err := h.searcher.Suggest(ctx, search.SuggestQuery{Text: text, Locale: locale, Kinds: types})
		if err != nil {
			return nil, err
		}

		items := make([]SuggestItem, 0, len(rows))
		for _, r := range rows {
			it := SuggestItem{ID: r.RefID, Type: r.Kind, Kind: r.Kind, Title: r.Title}
			if r.Year > 0 {
				it.Year = &r.Year
			}

			if r.ThumbKey != "" {
				it.CoverThumbURL = "/covers/" + strings.TrimPrefix(r.ThumbKey, "covers/")
			}

			if section, ok := itemSection[r.Kind]; ok {
				it.Path = prefix + "/" + section + "/" + r.Slug
			} else {
				it.Path = prefix + "/" + entitySection[r.Kind] + "/" + r.Slug
			}

			items = append(items, it)
		}

		return items, nil
	}

	var (
		items []SuggestItem
		err   error
	)

	if _, signed := h.signedIn(c); signed || cacheOff() {
		c.SetHeader("Cache-Control", "private, no-store")
		items, err = compute(c.Context())
	} else {
		store, cerr := h.app.Cache()
		if cerr != nil {
			return cerr
		}

		items, err = cache.Remember(
			c.Context(),
			store,
			cacheKey("suggest", text, locale, types, prefix),
			suggestTTL,
			compute,
		)
		c.SetHeader("Cache-Control", "public, max-age=60")
		c.SetHeader("Vary", "Cookie")
	}

	if err != nil {
		return badQuery(err)
	}

	body := map[string]any{"data": items}

	if _, signed := h.signedIn(c); !signed {
		etag, err := web.ETagOf(body)
		if err != nil {
			return err
		}

		if c.NotModified(etag, time.Time{}) {
			return nil
		}
	}

	return c.JSON(http.StatusOK, body)
}

// Event records a result click (POST /search/events, 204) and dispatches
// search.ResultClicked once it is stored. An unknown search id is ignored.
func (h *Search) Event(c *web.Ctx) error {
	in, err := web.Parse(c, requests.SearchClickSchema)
	if err != nil {
		return err
	}

	db, err := h.app.DB()
	if err != nil {
		return err
	}

	ev, ok, err := search.RecordClick(c.Context(), db,
		search.Click{SearchID: in.SearchID, RefID: in.RefID, Position: in.Position, Group: in.Group}, h.app.Now())
	if err != nil {
		return err
	}

	if ok && h.app.Events != nil {
		if err := h.app.Events.Dispatch(c.Context(), ev); err != nil {
			c.Log().WarnContext(c.Context(), "search click event", "err", err)
		}
	}

	return c.NoContent()
}

// LegacySuggest redirects the legacy /api/search/suggest to /search/suggest
// with 308, keeping the query string.
func LegacySuggest(c *web.Ctx) error {
	target := "/search/suggest"
	if raw := c.Request().URL.RawQuery; raw != "" {
		target += "?" + raw
	}

	http.Redirect(c.Writer(), c.Request(), target, http.StatusPermanentRedirect)

	return nil
}
