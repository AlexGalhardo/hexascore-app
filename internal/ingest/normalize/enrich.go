package normalize

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/AlexGalhardo/hexascore-app/internal/ingest/store"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikidata"
	"github.com/AlexGalhardo/hexascore-app/internal/ingest/wikipedia"
)

// Synopsis is one item_translations synopsis with its source and licence.
type Synopsis struct {
	Text, Source, Licence, URL string
}

// extraPayload is the shared shape of the TMDB and IGDB records. Their
// overview text is deliberately not decoded: it is licensed input for the
// rewriting step and never reaches item_translations (FR-06-3).
type extraPayload struct {
	ID          int      `json:"id"`
	URL         string   `json:"url"`
	Cover       string   `json:"cover"`
	Genres      []string `json:"genres"`
	Platforms   []string `json:"platforms"`
	Creators    []string `json:"creators"`
	Countries   []string `json:"countries"`
	Trailers    []string `json:"trailers"`
	Runtime     int      `json:"runtime"`
	Seasons     int      `json:"seasons"`
	Episodes    int      `json:"episodes"`
	ReleaseDate string   `json:"release_date"`
}

// sources holds the decoded non-Wikidata records of one QID.
type sources struct {
	extra       extraPayload
	extraSource string
	steam       struct {
		Platforms   []string `json:"platforms"`
		HeaderImage string   `json:"header_image"`
	}
	gog struct {
		Platforms []string `json:"platforms"`
		Cover     string   `json:"cover"`
	}
	openlibrary struct {
		Pages  int   `json:"pages"`
		Covers []int `json:"covers"`
	}
	youtube *struct {
		Available []string `json:"available"`
	}
	own map[string]string
}

func decodeSources(qid string, recs []store.Record) (*sources, error) {
	by := map[string]json.RawMessage{}

	for _, r := range recs {
		if r.ExternalID == qid {
			by[r.Source] = r.Payload
		}
	}

	s := &sources{}

	// TMDB only has movies and series and IGDB only games, so at most one exists.
	for _, src := range []string{SourceTMDB, SourceIGDB} {
		if raw, ok := by[src]; ok {
			s.extraSource = src
			if err := json.Unmarshal(raw, &s.extra); err != nil {
				return nil, fmt.Errorf("decode %s payload of %s: %w", src, qid, err)
			}

			break
		}
	}

	targets := map[string]any{
		SourceSteam:       &s.steam,
		SourceGOG:         &s.gog,
		SourceOpenLibrary: &s.openlibrary,
		"synopsis":        &s.own,
	}
	for src, dst := range targets {
		if raw, ok := by[src]; ok {
			if err := json.Unmarshal(raw, dst); err != nil {
				return nil, fmt.Errorf("decode %s payload of %s: %w", src, qid, err)
			}
		}
	}

	if raw, ok := by["youtube"]; ok {
		if err := json.Unmarshal(raw, &s.youtube); err != nil {
			return nil, fmt.Errorf("decode youtube payload of %s: %w", qid, err)
		}
	}

	return s, nil
}

type candidate struct {
	values []string
	source string
}

// firstList is the Wikidata list when it has values, else the first non-empty
// fallback: Wikidata wins, other sources only fill (FR-06-3).
func firstList(p Provenance, field string, wd []string, fallbacks ...candidate) []string {
	if len(wd) > 0 {
		p[field] = SourceWikidata

		return wd
	}

	for _, c := range fallbacks {
		if len(c.values) > 0 {
			p[field] = c.source

			return c.values
		}
	}

	return nil
}

