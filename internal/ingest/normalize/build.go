package normalize

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikidata"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikipedia"
)

// Release statuses stored in items.release_status. Cancelled is set by
// editors, never by the normaliser.
const (
	StatusReleased  = "released"
	StatusAnnounced = "announced"
)

// MaxTrailers is the trailer cap kept in meta.
const MaxTrailers = 2

// Skip reasons: the item is not built, which is not a failure.
var (
	// ErrNoWikidata means the QID has no Wikidata record.
	ErrNoWikidata = errors.New("normalize: no wikidata record")
	// ErrNoEnglishTitle means no English label, sitelink or summary title.
	ErrNoEnglishTitle = errors.New("normalize: no english title")
)

// externalIDProps maps an external id key to its Wikidata property.
var externalIDProps = map[string]string{
	"imdb": "P345", "igdb": "P5794", "steam": "P1733", "isbn13": "P212", "openlibrary": "P648",
	"hltb": "P2816", "goodreads": "P8383", "metacritic": "P1712", "rottentomatoes": "P1258",
	"epic": "P6278", "gog": "P2725",
}

var (
	tmdbProp    = map[string]string{"movie": "P4947", "series": "P4983"}
	creatorProp = map[string]string{"movie": "P57", "series": "P170", "game": "P178", "book": "P50"}
	// minutesPerUnit converts runtime units: minute, hour, second, "1".
	minutesPerUnit = map[string]float64{"Q7727": 1, "Q25235": 60, "Q11574": 1.0 / 60, "1": 1}
	nonISBN        = regexp.MustCompile(`[^0-9X]`)
)

// locales lists, per catalogue locale, the Wikidata label languages in order
// of preference and the Wikipedia edition used as a title fallback.
var locales = []struct{ locale, wiki string }{{"en", "en"}, {"pt-BR", "pt"}, {"es", "es"}}

var labelLangs = map[string][]string{"en": {"en", "mul"}, "pt-BR": {"pt-br", "pt", "mul"}, "es": {"es", "mul"}}

// Meta holds the Wikidata-derived item metadata. Empty fields are omitted
// when stored.
type Meta struct {
	Genres         []string `json:"genres,omitempty"`
	Platforms      []string `json:"platforms,omitempty"`
	Creators       []string `json:"creators,omitempty"`
	Countries      []string `json:"countries,omitempty"`
	Trailers       []string `json:"trailers,omitempty"`
	RuntimeMinutes int      `json:"runtimeMinutes,omitempty"`
	Pages          int      `json:"pages,omitempty"`
	Seasons        int      `json:"seasons,omitempty"`
	Episodes       int      `json:"episodes,omitempty"`
}

// Item is the core mapping of one QID. Enrichment from other sources and
// synopses are layered on by the merge (P6.A.3.2).
type Item struct {
	QID              string
	Type             string
	ReleaseDate      *time.Time
	ReleasePrecision string
	ReleaseStatus    string
	ExternalIDs      map[string]string
	Meta             Meta
	// Titles is keyed by locale (en, pt-BR, es); a locale without a title is absent.
	Titles     map[string]string
	Popularity float64
}

// BuildItem maps the stored records of one QID to an Item. It is a pure
// function of the set of records, so their order does not matter (FR-06-2);
// now only decides announced versus released. It returns ErrNoWikidata or
// ErrNoEnglishTitle when the item must be skipped.
func BuildItem(qid string, recs []store.Record, now time.Time) (*Item, error) {
	wd, summaries, err := decode(qid, recs)
	if err != nil {
		return nil, err
	}

	titles := buildTitles(wd, summaries)
	if titles["en"] == "" {
		return nil, ErrNoEnglishTitle
	}

	item := &Item{
		QID:              qid,
		Type:             wd.Type,
		ReleasePrecision: PrecisionUnknown,
		ReleaseStatus:    StatusReleased,
		ExternalIDs:      externalIDs(wd),
		Meta:             buildMeta(wd),
		Titles:           titles,
		Popularity:       float64(wd.SitelinksCount),
	}

	if d, ok := releaseDate(wd); ok {
		item.ReleaseDate, item.ReleasePrecision = &d.Time, d.Precision

		today := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)
		if d.Time.After(today) {
			item.ReleaseStatus = StatusAnnounced
		}
	}

	return item, nil
}

func buildTitles(wd wikidata.Payload, summaries map[string]wikipedia.Payload) map[string]string {
	titles := map[string]string{}

	for _, l := range locales {
		title := ""

		for _, lang := range labelLangs[l.locale] {
			if wd.Labels[lang] != "" {
				title = wd.Labels[lang]

				break
			}
		}

		if l.locale == "en" && title == "" {
			title = wd.Sitelinks["enwiki"]
		}

		if title == "" {
			title = summaries[l.wiki].Title
		}

		if title != "" {
			titles[l.locale] = title
		}
	}

	return titles
}

