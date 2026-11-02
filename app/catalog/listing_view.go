package catalog

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Summaries returns the listing cards for ids, in the same order. Titles fall
// back from locale to en, then to the original title; the score follows the
// 5-rating read gate.
func Summaries(ctx context.Context, db Querier, ids []string, locale string) ([]Highlight, error) {
	if len(ids) == 0 {
		return []Highlight{}, nil
	}

	rows, err := db.QueryContext(ctx, `SELECT i.id, i.type, i.slug,
  coalesce(t.title, te.title, i.original_title, i.slug),
  extract(year FROM i.release_date)::int,
  `+starsScore+`,
  i.rating_count, i.cover_key, i.cover_thumb_key
FROM unnest($1::uuid[]) WITH ORDINALITY AS o(id, n)
JOIN items i ON i.id = o.id
LEFT JOIN item_translations t ON t.item_id = i.id AND t.locale = $2
LEFT JOIN item_translations te ON te.item_id = i.id AND te.locale = 'en'
ORDER BY o.n`, "{"+strings.Join(ids, ",")+"}", locale)
	if err != nil {
		return nil, fmt.Errorf("listing summaries: %w", err)
	}

	return scanHighlights(rows)
}

// Facet is one selectable genre or platform with its public item count.
type Facet struct {
	Slug  string `json:"slug"`
	Name  string `json:"name"`
	Count int    `json:"count"`
	// Canonical marks a HexaScore genre (owner 2026-11-02): its translated name is ours, so the API may show it.
	Canonical bool `json:"-"`
}

// Facets lists the genres and platforms used by public items of typ, most used first, named in locale; non-nil
// sources keeps only relations from those sources (API).
func Facets(ctx context.Context, db Querier, typ, locale string, sources []string) (genres, platforms []Facet, err error) {
	if genres, err = FacetsOf(ctx, db, typ, "genre", locale, sources); err != nil {
		return nil, nil, err
	}
	platforms, err = FacetsOf(ctx, db, typ, "platform", locale, sources)

	return genres, platforms, err
}

// FacetsOf lists the entities of kind used by public items of typ, most used first (Facets per kind; also the
// game modes select of /games, owner 2026-11-02).
func FacetsOf(ctx context.Context, db Querier, typ, kind, locale string, sources []string) ([]Facet, error) {
	var srcArray any // NULL = every source (site)
	if sources != nil {
		srcArray = "{" + strings.Join(sources, ",") + "}"
	}
	rows, err := db.QueryContext(ctx, `SELECT e.slug, coalesce(t.name, e.default_name), count(*),
  e.provenance->>'canonical' IS NOT DISTINCT FROM e.kind
FROM entities e
JOIN item_entities ie ON ie.entity_id = e.id AND ie.role = e.kind AND ($3::text[] IS NULL OR ie.source = ANY($3))
JOIN items i ON i.id = ie.item_id AND i.type = $1 AND i.visibility = 'public'
LEFT JOIN entity_translations t ON t.entity_id = e.id AND t.locale = $4
WHERE e.kind = $2
GROUP BY e.slug, e.default_name, e.provenance, e.kind, t.name
ORDER BY count(*) DESC, 2
LIMIT 200`, typ, kind, srcArray, locale)
	if err != nil {
		return nil, fmt.Errorf("facets %s: %w", kind, err)
	}
	defer rows.Close()

	out := []Facet{}
	for rows.Next() {
		var f Facet
		if err := rows.Scan(&f.Slug, &f.Name, &f.Count, &f.Canonical); err != nil {
			return nil, fmt.Errorf("scan facet: %w", err)
		}
		out = append(out, f)
	}

	return out, rows.Err()
}

// ActiveFilter is one entity filter in force on a listing, shown as a removable chip.
type ActiveFilter struct {
	Param string `json:"param"`
	Value string `json:"value"`
	Name  string `json:"name"`
}

// ActiveEntityFilters names the entity filters in force (kind -> slug or code); an unknown key keeps its value as
// the name, so the chip still says what filters the list.
func ActiveEntityFilters(ctx context.Context, db Querier, filters map[string]string, locale string) ([]ActiveFilter, error) {
	out := []ActiveFilter{}
	for _, kind := range slices.Sorted(maps.Keys(filters)) {
		if kind == "game_mode" {
			continue // the game modes select shows it
		}
		v := filters[kind]
		key := "slug = $2"
		if kind == "country" || kind == "language" {
			key, v = "code = upper($2)", strings.ToUpper(v)
		}
		name := v
		err := db.QueryRowContext(ctx, `SELECT coalesce(t.name, e.default_name) FROM entities e
LEFT JOIN entity_translations t ON t.entity_id = e.id AND t.locale = $3
WHERE e.kind = $1 AND e.`+key+` LIMIT 1`, kind, v, locale).Scan(&name)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("active filter %s: %w", kind, err)
		}
		out = append(out, ActiveFilter{Param: ListingFilters[kind], Value: v, Name: name})
	}

	return out, nil
}