// enrich layers the other sources over the Wikidata mapping and records the
// provenance of every field. It never touches a field Wikidata filled.
func enrich(item *Item, wd wikidata.Payload, summaries map[string]wikipedia.Payload, s *sources) {
	p := item.Provenance
	x, xs := s.extra, s.extraSource

	item.Meta.Genres = firstList(p, "meta.genres", item.Meta.Genres, candidate{x.Genres, xs})
	item.Meta.Creators = firstList(p, "meta.creators", item.Meta.Creators, candidate{x.Creators, xs})
	item.Meta.Countries = firstList(p, "meta.countries", item.Meta.Countries, candidate{x.Countries, xs})

	if wd.Type == "game" {
		// IGDB lists every platform; Steam and GOG only know the PC operating systems.
		item.Meta.Platforms = firstList(
			p,
			"meta.platforms",
			item.Meta.Platforms,
			candidate{
				x.Platforms,
				xs,
			},
			candidate{s.steam.Platforms, SourceSteam},
			candidate{s.gog.Platforms, SourceGOG},
		)
	}

	enrichCounts(item, wd, s)
	enrichTrailers(item, s)

	if item.ReleaseDate != nil {
		p["release_date"] = SourceWikidata
	} else if d, err := time.Parse("2006-01-02", x.ReleaseDate); err == nil {
		item.ReleaseDate, item.ReleasePrecision = &d, PrecisionDay
		p["release_date"] = xs
	}

	for k := range item.ExternalIDs {
		p["external_ids."+k] = SourceWikidata
	}

	if _, ok := item.ExternalIDs["tmdb"]; !ok && xs == SourceTMDB && x.ID > 0 {
		item.ExternalIDs["tmdb"] = strconv.Itoa(x.ID)
		p["external_ids.tmdb"] = SourceTMDB
	}

	for loc := range item.Titles {
		p["translations."+loc+".title"] = titleSource(wd, loc)
	}

	item.Synopses = synopses(item, summaries, s.own)
	for loc, syn := range item.Synopses {
		p["synopsis."+loc] = syn.Source
	}

	if item.Cover = pickCover(wd.Type, s); item.Cover != nil {
		p["cover"] = item.Cover.Source
	}
}

func enrichCounts(item *Item, wd wikidata.Payload, s *sources) {
	p, m, xs := item.Provenance, &item.Meta, s.extraSource

	fill := func(field string, dst *int, fallback int, fallbackSource string) {
		switch {
		case *dst > 0:
			p[field] = SourceWikidata
		case fallback > 0:
			*dst, p[field] = fallback, fallbackSource
		}
	}

	if wd.Type == "movie" || wd.Type == "series" {
		fill("runtime_minutes", &m.RuntimeMinutes, s.extra.Runtime, xs)
	}

	if wd.Type == "series" {
		fill("season_count", &m.Seasons, s.extra.Seasons, xs)
		fill("episode_count", &m.Episodes, s.extra.Episodes, xs)
	}

	if wd.Type == "book" {
		fill("page_count", &m.Pages, s.openlibrary.Pages, SourceOpenLibrary)
	}
}

// enrichTrailers lists Wikidata candidates first, then the source ones. Once
// the YouTube check ran, only the ids it found embeddable stay.
func enrichTrailers(item *Item, s *sources) {
	source := s.extraSource
	if len(item.Meta.Trailers) > 0 {
		source = SourceWikidata
	}

	for _, id := range s.extra.Trailers {
		if !slices.Contains(item.Meta.Trailers, id) {
			item.Meta.Trailers = append(item.Meta.Trailers, id)
		}
	}

	if s.youtube != nil {
		item.Meta.Trailers = slices.DeleteFunc(item.Meta.Trailers, func(id string) bool {
			return !slices.Contains(s.youtube.Available, id)
		})
	}

	if len(item.Meta.Trailers) > MaxTrailers {
		item.Meta.Trailers = item.Meta.Trailers[:MaxTrailers]
	}

	if len(item.Meta.Trailers) > 0 {
		item.Provenance["meta.trailers"] = source
	}
}

// titleSource is wikipedia only when the title is the summary fallback.
func titleSource(wd wikidata.Payload, locale string) string {
	for _, lang := range labelLangs[locale] {
		if wd.Labels[lang] != "" {
			return SourceWikidata
		}
	}

	if locale == "en" && wd.Sitelinks["enwiki"] != "" {
		return SourceWikidata
	}

	return SourceWikipedia
}

// synopses takes the Wikipedia extract of each locale that has a title, else
// the own rewritten text for a locale without a Wikipedia summary; no other
// source supplies synopsis text.
func synopses(item *Item, summaries map[string]wikipedia.Payload, own map[string]string) map[string]Synopsis {
	out := map[string]Synopsis{}

	for _, l := range locales {
		if item.Titles[l.locale] == "" {
			continue
		}

		if sum, ok := summaries[l.wiki]; ok {
			if sum.Extract != "" {
				out[l.locale] = Synopsis{
					Text:    sum.Extract,
					Source:  SourceWikipedia,
					Licence: Licence(SourceWikipedia),
					URL:     sum.URL,
				}
			}

			continue
		}

		if own[l.locale] != "" {
			out[l.locale] = Synopsis{Text: own[l.locale], Source: SourceOwn, Licence: Licence(SourceOwn)}
		}
	}

	return out
}