// ranked puts preferred statements first, keeping the stored order otherwise.
func ranked(claims []wikidata.Claim) []wikidata.Claim {
	return slices.SortedStableFunc(slices.Values(claims), func(a, b wikidata.Claim) int {
		return rankOrder(a.Rank) - rankOrder(b.Rank)
	})
}

func rankOrder(rank string) int {
	if rank == "preferred" {
		return 0
	}

	return 1
}

func firstValue(wd wikidata.Payload, prop string) any {
	if c := ranked(wd.Claims[prop]); len(c) > 0 {
		return c[0].Value
	}

	return nil
}

func firstString(wd wikidata.Payload, prop string) string {
	s, _ := firstValue(wd, prop).(string)

	return s
}

// quantity reads a quantity claim in minutes-style units; 0 means absent,
// not positive or an unknown unit.
func quantity(wd wikidata.Payload, prop string, units map[string]float64) int {
	v, ok := firstValue(wd, prop).(map[string]any)
	if !ok {
		return 0
	}

	amount, err := strconv.ParseFloat(fmt.Sprint(v["amount"]), 64)
	factor, known := units[fmt.Sprint(v["unit"])]

	if err != nil || !known || amount <= 0 {
		return 0
	}

	return int(math.Round(amount * factor))
}

var countUnits = map[string]float64{"1": 1}

func refLabels(wd wikidata.Payload, prop string) []string {
	var out []string

	for _, c := range ranked(wd.Claims[prop]) {
		id, _ := c.Value.(string)
		if label := wd.Refs[id]; label != "" && !slices.Contains(out, label) {
			out = append(out, label)
		}
	}

	return out
}

func releaseDate(wd wikidata.Payload) (Date, bool) {
	props := []string{"P577", "P580", "P571"}
	if wd.Type == "series" {
		props = []string{"P580", "P577", "P571"}
	}

	for _, prop := range props {
		var best Date

		found := false

		for _, c := range wd.Claims[prop] {
			d, ok := ParseTime(c.Value)
			if !ok {
				continue
			}

			if !found || d.Time.Before(best.Time) || (d.Time.Equal(best.Time) && finer(d, best)) {
				best, found = d, true
			}
		}

		if found {
			return best, true
		}
	}

	return Date{}, false
}

func externalIDs(wd wikidata.Payload) map[string]string {
	props := make(map[string]string, len(externalIDProps)+1)
	maps.Copy(props, externalIDProps)

	if p, ok := tmdbProp[wd.Type]; ok {
		props["tmdb"] = p
	}

	ids := map[string]string{}

	for key, prop := range props {
		if s := firstString(wd, prop); s != "" {
			ids[key] = s
		}
	}

	if v, ok := ids["isbn13"]; ok {
		ids["isbn13"] = nonISBN.ReplaceAllString(strings.ToUpper(v), "")
	}

	return ids
}

func buildMeta(wd wikidata.Payload) Meta {
	m := Meta{
		Genres:    refLabels(wd, "P136"),
		Creators:  refLabels(wd, creatorProp[wd.Type]),
		Countries: refLabels(wd, "P495"),
	}

	switch wd.Type {
	case "game":
		m.Platforms = refLabels(wd, "P400")
	case "book":
		m.Pages = quantity(wd, "P1104", countUnits)
	case "movie", "series":
		m.RuntimeMinutes = quantity(wd, "P2047", minutesPerUnit)
	}

	if wd.Type == "series" {
		m.Seasons = quantity(wd, "P2437", countUnits)
		m.Episodes = quantity(wd, "P1113", countUnits)
	}

	for _, c := range ranked(wd.Claims["P1651"]) {
		if id, _ := c.Value.(string); id != "" && !slices.Contains(m.Trailers, id) {
			m.Trailers = append(m.Trailers, id)
		}
	}

	if len(m.Trailers) > MaxTrailers {
		m.Trailers = m.Trailers[:MaxTrailers]
	}

	return m
}

// decode picks the wikidata payload and the wikipedia summaries of qid out of
// its records, in any order.
func decode(qid string, recs []store.Record) (wikidata.Payload, map[string]wikipedia.Payload, error) {
	bySource := map[string]json.RawMessage{}

	for _, r := range recs {
		if r.ExternalID == qid {
			bySource[r.Source] = r.Payload
		}
	}

	raw, ok := bySource["wikidata"]
	if !ok {
		return wikidata.Payload{}, nil, ErrNoWikidata
	}

	var wd wikidata.Payload
	if err := json.Unmarshal(raw, &wd); err != nil {
		return wikidata.Payload{}, nil, fmt.Errorf("decode wikidata payload of %s: %w", qid, err)
	}

	summaries := map[string]wikipedia.Payload{}

	for _, l := range locales {
		if r, ok := bySource["wikipedia:"+l.wiki]; ok {
			var p wikipedia.Payload
			if err := json.Unmarshal(r, &p); err != nil {
				return wikidata.Payload{}, nil, fmt.Errorf("decode wikipedia:%s payload of %s: %w", l.wiki, qid, err)
			}

			summaries[l.wiki] = p
		}
	}

	return wd, summaries, nil
}
